package app

import (
	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/httpserver"

	"go.uber.org/zap"
)

type Option func(*options)

type options struct {
	dotenvPath string
	config     *config.Config
	logger     *zap.Logger
	httpSetup  []func(*httpserver.HTTPServer)
}

func WithDotenvPath(path string) Option {
	return func(o *options) {
		o.dotenvPath = path
	}
}

func WithConfig(value config.Config) Option {
	return func(o *options) {
		v := value
		o.config = &v
	}
}

func WithLogger(value *zap.Logger) Option {
	return func(o *options) {
		o.logger = value
	}
}

func WithHTTPSetup(fn func(*httpserver.HTTPServer)) Option {
	return func(o *options) {
		if fn == nil {
			return
		}
		o.httpSetup = append(o.httpSetup, fn)
	}
}
