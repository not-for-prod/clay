package server

import (
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type serverSet struct {
	http chi.Router
	grpc *grpc.Server
}

func newServerSet(opts *serverOpts) *serverSet {
	http := chi.NewMux()

	if opts.HTTPMux != nil {
		http = opts.HTTPMux
	}

	if len(opts.HTTPMiddlewares) > 0 {
		http.Use(opts.HTTPMiddlewares...)
	}

	if len(opts.HTTPMiddlewares) > 0 {
		http.Use(opts.HTTPMiddlewares...)
	}

	grpcServer := grpc.NewServer(opts.GRPCOpts...)
	if opts.EnableReflection {
		reflection.Register(grpcServer)
	}

	srv := &serverSet{
		grpc: grpcServer,
		http: http,
	}
	return srv
}
