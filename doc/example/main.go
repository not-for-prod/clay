package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/server"
	"github.com/not-for-prod/clay/server/log"
	"github.com/not-for-prod/clay/server/middlewares/mwgrpc"
	"github.com/not-for-prod/clay/server/middlewares/mwhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sum "github.com/utrack/clay/doc/example/implementation/summator-server"
	example "github.com/utrack/clay/doc/example/pb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Create service
	service := sum.NewImplementation()

	registry := prometheus.NewRegistry()
	grpcMetrics := mwgrpc.NewServerMetrics(registry)
	registry.MustRegister(grpcMetrics)
	httpMetrics := mwhttp.NewServerMetrics()
	registry.MustRegister(httpMetrics)

	mux := chi.NewMux()
	mux.Handle(
		"/metrics",
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)

	httpServer := http.Server{
		Handler: mux,
	}

	httpListener, err := (&net.ListenConfig{}).Listen(
		context.Background(),
		"tcp",
		net.JoinHostPort("", strconv.Itoa(54321)),
	)
	if err != nil {
		return
	}

	group, ctx := errgroup.WithContext(ctx)
	group.Go(
		func() error {
			return server.NewServer(
				12345,
				server.WithHTTPMiddlewares(
					httpMetrics.Middleware(),
				),
				// Recover from both HTTP and gRPC panics and use our own middleware
				server.WithGRPCMiddlewares(
					grpcMetrics.UnaryServerInterceptor(),
					mwgrpc.UnaryPanicHandler(log.Default),
				),
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
		},
	)
	group.Go(
		func() error {
			return httpServer.Serve(httpListener)
		},
	)

	err = group.Wait()
	if err != nil {
		slog.Error("failed to run group", err)
	}
}
