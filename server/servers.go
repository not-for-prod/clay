package server

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/transport"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/pkg/errors"
)

type initFunc func() error

func (s *Server) initHTTPServer() error {
	router := chi.NewMux()

	// Use custom mux option
	if s.opts.HTTPMux != nil {
		router = s.opts.HTTPMux
	}

	// Apply http middlewares
	if len(s.opts.HTTPMiddlewares) > 0 {
		router.Use(s.opts.HTTPMiddlewares...)
	}

	// Inject static Swagger as root handler
	router.HandleFunc(
		"/swagger.json", func(w http.ResponseWriter, req *http.Request) {
			io.Copy(w, bytes.NewReader(s.serviceDesc.SwaggerDef()))
		},
	)
	router.HandleFunc(
		"/docs/*", func(w http.ResponseWriter, r *http.Request) {
			httpSwagger.Handler(httpSwagger.URL("swagger.json")).ServeHTTP(w, r)
		},
	)
	router.Get(
		"/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
		},
	)
	router.Get(
		"/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger.json", http.StatusMovedPermanently)
		},
	)

	// Register everything
	mux := runtime.NewServeMux(s.opts.RuntimeServeMuxOpts...)

	if err := s.serviceDesc.RegisterHTTP(context.Background(), mux); err != nil {
		return errors.Wrap(err, "couldn't register HTTP server")
	}

	router.Mount("/", mux)
	s.httpServer = &http.Server{
		Handler: router,
	}

	return nil
}

func (s *Server) initGRPCServer() error {
	grpcServer := grpc.NewServer(s.opts.GRPCOpts...)
	reflection.Register(grpcServer)

	// apply gRPC interceptor
	if d, ok := s.serviceDesc.(transport.ConfigurableServiceDesc); ok {
		d.Apply(transport.WithUnaryInterceptor(s.opts.GRPCUnaryInterceptor))
	}

	s.serviceDesc.RegisterGRPC(grpcServer)
	s.grpcServer = grpcServer

	return nil
}
