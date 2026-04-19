//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/uudecode/k8s-bot/internal/di"
	"github.com/uudecode/k8s-bot/internal/monitor"
)

func InitializeApp() (*monitor.Service, error) {
	wire.Build(di.SuperSet)
	return &monitor.Service{}, nil
}
