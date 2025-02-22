package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

func newTracerProvider(serviceName string) (*trace.TracerProvider, error) {
	ctx := context.Background()

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(otelCollectorURL),
		otlptracehttp.WithInsecure(),
	)

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(newResource(serviceName)),
	)

	return provider, nil
}

func newLoggerProvider(serviceName string) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpoint(otelCollectorURL),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	processor := log.NewBatchProcessor(exporter)

	provider := log.NewLoggerProvider(
		log.WithResource(newResource(serviceName)),
		log.WithProcessor(processor),
	)

	return provider, nil
}

func newMeterProvider(serviceName string) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpoint(otelCollectorURL),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(newResource(serviceName)),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(3*time.Second)),
		),
		sdkmetric.WithExemplarFilter(exemplar.AlwaysOnFilter),
	)

	return provider, nil
}

func newResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)
}

func setPropagator() {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)
}
