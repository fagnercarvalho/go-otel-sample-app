package main

import (
	"context"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"go-otel-sample-app/otel"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	otel.SetPropagators()

	var (
		otelCollectorURL  = "otel.local"
		pyroscopeURL      = "http://pyroscope.local"
		baseURL           = "http://localhost:8080"
		serverServiceName = "todo-service"
		clientServiceName = "todo-service-client"
		dbServiceName     = "todo-service-db"
		loggerName        = "server"
	)

	otelCollectorURLVar := os.Getenv("OTEL_COLLECTOR_URL")
	if otelCollectorURLVar != "" {
		otelCollectorURL = otelCollectorURLVar
	}

	pyroscopeURLVar := os.Getenv("PYROSCOPE_URL")
	if pyroscopeURLVar != "" {
		pyroscopeURL = pyroscopeURLVar
	}

	providers, err := otel.NewProviders(otelCollectorURL, serverServiceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := providers.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	profiler, err := startProfiler(pyroscopeURL, serverServiceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := profiler.Stop()
		if err != nil {
			panic(err)
		}
	}()

	logger = otelslog.NewLogger(
		loggerName,
		otelslog.WithLoggerProvider(providers.LoggerProvider),
	)

	db, err := NewDB(dbServiceName, providers.TraceProvider, providers.MeterProvider)
	if err != nil {
		panic(err)
	}

	redis, err := NewRedis(providers.TraceProvider, providers.MeterProvider)
	if err != nil {
		panic(err)
	}

	counter, err := providers.GetCounter()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	registerRoutes(e, db, redis, counter, providers.TraceProvider)

	logger.Info("Server running on :8080")

	go runClient(otelCollectorURL, clientServiceName, baseURL)

	e.Logger.Fatal(e.Start(strings.TrimPrefix(baseURL, "http://")))
}
