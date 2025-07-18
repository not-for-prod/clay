package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/server/middlewares/mwhttp"
	"google.golang.org/grpc"
)

// Option is an optional setting applied to the Server.
type Option func(*serverOpts)

type serverOpts struct {
	RPCPort int
	// If HTTPPort is the same then muxing listener is created.
	HTTPPort int
	HTTPMux  *chi.Mux

	HTTPMiddlewares []func(http.Handler) http.Handler

	GRPCOpts             []grpc.ServerOption
	GRPCUnaryInterceptor grpc.UnaryServerInterceptor

	EnableReflection    bool
	RuntimeServeMuxOpts []runtime.ServeMuxOption
}

func defaultServerOpts(mainPort int) *serverOpts {
	return &serverOpts{
		RPCPort:  mainPort,
		HTTPPort: mainPort,
		HTTPMux:  chi.NewMux(),
	}
}

// WithGRPCOpts sets gRPC server options.
func WithGRPCOpts(opts ...grpc.ServerOption) Option {
	return func(o *serverOpts) {
		o.GRPCOpts = append(o.GRPCOpts, opts...)
	}
}

// WithHTTPPort sets HTTP RPC port to listen on.
// Set same port as main to use single port.
func WithHTTPPort(port int) Option {
	return func(o *serverOpts) {
		o.HTTPPort = port
	}
}

// WithHTTPMiddlewares sets up HTTP middlewares to work with.
func WithHTTPMiddlewares(mws ...mwhttp.Middleware) Option {
	mwGeneric := make([]func(http.Handler) http.Handler, 0, len(mws))
	for _, mw := range mws {
		mwGeneric = append(mwGeneric, mw)
	}
	return func(o *serverOpts) {
		o.HTTPMiddlewares = mwGeneric
	}
}

// WithGRPCUnaryMiddlewares sets up unary middlewares for gRPC server.
func WithGRPCUnaryMiddlewares(mws ...grpc.UnaryServerInterceptor) Option {
	mw := grpc_middleware.ChainUnaryServer(mws...)
	return func(o *serverOpts) {
		o.GRPCOpts = append(o.GRPCOpts, grpc.UnaryInterceptor(mw))
		o.GRPCUnaryInterceptor = mw
	}
}

// WithGRPCStreamMiddlewares sets up stream middlewares for gRPC server.
func WithGRPCStreamMiddlewares(mws ...grpc.StreamServerInterceptor) Option {
	return func(o *serverOpts) {
		o.GRPCOpts = append(o.GRPCOpts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(mws...)))
	}
}

// WithHTTPMux sets existing HTTP muxer to use instead of creating new one.
func WithHTTPMux(mux *chi.Mux) Option {
	return func(o *serverOpts) {
		o.HTTPMux = mux
	}
}

func EnableReflection(enable bool) Option {
	return func(o *serverOpts) {
		o.EnableReflection = enable
	}
}

func WithRuntimeServeMuxOpts(opts ...runtime.ServeMuxOption) Option {
	return func(o *serverOpts) {
		o.RuntimeServeMuxOpts = append(o.RuntimeServeMuxOpts, opts...)
	}
}
