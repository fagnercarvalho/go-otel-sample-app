package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"go-otel-sample-app/otel"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func runClient(otelCollectorURL, serviceName, baseURL string) {
	var (
		rootSpanName = "client-span"
	)

	providers, err := otel.NewProviders(otelCollectorURL, serviceName)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := providers.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	logger := otelslog.NewLogger("client", otelslog.WithLoggerProvider(providers.LoggerProvider))

	transport := otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithTracerProvider(otelpyroscope.NewTracerProvider(providers.TraceProvider)),
		otelhttp.WithMeterProvider(providers.MeterProvider),
	)

	client := http.Client{Transport: transport}

	for {
		ctx := context.Background()

		tracer := providers.TraceProvider.Tracer(serviceName)
		newCtx, span := tracer.Start(ctx, rootSpanName)

		logger.InfoContext(newCtx, "Running 2 requests on Todo Service")

		err := doRequest(newCtx, client, baseURL, http.MethodGet, nil)
		if err != nil {
			logger.ErrorContext(ctx, "Error while doing GET request", "error", err)
		}

		b, err := json.Marshal(
			Todo{
				ID:   rand.Intn(100),
				Task: "Task " + fmt.Sprint(rand.Intn(50)),
				Done: rand.Float32() < 0.5,
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		err = doRequest(newCtx, client, baseURL, http.MethodPost, bytes.NewReader(b))
		if err != nil {
			logger.ErrorContext(ctx, "Error while doing POST request", "error", err)
		}

		// fake slowness
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		logger.InfoContext(newCtx, "Ran 2 requests on Todo Service")

		span.End()
	}
}

func doRequest(ctx context.Context, client http.Client, baseUrl, method string, body io.Reader) error {
	request, err := http.NewRequestWithContext(ctx, method, baseUrl+"/todos", body)
	if err != nil {
		return err
	}

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	res, err := client.Do(request)
	if err != nil {
		return err
	}

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	_ = res.Body.Close()

	if res.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("Error while doing request: %v (%v)", res.StatusCode, res.Status))
	}

	return nil
}
