package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/server"
	"github.com/not-for-prod/clay/server/log"
	"github.com/not-for-prod/clay/server/middlewares/mwgrpc"
	"github.com/sirupsen/logrus"
	sum "github.com/utrack/clay/doc/example/implementation"
	example "github.com/utrack/clay/doc/example/pb"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Create service
	service := sum.NewSummator()

	err := server.NewServer(
		12345,
		// Recover from both HTTP and gRPC panics and use our own middleware
		server.WithGRPCUnaryMiddlewares(mwgrpc.UnaryPanicHandler(log.Default)),
		server.WithRuntimeServeMuxOpts(
			// Remove runtime.MetadataHeaderPrefix for Set-Cookie headers.
			runtime.WithOutgoingHeaderMatcher(
				func(s string) (string, bool) {
					if s == "set-cookie" {
						return "set-cookie", true
					}

					return runtime.MetadataHeaderPrefix + s, false
				},
			),
			// Extract Cookie data into metadata.
			runtime.WithMetadata(
				func(ctx context.Context, req *http.Request) metadata.MD {
					tokenCookie, _ := req.Cookie("summator-session")
					if tokenCookie == nil {
						return nil // metadata.Pairs()
					}

					return metadata.Pairs("summator-session", tokenCookie.Value)
				},
			),
		),
	).Run(example.NewSummatorServiceDesc(service))
	if err != nil {
		logrus.Fatal(err)
	}
}
