package server

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/transport"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pkg/errors"
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
func (s *Server) Run(ctx context.Context, svc transport.Service) error {
	desc := svc.GetDescription()

	var err error
	s.listeners, err = newListenerSet(s.opts)
	if err != nil {
		return errors.Wrap(err, "couldn't create listeners")
	}

	s.srv = newServerSet(s.listeners, s.opts)

	// Inject static Swagger as root handler
	s.srv.http.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, bytes.NewReader(desc.SwaggerDef()))
	})
	s.srv.http.HandleFunc(
		"/docs/*", func(w http.ResponseWriter, r *http.Request) {
			httpSwagger.Handler(httpSwagger.URL("swagger.json")).ServeHTTP(w, r)
		},
	)
	s.srv.http.Get(
		"/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
		},
	)
	s.srv.http.Get(
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

	if err = desc.RegisterHTTP(ctx, mux); err != nil {
		return errors.Wrap(err, "couldn't register HTTP server")
	}

	s.srv.http.Mount("/", mux)

	desc.RegisterGRPC(s.srv.grpc)

	return s.run()
}

func (s *Server) run() error {
	errChan := make(chan error, 5)

	if s.listeners.mainListener != nil {
		go func() {
			err := s.listeners.mainListener.Serve()
			errChan <- err
		}()
	}
	go func() {
		err := http.Serve(s.listeners.HTTP, s.srv.http)
		errChan <- err
	}()
	go func() {
		err := s.srv.grpc.Serve(s.listeners.GRPC)
		errChan <- err
	}()

	return <-errChan
}

// Stop stops the server gracefully.
func (s *Server) Stop() {
	// TODO grace HTTP
	s.srv.grpc.GracefulStop()
}
