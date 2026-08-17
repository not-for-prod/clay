package mwhttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type serverMetricsOptions struct {
	namespace string
	subsystem string
}

type ServerMetricsOption func(*serverMetricsOptions)

func newServerMetricsOptions(opts ...ServerMetricsOption) *serverMetricsOptions {
	o := &serverMetricsOptions{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithNamespace allows you to add a Namespace to metrics.
func WithNamespace(namespace string) ServerMetricsOption {
	return func(o *serverMetricsOptions) {
		o.namespace = namespace
	}
}

// WithSubsystem allows you to add a Subsystem to metrics.
func WithSubsystem(subsystem string) ServerMetricsOption {
	return func(o *serverMetricsOptions) {
		o.subsystem = subsystem
	}
}

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a HTTP server.
type ServerMetrics struct {
	serverStartedCounter *prometheus.CounterVec
	serverHandledCounter *prometheus.CounterVec
	// serverHandledHistogram can be nil.
	serverHandledHistogram *prometheus.HistogramVec
}

func NewServerMetrics(opts ...ServerMetricsOption) *ServerMetrics {
	serverMetricsOpts := newServerMetricsOptions(opts...)
	defaultLabels := []string{"http_method", "http_path"}
	defaultLabelsWithCode := []string{"http_method", "http_path", "http_code"}

	return &ServerMetrics{
		serverStartedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serverMetricsOpts.namespace,
				Subsystem: serverMetricsOpts.subsystem,
				Name:      "http_server_started_total",
				Help:      "Total number of requests started on the server.",
			},
			defaultLabels,
		),
		serverHandledCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serverMetricsOpts.namespace,
				Subsystem: serverMetricsOpts.subsystem,
				Name:      "http_server_handled_total",
				Help:      "Total number of requests completed on the server, regardless of success or failure.",
			},
			defaultLabelsWithCode,
		),
		serverHandledHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: serverMetricsOpts.namespace,
				Subsystem: serverMetricsOpts.subsystem,
				Name:      "http_server_handling_seconds",
				Help:      "Histogram of response latency (seconds) of requests handled by the server.",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
			},
			defaultLabels,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	if m.serverHandledHistogram != nil {
		m.serverHandledHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	if m.serverHandledHistogram != nil {
		m.serverHandledHistogram.Collect(ch)
	}
}

// Middleware records HTTP server metrics using bounded route templates.
// Register Chi middleware that changes RoutePath or RouteMethod before it.
func (m *ServerMetrics) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				startedAt := time.Now()
				lwr := newLoggingResponseWriter(w)
				method := r.Method
				r, routeState := withRouteTemplateState(r, func(template string) {
					m.serverStartedCounter.WithLabelValues(method, template).Inc()
				})
				routeState.set(matchedChiRouteTemplate(r))
				completed := false

				defer func() {
					template := routeTemplate(r, routeState, lwr.statusCode)
					routeState.set(template)
					if !completed {
						return
					}

					m.serverHandledCounter.WithLabelValues(method, template, strconv.Itoa(lwr.statusCode)).Inc()
					m.serverHandledHistogram.WithLabelValues(method, template).Observe(time.Since(startedAt).Seconds())
				}()

				next.ServeHTTP(lwr, r)
				completed = true
			},
		)
	}
}

// Source - https://stackoverflow.com/a/53272925
// Posted by huangapple
// Retrieved 2026-03-16, License - CC BY-SA 4.0

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
