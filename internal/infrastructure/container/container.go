package container

import (
	"log/slog"
	"pills-taking-reminder/internal/api/grpc"
	httpHandler "pills-taking-reminder/internal/api/http"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/domain/repository"
	"pills-taking-reminder/internal/infrastructure/postgres"
	"pills-taking-reminder/internal/usecase"
	"pills-taking-reminder/pkg/logger"
)

type Container struct {
	Config          *config.Config
	Logger          *slog.Logger
	ScheduleUseCase *usecase.ScheduleUseCase
	HTTPHandler     *httpHandler.ScheduleHandler
	GRPCServer      *grpc.GRPCServer
}

func New(cfg *config.Config) (*Container, error) {
	log := logger.SetupLogger(cfg.Env)
	log.Info("initializing app",
		slog.String("env", cfg.Env),
		slog.String("http_server", cfg.HTTPServer.Address),
		slog.String("grpc_server", cfg.GRPCServer.Address))

	dbConfig := postgres.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Username: cfg.DB.Username,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
	}
	db, err := postgres.NewConnection(dbConfig, log)
	if err != nil {
		return nil, err
	}

	if err := postgres.InitializeSchema(db, log); err != nil {
		return nil, err
	}

	var scheduleRepo repository.ScheduleRepository
	scheduleRepo = postgres.NewScheduleRepository(db, log, cfg.NearTakingInterval)

	scheduleUseCase := usecase.NewScheduleUseCase(scheduleRepo, cfg.NearTakingInterval)

	httpHandler := httpHandler.NewScheduleHandler(scheduleUseCase, log)

	grpcServer := grpc.NewGRPCServer(scheduleUseCase, log)

	return &Container{
		Config:          cfg,
		Logger:          log,
		ScheduleUseCase: scheduleUseCase,
		HTTPHandler:     httpHandler,
		GRPCServer:      grpcServer,
	}, nil

}
