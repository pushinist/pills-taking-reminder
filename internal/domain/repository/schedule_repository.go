package repository

import (
	"context"
	"errors"
	"pills-taking-reminder/internal/domain/entities"
)

var (
	ErrAlreadyExists = errors.New("schedule already exists")
	ErrNotFound      = errors.New("schedule was not found")
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *entities.Schedule) (int64, error)
	GetByID(ctx context.Context, userID, scheduleID int64) (*entities.Schedule, error)
	GetSchedulesIDs(ctx context.Context, userID int64) ([]int64, error)
	GetNextTakings(ctx context.Context, userID int64, interval string) ([]entities.Taking, error)
}
