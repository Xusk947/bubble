package httpserver

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

const requestIDLocalKey = "requestid"

func requestID(c fiber.Ctx) string {
	if c == nil {
		return ""
	}
	if v := c.Locals(requestIDLocalKey); v != nil {
		return fmt.Sprint(v)
	}
	return c.Get("X-Request-ID")
}

