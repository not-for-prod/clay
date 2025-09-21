package server

import (
	"context"
	"net/http"

	"github.com/not-for-prod/clay/transport"
	"google.golang.org/grpc"
)

// Server is a transport server.
type Server struct {
	opts        *serverOpts
	listeners   *listenerSet
	serviceDesc transport.ServiceDesc
	httpServer  *http.Server
	grpcServer  *grpc.Server
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
func (s *Server) Run(descs ...transport.ServiceDesc) error {
	// Join several ServiceDescs in CompoundServiceDesc
	s.serviceDesc = transport.NewCompoundServiceDesc(descs...)

	// init Server
	for _, fn := range []initFunc{
		s.initListeners,
		s.initHTTPServer,
		s.initGRPCServer,
	} {
		if err := fn(); err != nil {
			return err
		}
	}

	// start Server
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

	if s.httpServer != nil {
		go func() {
			err := s.httpServer.Serve(s.listeners.HTTP)
			errChan <- err
		}()
	}

	if s.grpcServer != nil {
		go func() {
			err := s.grpcServer.Serve(s.listeners.GRPC)
			errChan <- err
		}()
	}

	return <-errChan
}

// Stop stops the server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}
