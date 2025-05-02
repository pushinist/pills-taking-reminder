package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func NewConnection(cfg Config, logger *slog.Logger) (*sql.DB, error) {

	const operation = "postgres.NewConnection"

	logger.Info("initing db connection...",
		slog.String("operation", operation),
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("dbname", cfg.DBName))

	DSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName)
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		logger.Error("failed to connect to db",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		logger.Error("failed to ping db",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	logger.Info("successfully connected to db", slog.String("operation", operation))
	return db, nil
}

func InitializeSchema(db *sql.DB, logger *slog.Logger) error {
	const operation = "postgres.InitializeSchema"

	logger.Info("initializing db schema", slog.String("operation", operation))

	_, err := db.Exec(createSchedulesQuery)
	if err != nil {
		logger.Error("failed to create schedules table",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", operation, err)
	}

	_, err = db.Exec(createTakingsQuery)
	if err != nil {
		logger.Error("failed to create takings table",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", operation, err)
	}

	logger.Info("db schema initialized successfully", slog.String("operation", operation))
	return nil

}
