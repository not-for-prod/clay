package main

import (
	"context"

	"github.com/sirupsen/logrus"
	sum "github.com/utrack/clay/doc/example/implementation"

	"github.com/not-for-prod/clay/log"
	"github.com/not-for-prod/clay/transport/middlewares/mwgrpc"
	"github.com/not-for-prod/clay/transport/server"
)

func main() {
	ctx := context.Background()
	// Wire up our bundled Swagger UI
	impl := sum.NewSummator()
	srv := server.NewServer(
		12345,
		// Recover from both HTTP and gRPC panics and use our own middleware
		server.WithGRPCUnaryMiddlewares(mwgrpc.UnaryPanicHandler(log.Default)),
	)
	err := srv.Run(ctx, impl)
	if err != nil {
		logrus.Fatal(err)
	}
}
