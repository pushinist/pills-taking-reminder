package main

import (
	"fmt"
	"log/slog"
	"os"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/storage/pg"
	"pills-taking-reminder/pkg/logger"
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

	//scheduleOnce := models.ScheduleRequest{
	//	MedicineName: "AAAA",
	//	Frequency:    models.Once,
	//	Duration:     3,
	//	UserID:       1,
	//}
	//
	//scheduleHourly := models.ScheduleRequest{
	//	MedicineName: "something1",
	//	Frequency:    models.Hourly,
	//	Duration:     0,
	//	UserID:       1,
	//}
	//secondSchedule := models.ScheduleRequest{
	//	MedicineName: "something2",
	//	Frequency:    models.Hourly,
	//	Duration:     2,
	//	UserID:       2,
	//}
	//
	//id, err := db.CreateSchedule(scheduleOnce)
	//if err != nil {
	//	log.Error("failed to create schedule", slog.String("error", err.Error()))
	//} else {
	//	log.Info("created schedule", "id", id)
	//}
	//
	//id, err = db.CreateSchedule(scheduleHourly)
	//if err != nil {
	//	log.Error("failed to create schedule", slog.String("error", err.Error()))
	//} else {
	//	log.Info("created schedule", "id", id)
	//}
	//
	//id, err = db.CreateSchedule(secondSchedule)
	//if err != nil {
	//	log.Error("failed to create schedule", slog.String("error", err.Error()))
	//} else {
	//	log.Info("created schedule", "id", id)
	//}

	schedules, err := db.GetSchedules(1)
	if err != nil {
		log.Error("failed to get schedules", slog.String("error", err.Error()))
	} else {
		log.Info("got schedules", "userid", 1)
		for _, schedule := range schedules {
			fmt.Println(schedule)
		}
	}

	schedules, err = db.GetSchedules(2)
	if err != nil {
		log.Error("failed to get schedules", slog.String("error", err.Error()))
	} else {
		log.Info("got schedules", "userid", 2)
		for _, schedule := range schedules {
			fmt.Println(schedule)
		}
	}

	//schedules, err := db.NextTakings(cfg.NearTakingInterval, 2)
	//if err != nil {
	//	log.Error("error getting storage", slog.String("error", err.Error()))
	//} else {
	//	log.Info("schedules for user", "userID", 2, "schedules", schedules)
	//}
	//
	//for _, schedule := range schedules {
	//	log.Info("schedule user", "userID", 2, "schedule", schedule)
	//}

	//fmt.Println(utils.CountTakings(models.Fourth))

	//schedule, takingTimes, err := db.GetSchedule(1, 2)
	//if err != nil {
	//	log.Error("failed to retrieve schedule", slog.String("error", err.Error()))
	//}
	//fmt.Println(schedule)
	//fmt.Println(takingTimes)
}
