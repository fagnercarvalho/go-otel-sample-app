package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

type Providers struct {
	LoggerProvider *log.LoggerProvider
	TraceProvider  *trace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

func NewProviders(otelCollectorURL, serviceName string) (Providers, error) {
	loggerProvider, err := newLoggerProvider(otelCollectorURL, serviceName)
	if err != nil {
		return Providers{}, err
	}

	tracerProvider, err := newTracerProvider(otelCollectorURL, serviceName)
	if err != nil {
		return Providers{}, err
	}

	meterProvider, err := newMeterProvider(otelCollectorURL, serviceName)
	if err != nil {
		return Providers{}, err
	}

	return Providers{
		LoggerProvider: loggerProvider,
		TraceProvider:  tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}

func SetPropagators() {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)
}

func (p *Providers) Shutdown(ctx context.Context) error {
	err := p.LoggerProvider.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = p.TraceProvider.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = p.MeterProvider.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p *Providers) GetCounter() (metric.Int64Counter, error) {
	var meter = p.MeterProvider.Meter("github.com/fagnercarvalho/go-otel-sample-app")

	return meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
}

func newTracerProvider(otelCollectorURL, serviceName string) (*trace.TracerProvider, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(otelCollectorURL),
		otlptracehttp.WithInsecure(),
	)

	traceExporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(newResource(serviceName)),
	)

	return provider, nil
}

func newLoggerProvider(otelCollectorURL, serviceName string) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(
		context.Background(),
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

func newMeterProvider(otelCollectorURL, serviceName string) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(
		context.Background(),
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
