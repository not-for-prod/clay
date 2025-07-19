package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type serverSet struct {
	router  chi.Router
	grpcSrv *grpc.Server
	httpSrv *http.Server
}

func newServerSet(opts *serverOpts) *serverSet {
	router := chi.NewMux()

	if opts.HTTPMux != nil {
		router = opts.HTTPMux
	}

	if len(opts.HTTPMiddlewares) > 0 {
		router.Use(opts.HTTPMiddlewares...)
	}

	grpcServer := grpc.NewServer(opts.GRPCOpts...)
	if opts.EnableReflection {
		reflection.Register(grpcServer)
	}

	httpSrv := &http.Server{
		Handler: router,
	}

	srv := &serverSet{
		router:  router,
		grpcSrv: grpcServer,
		httpSrv: httpSrv,
	}
	return srv
}
