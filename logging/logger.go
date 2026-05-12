package logging

import (
	"go.uber.org/zap"
)

type Encoding int

const (
	EncodingUnspecified Encoding = iota
	EncodingJSON
	EncodingConsole
)

type OutputConfig struct {
	Stdout bool
	Stderr bool
}

type Config struct {
	Level       Level
	Development bool
	Encoding    Encoding
	Output      OutputConfig
}

func DefaultConfig() Config {
	return Config{
		Level:       LevelInfo,
		Development: false,
		Encoding:    EncodingJSON,
		Output: OutputConfig{
			Stdout: true,
			Stderr: false,
		},
	}
}

func New(cfg Config, opts ...zap.Option) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	}

	zapCfg.Level = zap.NewAtomicLevelAt(cfg.Level.ZapLevel())
	zapCfg.Encoding = encodingString(cfg.Encoding)
	zapCfg.OutputPaths = outputPaths(cfg.Output)
	zapCfg.ErrorOutputPaths = outputPaths(OutputConfig{Stdout: false, Stderr: true})

	return zapCfg.Build(opts...)
}

func encodingString(value Encoding) string {
	switch value {
	case EncodingConsole:
		return "console"
	case EncodingJSON, EncodingUnspecified:
		return "json"
	default:
		return "json"
	}
}

func outputPaths(value OutputConfig) []string {
	const stdoutPath = "stdout"
	const stderrPath = "stderr"

	paths := make([]string, 0, 2)
	if value.Stdout {
		paths = append(paths, stdoutPath)
	}
	if value.Stderr {
		paths = append(paths, stderrPath)
	}
	if len(paths) == 0 {
		paths = append(paths, stdoutPath)
	}
	return paths
}
