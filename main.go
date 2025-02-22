package main

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/metric"
)

var (
	otelCollectorURL = "otel.local"
	baseUrl          = "http://localhost:8080"
)

func main() {
	setPropagator()

	serviceName := "todo-service"

	tracerProvider, err := newTracerProvider(serviceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := tracerProvider.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	loggerProvider, err := newLoggerProvider(serviceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := loggerProvider.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	meterProvider, err := newMeterProvider(serviceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := meterProvider.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	profiler, err := startProfiler()
	if err != nil {
		panic(err)
	}

	defer func() {
		err := profiler.Stop()
		if err != nil {
			panic(err)
		}
	}()

	logger = otelslog.NewLogger("server", otelslog.WithLoggerProvider(loggerProvider))

	var meter = meterProvider.Meter("github.com/fagnercarvalho/go-otel-sample-app")

	counter, err := meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	db, err := NewDB(tracerProvider, meterProvider)
	if err != nil {
		panic(err)
	}

	err = db.Initialize()
	if err != nil {
		panic(err)
	}

	redis, err := NewRedis(tracerProvider, meterProvider)
	if err != nil {
		panic(err)
	}

	registerRoutes(e, db, redis, counter, tracerProvider)

	logger.Info("Server running on :8080")

	go runClient(baseUrl)

	e.Logger.Fatal(e.Start(":8080"))
}
