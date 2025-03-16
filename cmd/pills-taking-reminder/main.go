package main

import (
	"log/slog"
	"os"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/server"
	"pills-taking-reminder/internal/storage/pg"
	"pills-taking-reminder/pkg/logger"

	s "pills-taking-reminder/internal/service"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("logger initialized, starting pills-taking-reminder", slog.String("env", cfg.Env))

	db, err := pg.New(cfg.DB, cfg.NearTakingInterval)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("storage initialized")

	service := s.NewService(db)

	srv := server.NewServer(service)
	srv.RegisterRoutes()
	err = srv.Run(":8080")
	if err != nil {
		log.Error("failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
