package service

import (
	"fmt"
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
		slog.Int("frequency", schedule.Frequency),
		slog.Int64("user_id", schedule.UserID),
	)

	scheduleID, err := s.repo.CreateSchedule(schedule)
	if err != nil {

		s.logger.Info("failed to create a schedule",
			slog.String("operation", operation),
			slog.String("medicine_name", schedule.MedicineName),
			slog.Int("duration", schedule.Duration),
			slog.Int("frequency", schedule.Frequency),
			slog.Int64("user_id", schedule.UserID),
			slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("schedule was successfully created!",
		slog.String("operation", operation),
		slog.String("medicine_name", schedule.MedicineName),
		slog.Int("duration", schedule.Duration),
		slog.Int("frequency", schedule.Frequency),
		slog.Int64("user_id", schedule.UserID),
	)

	return scheduleID, nil

}

func (s *Service) GetSchedulesIDs(userID int64) ([]int64, error) {
	const operation = "service.GetSchedulesIDs"

	s.logger.Info("getting ids of schedules by userID",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	schedulesIDs, err := s.repo.GetSchedulesIDs(userID)
	if err != nil {
		s.logger.Info("failed to get ids of schedules by userID",
			slog.String("operation", operation),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if len(schedulesIDs) == 0 {

		s.logger.Info("no schedules found for such userID",
			slog.String("operation", operation),
			slog.Int64("user_id", userID),
			slog.Any("schedules ids", schedulesIDs))
	} else {

		s.logger.Info("ids of schedules were successfully found!",
			slog.String("operation", operation),
			slog.Int64("user_id", userID),
			slog.Any("schedules ids", schedulesIDs))
	}
	return schedulesIDs, nil
}

func (s *Service) GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error) {
	const operation = "service.GetSchedule"

	s.logger.Info("getting schedule by userID and scheduleID",
		slog.String("operation", operation),
		slog.Int64("user_id", userID),
		slog.Int64("schedule_id", scheduleID))

	schedule, err := s.repo.GetSchedule(userID, scheduleID)
	if err != nil {
		s.logger.Info("failed to get schedule by userID and scheduleID",
			slog.String("operation", operation),
			slog.Int64("user_id", userID),
			slog.Int64("schedule_id", scheduleID),
			slog.String("error", err.Error()))

		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("schedule was successfuly found!",
		slog.String("operation", operation),
		slog.Int64("user_id", userID),
		slog.Int64("schedule_id", scheduleID),
		slog.Any("schedule", schedule))

	return schedule, nil
}

func (s *Service) GetNextTakings(userID int64) ([]models.Taking, error) {
	const operation = "service.GetNextTakings"

	s.logger.Info("getting next takings for user",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	nextTakings, err := s.repo.NextTakings(userID)
	if err != nil {

		s.logger.Info("failed to get next takings for user",
			slog.String("operation", operation),
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("next takings were successfuly found!",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	return nextTakings, nil
}
