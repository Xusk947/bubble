package httpserver

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/health"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/zap"
)

type HealthProvider interface {
	Health(ctx context.Context) health.Status
}

type Deps struct {
	Logger        *zap.Logger
	Health        HealthProvider
	ErrorToStatus func(error) int
}

type HTTPServer struct {
	App      *fiber.App
	Config   config.HTTPConfig
	Listener net.Listener
	Logger   *zap.Logger
}

func NewHTTPServer(cfg config.HTTPConfig, deps Deps) (*HTTPServer, error) {
	logger := deps.Logger

	errorToStatus := deps.ErrorToStatus
	if errorToStatus == nil {
		errorToStatus = defaultErrorToStatus
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return writeError(c, err, errorToStatus(err))
		},
	})

	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-ID",
	}))
	app.Use(recover.New())
	app.Use(zapLoggerMiddleware(logger))

	if cfg.EnableCORS {
		app.Use(cors.New())
	}
	if cfg.EnablePprof {
		app.Use(pprof.New())
	}

	if deps.Health != nil {
		app.Get("/health/live", func(c fiber.Ctx) error {
			s := deps.Health.Health(context.Background())
			return c.Status(statusForLive(s)).JSON(s)
		})
		app.Get("/health/ready", func(c fiber.Ctx) error {
			s := deps.Health.Health(context.Background())
			return c.Status(statusForHealth(s)).JSON(s)
		})
	}

	return &HTTPServer{
		App:    app,
		Config: cfg,
		Logger: logger,
	}, nil
}

func (s *HTTPServer) Start(ctx context.Context) error {
	_ = ctx
	if s.Listener != nil {
		return errors.New("http server already started")
	}

	ln, err := net.Listen("tcp", s.Config.Address)
	if err != nil {
		return err
	}
	s.Listener = ln

	if s.Logger != nil {
		s.Logger.Info("http listening", zap.String("http_address", s.Config.Address))
	}

	go func() {
		_ = s.App.Listener(ln)
	}()

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s == nil || s.App == nil {
		return nil
	}
	if s.Listener == nil {
		return nil
	}

	stopCtx := ctx
	if stopCtx == nil {
		stopCtx = context.Background()
	}
	if deadline, ok := stopCtx.Deadline(); ok {
		_ = deadline
	} else {
		var cancel context.CancelFunc
		stopCtx, cancel = context.WithTimeout(stopCtx, 10*time.Second)
		defer cancel()
	}

	err := s.App.ShutdownWithContext(stopCtx)
	_ = s.Listener.Close()
	s.Listener = nil
	return err
}
