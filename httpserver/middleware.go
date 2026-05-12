package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func zapLoggerMiddleware(logger *zap.Logger) fiber.Handler {
	if logger == nil {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		reqID := requestID(c)
		status := c.Response().StatusCode()

		logger.Info(
			"http request",
			zap.String("request_id", reqID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("remote_ip", c.IP()),
			zap.Int("bytes_out", len(c.Response().Body())),
		)

		return err
	}
}

