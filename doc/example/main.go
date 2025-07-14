package main

import (
	"github.com/not-for-prod/clay/transport"
	"github.com/sirupsen/logrus"
	sum "github.com/utrack/clay/doc/example/implementation"

	"github.com/not-for-prod/clay/server"
	"github.com/not-for-prod/clay/server/log"
	"github.com/not-for-prod/clay/server/middlewares/mwgrpc"
)

func main() {
	// Create service
	service := sum.NewSummator()
	// Join several services in compound service supporting both contracts
	// swaggers will be merges
	multipleServices := transport.NewCompoundService(service)

	err := server.NewServer(
		12345,
		// Recover from both HTTP and gRPC panics and use our own middleware
		server.WithGRPCUnaryMiddlewares(mwgrpc.UnaryPanicHandler(log.Default)),
		server.EnableReflection(true),
	).Run(multipleServices)
	if err != nil {
		logrus.Fatal(err)
	}
}
