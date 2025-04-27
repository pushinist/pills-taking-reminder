package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"pills-taking-reminder/internal/config"
	grpcServer "pills-taking-reminder/internal/grpc/server"
	"pills-taking-reminder/internal/server"
	"pills-taking-reminder/internal/storage/pg"
	"pills-taking-reminder/pkg/logger"
	"sync"
	"syscall"

	s "pills-taking-reminder/internal/service"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("logger initialized, starting pills-taking-reminder", slog.String("env", cfg.Env))

	log.Info("config", slog.Any("", cfg))

	db, err := pg.New(cfg.DB, log, cfg.NearTakingInterval)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("storage initialized")

	service := s.NewService(db, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}

	httpSrv := server.NewServer(service)
	httpSrv.RegisterRoutes()

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Info("starting HTTP Server", slog.String("address", cfg.HTTPServer.Address))

		err = httpSrv.Run(cfg.HTTPServer.Address)
		if err != nil {
			log.Error("failed to start http server", slog.String("error", err.Error()))
			cancel()
		}
	}()

	grpcSrv := grpcServer.NewGRPCServer(service, log)

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Info("starting gRPC Server", slog.String("address", cfg.GRPCServer.Address))

		err = grpcSrv.Run(cfg.GRPCServer.Address)
		if err != nil {
			log.Error("failed to start grpc server", slog.String("error", err.Error()))
			cancel()
		}
	}()

	log.Info("both server started")

	select {
	case <-sigChan:
		log.Info("got shutdown signal")
	case <-ctx.Done():
		log.Info("got context cancelled")
	}

	log.Info("shutting both servers down...")

	httpSrv.Stop()
	grpcSrv.Stop()

	wg.Wait()
	log.Info("app down!")
}
