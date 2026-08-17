package mwhttp

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMiddlewareUsesChiRouteTemplate(t *testing.T) {
	registry, metrics := newTestMetrics(t)
	router := chi.NewRouter()
	router.Use(metrics.Middleware())
	router.Get("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		assertMetricRouteTemplatesForFamilies(
			t,
			registry,
			[]string{"http_server_started_total"},
			"/users/{id}",
		)
		w.WriteHeader(http.StatusNoContent)
	})

	serveRequests(router, "/users/123", "/users/456")

	assertMetricRouteTemplates(t, registry, "/users/{id}")
}

func TestMiddlewareUsesSameEscapedPathAsChiRouter(t *testing.T) {
	registry, metrics := newTestMetrics(t)
	router := chi.NewRouter()
	router.Use(metrics.Middleware())
	router.Get("/files/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.Get("/files/{directory}/{file}", func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("request matched the decoded path instead of URL.RawPath")
	})

	serveRequests(router, "/files/a%2Fb")

	assertMetricRouteTemplates(t, registry, "/files/{id}")
}

func TestMiddlewareStartsChiWildcardRouteBeforeHandler(t *testing.T) {
	registry, metrics := newTestMetrics(t)
	router := chi.NewRouter()
	router.Use(metrics.Middleware())
	router.Get("/assets/*", func(w http.ResponseWriter, _ *http.Request) {
		assertMetricRouteTemplatesForFamilies(
			t,
			registry,
			[]string{"http_server_started_total"},
			"/assets/*",
		)
		w.WriteHeader(http.StatusNoContent)
	})

	serveRequests(router, "/assets/scripts/app.js")

	assertMetricRouteTemplates(t, registry, "/assets/*")
}

func TestMiddlewareUsesGRPCGatewayRouteTemplate(t *testing.T) {
	registry, metrics := newTestMetrics(t)
	mux := runtime.NewServeMux(GRPCGatewayRouteTemplateOption())
	err := mux.HandlePath(http.MethodGet, "/v1/users/{id}", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("register gateway route: %v", err)
	}

	serveRequests(metrics.Middleware()(mux), "/v1/users/123", "/v1/users/456")

	assertMetricRouteTemplates(t, registry, "/v1/users/{id=*}")
}

func TestGRPCGatewayRouteTemplateOptionWithoutMetricsState(t *testing.T) {
	mux := runtime.NewServeMux(GRPCGatewayRouteTemplateOption())
	err := mux.HandlePath(http.MethodGet, "/v1/users/{id}", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("register gateway route: %v", err)
	}

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/users/123", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestMiddlewareBoundsUnknownRoutes(t *testing.T) {
	registry, metrics := newTestMetrics(t)
	router := chi.NewRouter()
	router.Use(metrics.Middleware())
	router.Get("/known", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	serveRequests(router, "/missing/123", "/missing/456")

	assertMetricRouteTemplates(t, registry, notFoundRouteTemplate)
}

func newTestMetrics(t *testing.T) (*prometheus.Registry, *ServerMetrics) {
	t.Helper()

	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics()
	registry.MustRegister(metrics)

	return registry, metrics
}

func serveRequests(handler http.Handler, paths ...string) {
	for _, path := range paths {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func assertMetricRouteTemplates(t *testing.T, registry *prometheus.Registry, want ...string) {
	t.Helper()
	metricFamilyNames := []string{
		"http_server_started_total",
		"http_server_handled_total",
		"http_server_handling_seconds",
	}
	assertMetricRouteTemplatesForFamilies(t, registry, metricFamilyNames, want...)
}

func assertMetricRouteTemplatesForFamilies(
	t *testing.T,
	registry *prometheus.Registry,
	metricFamilyNames []string,
	want ...string,
) {
	t.Helper()

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	want = append([]string(nil), want...)
	sort.Strings(want)

	for _, familyName := range metricFamilyNames {
		var got []string
		for _, family := range metricFamilies {
			if family.GetName() != familyName {
				continue
			}

			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "http_path" {
						got = append(got, label.GetValue())
					}
				}
			}
		}

		sort.Strings(got)
		if !equalStrings(got, want) {
			t.Errorf("%s route templates = %v, want %v", familyName, got, want)
		}
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
