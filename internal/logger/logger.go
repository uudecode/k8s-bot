package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/uudecode/k8s-bot/internal/config"
)

func NewLogger(cfg *config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.App.LogLevel))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	return zerolog.New(output).With().Timestamp().Logger()
	//return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
