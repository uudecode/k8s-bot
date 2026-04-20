package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/uudecode/k8s-bot/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("application failed")
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service, err := app.InitializeApp()
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}

	if err := service.Run(ctx); err != nil {
		return fmt.Errorf("run application: %w", err)
	}

	return nil
}
