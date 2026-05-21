package main

import (
	"log/slog"

	"butterfly.orx.me/core/app"
	"github.com/kongken/kube-dash/internal/dashboard"
)

type Config struct{}

func (Config) Print() {}

func main() {
	server := dashboard.NewServer()

	application := app.New(&app.Config{
		Service:   "kube-dash",
		Namespace: "kube-dash",
		Config:    &Config{},
		Router:    server.RegisterRoutes,
		InitFunc: []func() error{
			server.Init,
		},
		TeardownFunc: []func() error{
			server.Close,
		},
	})

	slog.Info("starting kube-dash")
	application.Run()
}
