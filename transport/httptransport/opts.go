package httptransport

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// DescOptions provides options for a ServiceDesc compiled code.
type DescOptions struct {
	UnaryInterceptor grpc.UnaryServerInterceptor
}

// OptionUnaryInterceptor sets up the gRPC unary interceptor.
type OptionUnaryInterceptor struct {
	Interceptor grpc.UnaryServerInterceptor
}

// Apply implements transport.DescOption.
func (o OptionUnaryInterceptor) Apply(oo *DescOptions) {
	if oo.UnaryInterceptor != nil {
		oo.UnaryInterceptor = grpc_middleware.ChainUnaryServer(
			oo.UnaryInterceptor,
			o.Interceptor,
		)
		return
	}
	oo.UnaryInterceptor = o.Interceptor
}
