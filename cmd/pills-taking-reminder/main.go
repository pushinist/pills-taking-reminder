package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/infrastructure/container"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	container, err := container.New(cfg)
	if err != nil {
		slog.Error("failed to init container",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := container.Logger
	log.Info("pills-taking-reminder starting",
		slog.String("env", cfg.Env),
		slog.String("http_address", cfg.HTTPServer.Address),
		slog.String("grpc_adress", cfg.GRPCServer.Address))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	container.HTTPHandler.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("starting http server", slog.String("address", cfg.HTTPServer.Address))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start http server", slog.String("error", err.Error()))
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("starting gRPC server", slog.String("address", cfg.GRPCServer.Address))
		if err := container.GRPCServer.Run(cfg.GRPCServer.Address); err != nil {
			log.Error("failed to start gRPC server", slog.String("error", err.Error()))
			cancel()
		}
	}()

	log.Info("both servers have been started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("got shutdown signal")
	case <-ctx.Done():
		log.Info("got context cancelled")
	}

	log.Info("shutting both servers down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to graceful shutdown http server", slog.String("error", err.Error()))
	}

	container.GRPCServer.Stop()

	wg.Wait()
	log.Info("app down!")
}
