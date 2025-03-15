package service

import (
	"fmt"
	"log"
	"pills-taking-reminder/internal/models"
	"pills-taking-reminder/internal/storage/pg"
	"time"
)

type StorageService interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	GetNextTakings(id int64) ([]models.Taking, error)
	GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type Service struct {
	repo     pg.StorageRepository
	interval time.Duration
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
