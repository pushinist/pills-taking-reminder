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

	//schedules, err := db.GetSchedules(1)
	//if err != nil {
	//	log.Error("failed to get schedules", slog.String("error", err.Error()))
	//} else {
	//	log.Info("got schedules", "userid", 1)
	//	for _, schedule := range schedules {
	//		fmt.Println(schedule)
	//	}
	//}
	//
	//schedules, err = db.GetSchedules(2)
	//if err != nil {
	//	log.Error("failed to get schedules", slog.String("error", err.Error()))
	//} else {
	//	log.Info("got schedules", "userid", 2)
	//	for _, schedule := range schedules {
	//		fmt.Println(schedule)
	//	}
	//}

	//schedules, err := db.NextTakings(cfg.NearTakingInterval, 3)
	//if err != nil {
	//	log.Error("error getting storage", slog.String("error", err.Error()))
	//} else {
	//	log.Info("schedules for user", "userID", 1)
	//	for _, schedule := range schedules {
	//		log.Info("schedule", "schedule", schedule)
	//	}
	//}

	//fmt.Println(utils.CountTakings(models.Fourth))

	//schedule, err := db.GetSchedule(2, 3)
	//if err != nil {
	//	log.Error("failed to retrieve schedule", slog.String("error", err.Error()))
	//} else {
	//	fmt.Println(schedule)
	//}

}
