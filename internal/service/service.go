package service

import (
	"fmt"
	"pills-taking-reminder/internal/models"
	"pills-taking-reminder/internal/storage/pg"
)

type StorageService interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	//NextTakings(interval time.Duration, id int64) ([]models.Taking, error)
	//GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type Service struct {
	repo pg.StorageRepository
}

func NewService(repo pg.StorageRepository) StorageService {
	return &Service{repo: repo}
}

func (s *Service) CreateSchedule(schedule models.ScheduleRequest) (int64, error) {
	const operation = "service.CreateSchedule"

	scheduleID, err := s.repo.CreateSchedule(schedule)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return scheduleID, nil

}

func (s *Service) GetSchedulesIDs(userID int64) ([]int64, error) {
	const operation = "service.GetSchedulesIDs"

	schedulesIDs, err := s.repo.GetSchedulesIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return schedulesIDs, nil
}
