package otel

import (
	"context"
	"net/http"

	"github.com/richer421/q-deploy/conf"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var promHandler http.Handler

// PrometheusHandler returns the HTTP handler for /metrics endpoint.
// Returns nil if OTel is disabled or Prometheus exporter is not initialized.
func PrometheusHandler() http.Handler {
	return promHandler
}

// Init initializes OpenTelemetry TracerProvider and MeterProvider.
// Returns a shutdown function that should be called on application exit.
// If cfg.Enabled is false, returns a no-op shutdown function.
func Init(cfg conf.OTelConfig) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if !cfg.Enabled {
		return noop, nil
	}

	ctx := context.Background()

	// Resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return noop, err
	}

	// --- Trace ---
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return noop, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// --- Metric ---
	var metricReaders []sdkmetric.Reader

	// OTLP metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return noop, err
	}
	metricReaders = append(metricReaders, sdkmetric.NewPeriodicReader(metricExporter))

	// Prometheus exporter
	if cfg.Prometheus.Enabled {
		promExporter, err := prometheus.New()
		if err != nil {
			return noop, err
		}
		metricReaders = append(metricReaders, promExporter)
	}

	var mpOpts []sdkmetric.Option
	mpOpts = append(mpOpts, sdkmetric.WithResource(res))
	for _, r := range metricReaders {
		mpOpts = append(mpOpts, sdkmetric.WithReader(r))
	}
	mp := sdkmetric.NewMeterProvider(mpOpts...)
	otel.SetMeterProvider(mp)

	// Prometheus HTTP handler
	if cfg.Prometheus.Enabled {
		// prometheus exporter implements http.Handler via promhttp
		// We use the SDK's built-in ServeHTTP from the exporter
		for _, r := range metricReaders {
			if h, ok := r.(http.Handler); ok {
				promHandler = h
				break
			}
		}
	}

	// Shutdown function
	shutdown := func(ctx context.Context) error {
		var firstErr error
		if err := tp.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := mp.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	return shutdown, nil
}
