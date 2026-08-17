package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/not-for-prod/clay/server/middlewares/mwhttp"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func TestDefaultServerCapturesGRPCGatewayRouteTemplate(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := mwhttp.NewServerMetrics()
	registry.MustRegister(metrics)

	testServer := NewServer(0, WithHTTPMiddlewares(metrics.Middleware()))
	testServer.serviceDesc = routeTemplateServiceDesc{}
	if err := testServer.initHTTPServer(); err != nil {
		t.Fatalf("initialize HTTP server: %v", err)
	}

	for _, path := range []string{"/v1/users/123", "/v1/users/456"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		testServer.httpServer.Handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	for _, family := range metricFamilies {
		if family.GetName() != "http_server_started_total" {
			continue
		}

		metrics := family.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("started metric series count = %d, want 1", len(metrics))
		}

		var gotPath string
		for _, label := range metrics[0].GetLabel() {
			if label.GetName() == "http_path" {
				gotPath = label.GetValue()
			}
		}
		if wantPath := "/v1/users/{id=*}"; gotPath != wantPath {
			t.Fatalf("http_path = %q, want %q", gotPath, wantPath)
		}

		if got := metrics[0].GetCounter().GetValue(); got != 2 {
			t.Fatalf("started metric value = %v, want 2", got)
		}

		return
	}

	t.Fatal("http_server_started_total was not collected")
}

type routeTemplateServiceDesc struct{}

func (routeTemplateServiceDesc) RegisterGRPC(*grpc.Server) {}

func (routeTemplateServiceDesc) RegisterHTTP(_ context.Context, mux *runtime.ServeMux) error {
	return mux.HandlePath(
		http.MethodGet,
		"/v1/users/{id}",
		func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func (routeTemplateServiceDesc) SwaggerDef() []byte {
	return nil
}
