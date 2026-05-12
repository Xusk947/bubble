package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/health"

	"github.com/gofiber/fiber/v3"
)

type stubHealth struct {
	status health.Status
}

func (s stubHealth) Health(ctx context.Context) health.Status {
	_ = ctx
	return s.status
}

func TestHealthEndpoints(t *testing.T) {
	srv, err := NewHTTPServer(config.HTTPConfig{
		Address:      ":0",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, Deps{
		Health: stubHealth{status: health.Status{Live: true, Ready: false}},
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	resp, err := srv.App.Test(httpreq(t, "GET", "/health/live", ""), fiber.TestConfig{Timeout: time.Second})
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	resp, err = srv.App.Test(httpreq(t, "GET", "/health/ready", ""), fiber.TestConfig{Timeout: time.Second})
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestErrorHandler_ReturnsJSON(t *testing.T) {
	srv, err := NewHTTPServer(config.HTTPConfig{
		Address:      ":0",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, Deps{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	srv.App.Get("/err", func(c fiber.Ctx) error {
		return fiber.NewError(http.StatusBadRequest, "bad request")
	})

	resp, err := srv.App.Test(httpreq(t, "GET", "/err", ""), fiber.TestConfig{Timeout: time.Second})
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var decoded ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Code != http.StatusBadRequest {
		t.Fatalf("unexpected code: %d", decoded.Code)
	}
	if decoded.Message != "bad request" {
		t.Fatalf("unexpected message: %q", decoded.Message)
	}
}

func httpreq(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, "http://example"+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}
