package usecase

import (
	"context"
	"errors"
	"fmt"
	"pills-taking-reminder/internal/domain/entities"
	"pills-taking-reminder/internal/domain/repository"
	"time"
)

var (
	ErrInvalidInput     = errors.New("invalid input parameters")
	ErrScheduleNotFound = errors.New("schedule was not found")
	ErrScheduleExists   = errors.New("schedule already exists")
	ErrDatabaseError    = errors.New("database error")
)

type ScheduleInput struct {
	MedicineName string
	Frequency    int
	Duration     int
	UserID       int64
}

type ScheduleOutput struct {
	ID           int64
	MedicineName string
	StartDate    string
	EndDate      string
	UserID       int64
	TakingTimes  []string
}

type TakingOutput struct {
	MedicineName string
	TakingTime   string
}

type ScheduleUseCase struct {
	scheduleRepo repository.ScheduleRepository
	interval     time.Duration
}

func NewScheduleUseCase(scheduleRepo repository.ScheduleRepository, interval time.Duration) *ScheduleUseCase {
	return &ScheduleUseCase{
		scheduleRepo: scheduleRepo,
		interval:     interval,
	}
}

func (uc *ScheduleUseCase) CreateSchedule(ctx context.Context, input ScheduleInput) (int64, error) {
	if input.MedicineName == "" || input.Frequency < 1 || input.Duration < 0 || input.UserID <= 0 {
		return 0, ErrInvalidInput
	}

	schedule, err := entities.NewSchedule(input.MedicineName, input.Frequency, input.Duration, input.UserID)
	if err != nil {
		return 0, err
	}

	id, err := uc.scheduleRepo.Create(ctx, schedule)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return 0, ErrScheduleExists
		}
		return 0, fmt.Errorf("failed to create a schedule: %w", err)
	}

	return id, nil
}

func (uc *ScheduleUseCase) GetSchedule(ctx context.Context, userID, scheduleID int64) (*ScheduleOutput, error) {
	if userID <= 0 || scheduleID <= 0 {
		return nil, ErrInvalidInput
	}

	schedule, err := uc.scheduleRepo.GetByID(ctx, userID, scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	output := &ScheduleOutput{
		ID:           schedule.ID,
		MedicineName: schedule.MedicineName,
		StartDate:    schedule.StartDate.Format("02 Jan 2006"),
		UserID:       schedule.UserID,
		TakingTimes:  make([]string, len(schedule.TakingTimes)),
	}

	if schedule.EndDate != nil {
		output.EndDate = schedule.EndDate.Format("02 Jan 2006")
	} else {
		output.EndDate = "infinite"
	}

	for i, tt := range schedule.TakingTimes {
		output.TakingTimes[i] = fmt.Sprintf("%02d:%02d", tt.Time.Hour(), tt.Time.Minute())
	}

	return output, nil
}

func (uc *ScheduleUseCase) GetScheduleIDs(ctx context.Context, userID int64) ([]int64, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}

	ids, err := uc.scheduleRepo.GetSchedulesIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule IDs: %w", err)
	}

	return ids, nil
}

func (uc *ScheduleUseCase) GetNextTakings(ctx context.Context, userID int64) ([]TakingOutput, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}

	takings, err := uc.scheduleRepo.GetNextTakings(ctx, userID, uc.interval.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get next takings: %w", err)
	}

	output := make([]TakingOutput, len(takings))
	for i, taking := range takings {
		output[i] = TakingOutput{
			MedicineName: taking.MedicineName,
			TakingTime:   taking.FormatTime(),
		}
	}

	return output, nil
}
