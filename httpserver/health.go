package httpserver

import (
	"github.com/Xusk947/bubble/health"

	"github.com/gofiber/fiber/v3"
)

func statusForHealth(s health.Status) int {
	if s.Ready {
		return fiber.StatusOK
	}
	return fiber.StatusServiceUnavailable
}

func statusForLive(s health.Status) int {
	if s.Live {
		return fiber.StatusOK
	}
	return fiber.StatusServiceUnavailable
}
