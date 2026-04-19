package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/uudecode/k8s-bot/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service, err := app.InitializeApp()
	if err != nil {
		panic(fmt.Errorf("failed to initialize application: %w", err))
	}

	if err := service.Run(ctx); err != nil {
		panic(fmt.Errorf("application runtime error: %w", err))
	}
}
