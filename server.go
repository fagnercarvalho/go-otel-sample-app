package main

import (
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

var logger *slog.Logger

func registerRoutes(
	e *echo.Echo,
	db DB,
	redis RedisClient,
	counter metric.Int64Counter,
	tracerProvider trace.TracerProvider,
) {
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("server", otelecho.WithTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))))
	e.Use(RateLimiterMiddleware(redis))

	e.GET("/todos", getTodos(db, counter))
	e.POST("/todos", addTodo(db, counter))
}

func getTodos(db DB, counter metric.Int64Counter) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		counter.Add(ctx, 1)

		// fake slowness
		time.Sleep(time.Duration(rand.Intn(2)) * time.Second)

		logger.InfoContext(ctx, "Handling GET /todos request")

		todos, err := db.FetchAllTodos(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Error while fetching todos", "error", err)

			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, todos)
	}
}

func addTodo(db DB, counter metric.Int64Counter) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		counter.Add(ctx, 1)

		logger.InfoContext(ctx, "Handling POST /todos request")

		var todo Todo
		if err := c.Bind(&todo); err != nil {
			logger.Error("Error binding request", "error", err)

			return c.NoContent(http.StatusBadRequest)
		}

		err := db.AddTodo(ctx, todo.Task, todo.Done)
		if err != nil {
			logger.ErrorContext(ctx, "Error while adding todo", "error", err, "todo", todo)

			return c.NoContent(http.StatusInternalServerError)
		}

		// fake slowness
		time.Sleep(time.Duration(rand.Intn(3)) * time.Second)

		logger.InfoContext(ctx, "Added new todo", "todo", todo)

		return c.JSON(http.StatusCreated, todo)
	}
}
