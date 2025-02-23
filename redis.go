package main

import (
	"context"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type RedisClient struct {
	client  *redis.Client
	limiter *redis_rate.Limiter
}

type LimiterResponse struct {
	Allowed   bool
	Remaining int
}

func NewRedis(
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (RedisClient, error) {
	ctx := context.Background()

	server, err := miniredis.Run()
	if err != nil {
		return RedisClient{}, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	err = redisotel.InstrumentTracing(client, redisotel.WithTracerProvider(tracerProvider))
	if err != nil {
		return RedisClient{}, err
	}

	err = redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(meterProvider))
	if err != nil {
		return RedisClient{}, err
	}

	return RedisClient{
		client:  client,
		limiter: redis_rate.NewLimiter(client),
	}, client.FlushDB(ctx).Err()
}

func (r RedisClient) Allow(ctx context.Context, key string) (LimiterResponse, error) {
	res, err := r.limiter.Allow(ctx, key, redis_rate.PerMinute(10))
	if err != nil {
		return LimiterResponse{}, err
	}

	return LimiterResponse{Allowed: res.Allowed == 1, Remaining: res.Remaining}, nil
}
