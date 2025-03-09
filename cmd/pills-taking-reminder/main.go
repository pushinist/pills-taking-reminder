package main

import (
	"log/slog"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("logger initialized, starting pills-taking-reminder", slog.String("env", cfg.Env))

}
