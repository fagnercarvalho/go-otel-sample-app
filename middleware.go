package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RateLimiterMiddleware(client RedisClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			response, err := client.Allow(ctx, c.Request().Method+c.Path())
			if err != nil {
				return err
			}

			if response.Allowed {
				logger.InfoContext(ctx, "Request allowed by rate limiter", "response", response)

				return next(c)
			}

			logger.WarnContext(ctx, "Request not allowed by rate limiter", "response", response)

			return c.NoContent(http.StatusTooManyRequests)
		}
	}
}
