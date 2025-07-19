package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/not-for-prod/clay/server/log"
	"github.com/not-for-prod/clay/server/shutdown"
	"github.com/not-for-prod/clay/transport"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server is a transport server.
type Server struct {
	opts      *serverOpts
	listeners *listenerSet
	srv       *serverSet
}

// NewServer creates a Server listening on the rpcPort.
// Pass additional Options to mutate its behaviour.
// By default, HTTP JSON handler and gRPC are listening on the same
// port, admin port is p+2 and profile port is p+4.
func NewServer(rpcPort int, opts ...Option) *Server {
	serverOpts := defaultServerOpts(rpcPort)
	for _, opt := range opts {
		opt(serverOpts)
	}
	return &Server{opts: serverOpts}
}

// Run starts processing requests to the service.
// It blocks indefinitely, run asynchronously to do anything after that.
func (s *Server) Run(svc transport.Service) error {
	desc := svc.GetDescription()

	var err error
	s.listeners, err = newListenerSet(s.opts)
	if err != nil {
		return errors.Wrap(err, "couldn't create listeners")
	}

	s.srv = newServerSet(s.opts)

	// Inject static Swagger as root handler
	s.srv.router.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, bytes.NewReader(desc.SwaggerDef()))
	})
	s.srv.router.HandleFunc(
		"/docs/*", func(w http.ResponseWriter, r *http.Request) {
			httpSwagger.Handler(httpSwagger.URL("swagger.json")).ServeHTTP(w, r)
		},
	)
	s.srv.router.Get(
		"/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
		},
	)
	s.srv.router.Get(
		"/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger.json", http.StatusMovedPermanently)
		},
	)

	// apply gRPC interceptor
	if d, ok := desc.(transport.ConfigurableServiceDesc); ok {
		d.Apply(transport.WithUnaryInterceptor(s.opts.GRPCUnaryInterceptor))
	}

	// Register everything
	mux := runtime.NewServeMux(s.opts.RuntimeServeMuxOpts...)

	if err = desc.RegisterHTTP(context.Background(), mux); err != nil {
		return errors.Wrap(err, "couldn't register HTTP server")
	}

	s.srv.router.Mount("/", mux)

	desc.RegisterGRPC(s.srv.grpcSrv)

	return s.run()
}

func (s *Server) run() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Shutdown manager with automatic signal handling
	shutdownMgr := shutdown.NewManager(os.Interrupt, syscall.SIGTERM)

	// Register the shutdown function
	shutdownMgr.Register(func(ctx context.Context) error {
		log.Info("shutting down servers")

		if err := s.srv.httpSrv.Shutdown(ctx); err != nil {
			log.Error("HTTP shutdown error:", err)
		}

		s.srv.grpcSrv.GracefulStop()

		// Do not close mainListener manually — it causes a panic in cmux.Serve
		return nil
	})

	// Setting a timeout for shutdown (default is 30 seconds, can be changed)
	shutdownMgr.SetTimeout(10 * time.Second)

	// Start CMux
	if s.listeners.mainListener != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := s.listeners.mainListener.Serve(); err != nil {
				if !isExpectedNetErr(err) {
					log.Error("CMux serve error:", err)
					errChan <- fmt.Errorf("cmux serve error: %w", err)
				}
			}
		}()
	}

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.srv.httpSrv.Serve(s.listeners.HTTP); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("HTTP serve error:", err)
				errChan <- fmt.Errorf("http serve error: %w", err)
			}
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.srv.grpcSrv.Serve(s.listeners.GRPC); err != nil {
			if !isExpectedNetErr(err) {
				log.Error("gRPC serve error:", err)
				errChan <- fmt.Errorf("grpc serve error: %w", err)
			}
		}
	}()

	// Close error channel when all servers finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Wait for shutdown signal
	shutdownMgr.AwaitTermination()

	// Return first error if any occurred
	if err, ok := <-errChan; ok {
		return err
	}
	return nil
}

// isExpectedNetErr checks if the error is expected during graceful server shutdown.
// Returns true for network errors that occur when connections are closed during
// normal shutdown process (e.g., net.ErrClosed, "use of closed network connection",
// "mux: server closed"). These errors should not be treated as actual failures.
func isExpectedNetErr(err error) bool {
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	if strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "mux: server closed") {
		return true
	}
	return false
}
