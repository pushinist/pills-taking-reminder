package service

import (
	"fmt"
	"log"
	"log/slog"
	"pills-taking-reminder/internal/models"
	"pills-taking-reminder/internal/server"
	"pills-taking-reminder/internal/storage/pg"
	"time"
)

type Service struct {
	repo     pg.StorageRepository
	logger   *slog.Logger
	interval time.Duration
}

func NewService(repo pg.StorageRepository, logger *slog.Logger) server.StorageService {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateSchedule(schedule models.ScheduleRequest) (int64, error) {
	const operation = "service.CreateSchedule"

	s.logger.Info("creating a schedule",
		slog.String("operation", operation),
		slog.String("medicine_name", schedule.MedicineName),
		slog.Int("duration", schedule.Duration),
		slog.Int("frequency", schedule.Frequency))

	scheduleID, err := s.repo.CreateSchedule(schedule)
	if err != nil {

		s.logger.Info("failed to create a schedule",
			slog.String("operation", operation),
			slog.String("medicine_name", schedule.MedicineName),
			slog.Int("duration", schedule.Duration),
			slog.Int("frequency", schedule.Frequency),
			slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("schedule was successfully created!",
		slog.String("operation", operation),
		slog.String("medicine_name", schedule.MedicineName),
		slog.Int("duration", schedule.Duration),
		slog.Int("frequency", schedule.Frequency))

	return scheduleID, nil

}

func (s *Service) GetSchedulesIDs(userID int64) ([]int64, error) {
	const operation = "service.GetSchedulesIDs"

	s.logger.Info("getting schedules by userID",
		slog.String("operation", operation))

	schedulesIDs, err := s.repo.GetSchedulesIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return schedulesIDs, nil
}

func (s *Service) GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error) {
	const operation = "service.GetSchedule"

	schedule, err := s.repo.GetSchedule(userID, scheduleID)
	if err != nil {
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}
	log.Print(schedule)
	return schedule, nil
}

func (s *Service) GetNextTakings(id int64) ([]models.Taking, error) {
	const operation = "service.GetNextTakings"

	nextTakings, err := s.repo.NextTakings(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return nextTakings, nil
}
