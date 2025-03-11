package main

import (
	"log/slog"
	"os"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/models"
	"pills-taking-reminder/internal/storage/pg"
	"pills-taking-reminder/pkg/logger"
	"time"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("logger initialized, starting pills-taking-reminder", slog.String("env", cfg.Env))

	db, err := pg.New(cfg.DB)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("storage initialized")

	schedule := models.Schedule{
		MedicineName: "AAAA",
		Frequency:    models.Once,
		StartDate:    time.Now(),
		UserID:       1,
	}

	id, err := db.CreateSchedule(schedule)
	if err != nil {
		log.Error("failed to create schedule", slog.String("error", err.Error()))
	} else {
		log.Info("created schedule", "id", id)
	}

	//schedules, err := db.NextTakings(cfg.NearTakingInterval)
	//if err != nil {
	//	log.Error("error getting storage", slog.String("error", err.Error()))
	//} else {
	//	log.Info("schedules for user", "userID", 1, "schedules", schedules)
	//}
	//
	//for _, schedule := range schedules {
	//	log.Info("scheduling user", "userID", 1, "schedule", schedule)
	//}

	//fmt.Println(utils.CountTakings(models.Fourth))
}
