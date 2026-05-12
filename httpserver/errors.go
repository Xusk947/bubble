package httpserver

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func defaultErrorToStatus(err error) int {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}
	return fiber.StatusInternalServerError
}

func writeError(c fiber.Ctx, err error, status int) error {
	msg := "internal error"
	var fe *fiber.Error
	if errors.As(err, &fe) {
		msg = fe.Message
	}
	resp := ErrorResponse{
		Code:      status,
		Message:   msg,
		RequestID: requestID(c),
	}
	return c.Status(status).JSON(resp)
}

