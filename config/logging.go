package config

import "github.com/Xusk947/bubble/logging"

func (c LogConfig) AsLoggingConfig() logging.Config {
	return logging.Config{
		Level:       c.Level,
		Development: c.Development,
		Encoding:    c.Encoding,
		Output:      c.Output,
	}
}
