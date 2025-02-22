package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func runClient(baseUrl string) {
	serviceName := "todo-service-client"

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

	logger := otelslog.NewLogger("client", otelslog.WithLoggerProvider(loggerProvider))

	transport := otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider)),
		otelhttp.WithMeterProvider(meterProvider),
	)

	client := http.Client{Transport: transport}

	for {
		ctx := context.Background()

		tracer := tracerProvider.Tracer(serviceName)
		newCtx, span := tracer.Start(ctx, "client-span")

		logger.InfoContext(newCtx, "Running 2 requests on Todo Service")

		err := doRequest(newCtx, client, baseUrl, http.MethodGet, nil)
		if err != nil {
			log.Fatal(err)
		}

		b, err := json.Marshal(
			Todo{ID: rand.Intn(100),
				Task: "Task " + fmt.Sprint(rand.Intn(50)),
				Done: rand.Float32() < 0.5},
		)
		if err != nil {
			log.Fatal(err)
		}

		err = doRequest(newCtx, client, baseUrl, http.MethodPost, bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}

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

	return err
}
