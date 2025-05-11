package tests

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"pills-taking-reminder/internal/infrastructure/postgres"
	"pills-taking-reminder/pkg/logger"
	"testing"
	"time"
)

var (
	testDB   *sql.DB
	testRepo *postgres.ScheduleRepository
)

func TestMain(m *testing.M) {
	var err error

	dbConfig := postgres.Config{
		Host:     "localhost",
		Port:     "5433",
		Username: "postgres",
		Password: "postgres",
		DBName:   "postgres",
	}

	logger := logger.SetupLogger("local")

	logger.Info("Connecting to test database",
		slog.String("host", dbConfig.Host),
		slog.String("port", dbConfig.Port),
		slog.String("database", dbConfig.DBName),
		slog.String("user", dbConfig.Username))

	testDB, err = postgres.NewConnection(dbConfig, logger)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	if err = testDB.Ping(); err != nil {
		fmt.Printf("Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Successfully connected to test database")

	cleanupDatabase()

	err = postgres.InitializeSchema(testDB, logger)
	if err != nil {
		fmt.Printf("Failed to initialize test schema: %v\n", err)
		os.Exit(1)
	}

	interval := 90 * time.Minute
	testRepo = postgres.NewScheduleRepository(testDB, logger, interval)

	exitCode := m.Run()

	cleanupDatabase()
	testDB.Close()

	os.Exit(exitCode)
}

// Cleanup function to reset the database between test runs
func cleanupDatabase() {
	// Clean up the data but keep the tables
	_, err := testDB.Exec("DELETE FROM takings")
	if err != nil {
		fmt.Printf("Failed to clean up takings: %v\n", err)
	}

	_, err = testDB.Exec("DELETE FROM schedules")
	if err != nil {
		fmt.Printf("Failed to clean up schedules: %v\n", err)
	}
}
