package mwgrpc

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
)

var defaultBuckets = []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}

type serverMetricsOptions struct {
	namespace string
	subsystem string
	buckets   []float64
}

type ServerMetricsOption func(*serverMetricsOptions)

func newServerMetricsOptions(opts ...ServerMetricsOption) *serverMetricsOptions {
	o := &serverMetricsOptions{
		buckets: defaultBuckets,
	}

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

func WithHistogramBuckets(buckets []float64) ServerMetricsOption {
	return func(o *serverMetricsOptions) {
		o.buckets = buckets
	}
}

func NewServerMetrics(opts ...ServerMetricsOption) *grpcprom.ServerMetrics {
	var serverCounterOptions []grpcprom.CounterOption
	var serverHistogramOptions []grpcprom.HistogramOption

	serverMetricsOpts := newServerMetricsOptions(opts...)

	serverHistogramOptions = append(
		serverHistogramOptions, grpcprom.WithHistogramBuckets(
			serverMetricsOpts.buckets,
		),
	)

	if serverMetricsOpts.namespace != "" {
		serverCounterOptions = append(serverCounterOptions, grpcprom.WithNamespace(serverMetricsOpts.namespace))
		serverHistogramOptions = append(
			serverHistogramOptions,
			grpcprom.WithHistogramNamespace(serverMetricsOpts.namespace),
		)
	}

	if serverMetricsOpts.subsystem != "" {
		serverCounterOptions = append(serverCounterOptions, grpcprom.WithSubsystem(serverMetricsOpts.subsystem))
		serverHistogramOptions = append(
			serverHistogramOptions,
			grpcprom.WithHistogramSubsystem(serverMetricsOpts.subsystem),
		)
	}

	return grpcprom.NewServerMetrics(
		grpcprom.WithServerCounterOptions(serverCounterOptions...),
		grpcprom.WithServerHandlingTimeHistogram(serverHistogramOptions...),
	)
}
