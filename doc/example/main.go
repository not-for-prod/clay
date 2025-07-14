package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	sum "github.com/utrack/clay/doc/example/implementation"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/utrack/clay/v3/log"
	"github.com/utrack/clay/v3/transport/middlewares/mwgrpc"
	"github.com/utrack/clay/v3/transport/server"
)

func main() {
	ctx := context.Background()
	hmux := chi.NewMux()
	hmux.HandleFunc(
		"/docs/*", func(w http.ResponseWriter, r *http.Request) {
			httpSwagger.Handler(httpSwagger.URL("swagger.json")).ServeHTTP(w, r)
		},
	)
	hmux.Get(
		"/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
		},
	)
	hmux.Get(
		"/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger.json", http.StatusMovedPermanently)
		},
	)

	// Wire up our bundled Swagger UI
	impl := sum.NewSummator()
	srv := server.NewServer(
		12345,
		// Pass our mux with Swagger UI
		server.WithHTTPMux(hmux),
		// Recover from both HTTP and gRPC panics and use our own middleware
		server.WithGRPCUnaryMiddlewares(mwgrpc.UnaryPanicHandler(log.Default)),
	)
	err := srv.Run(ctx, impl)
	if err != nil {
		logrus.Fatal(err)
	}
}
