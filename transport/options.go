package transport

import (
	"github.com/not-for-prod/clay/transport/httptransport"
	"google.golang.org/grpc"
)

// DescOption modifies the ServiceDesc's behaviour.
type DescOption interface {
	Apply(*httptransport.DescOptions)
}

// WithUnaryInterceptor sets up the interceptor for incoming calls.
func WithUnaryInterceptor(i grpc.UnaryServerInterceptor) DescOption {
	return httptransport.OptionUnaryInterceptor{Interceptor: i}
}
