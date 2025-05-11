package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"pills-taking-reminder/internal/domain/entities"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrAlreadyExists = errors.New("schedule already exists")
	ErrNotFound      = errors.New("schedule was not found")
)

type ScheduleRepository struct {
	db       *sql.DB
	logger   *slog.Logger
	interval time.Duration
}

func NewScheduleRepository(db *sql.DB, logger *slog.Logger, interval time.Duration) *ScheduleRepository {
	return &ScheduleRepository{
		db:       db,
		logger:   logger,
		interval: interval,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *entities.Schedule) (int64, error) {
	const operation = "postgres.ScheduleRepository.Create"

	r.logger.Info("creating a schedule in db",
		slog.String("operation", operation),
		slog.Any("schedule", schedule))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback()

	var id int64
	var query string
	var args []any

	if schedule.EndDate == nil {
		query = addInfiniteScheduleQuery
		args = []any{schedule.MedicineName, schedule.StartDate.Format("2006-01-02"), schedule.UserID}
	} else {
		query = addTemporaryScheduleQuery
		args = []any{schedule.MedicineName, schedule.StartDate.Format("2006-01-02"), schedule.EndDate.Format("2006-01-02"), schedule.UserID}
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		if isPgUniqueViolation(err) {
			r.logger.Info("schedule already exists", slog.String("operation", operation))
			return 0, ErrAlreadyExists
		}
		r.logger.Error("failed to insert schedule",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	for _, tt := range schedule.TakingTimes {
		takingTime := fmt.Sprintf("%02d:%02d", tt.Time.Hour(), tt.Time.Minute())
		_, err = tx.ExecContext(ctx,
			addTakingTimeQuery, id, takingTime)
		if err != nil {
			r.logger.Error("failed to insert taking time",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return 0, fmt.Errorf("%s: %w", operation, err)
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	r.logger.Info("schedule was created successfully",
		slog.String("operation", operation),
		slog.Int64("id", id))

	return id, nil
}

func (r *ScheduleRepository) GetSchedulesIDs(ctx context.Context, userID int64) ([]int64, error) {
	const operation = "postgres.ScheduleRepository.GetScheduleIDs"

	r.logger.Info("getting schedule IDs for user",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	rows, err := r.db.QueryContext(ctx, getSchedulesQuery, userID, time.Now().Format("2006-01-02"))
	if err != nil {
		r.logger.Error("failed to get schedule IDs",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("failed to scan row",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error in rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return ids, nil
}

func (r *ScheduleRepository) GetNextTakings(ctx context.Context, userID int64, interval string) ([]entities.Taking, error) {
	const operation = "postgres.ScheduleRepository.GetNextTakings"

	r.logger.Info("getting next takings for user",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	now := TimeNow()
	currentTime := now.Format("15:04")

	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		r.logger.Error("failed to parse interval",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	endTime := now.Add(intervalDuration).Format("15:04")

	rows, err := r.db.QueryContext(ctx, getNextTakingsQuery, userID, currentTime, endTime, now.Format("2006-01-02"))
	if err != nil {
		r.logger.Error("failed to get next takings",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var takings []entities.Taking
	for rows.Next() {
		var medicineName string
		var takingTimeStr string

		if err := rows.Scan(&medicineName, &takingTimeStr); err != nil {
			r.logger.Error("failed to scan row",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		tt, err := time.Parse("15:04", takingTimeStr)
		if err != nil {
			r.logger.Error("failed to parse taking time",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		takingTime := time.Date(now.Year(), now.Month(), now.Day(), tt.Hour(), tt.Minute(), 0, 0, now.Location())

		takings = append(takings, entities.Taking{
			MedicineName: medicineName,
			TakingTime:   takingTime,
		})

	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error in rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return takings, nil
}

func (r *ScheduleRepository) GetByID(ctx context.Context, userID, scheduleID int64) (*entities.Schedule, error) {
	const operation = "postgres.ScheduleRepository.GetByID"

	r.logger.Info("gettings schedule by ID",
		slog.String("operation", operation),
		slog.Int64("user_id", userID),
		slog.Int64("schedule_id", scheduleID))

	rows, err := r.db.QueryContext(ctx, getScheduleQuery, userID, scheduleID)
	if err != nil {
		r.logger.Error("failed to query schedule",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var schedule *entities.Schedule
	var takingTimes []entities.TakingTime
	var count int

	for rows.Next() {
		if schedule == nil {
			schedule = &entities.Schedule{
				ID:          scheduleID,
				UserID:      userID,
				TakingTimes: []entities.TakingTime{},
			}
		}

		var id int64
		var medicineName string
		var startDate time.Time
		var endDate sql.NullTime
		var userId int64
		var takingTime time.Time

		if err := rows.Scan(&id, &medicineName, &startDate, &endDate, &userId, &takingTime); err != nil {
			r.logger.Error("failed to scan row",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
		}

		if count == 0 {
			schedule.MedicineName = medicineName
			schedule.StartDate = startDate
			if endDate.Valid {
				schedule.EndDate = &endDate.Time
			}
		}

		takingTimes = append(takingTimes, entities.TakingTime{
			Time: time.Date(0, 0, 0, takingTime.Hour(), takingTime.Minute(), 0, 0, time.UTC),
		})

		count++
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error in rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if schedule == nil {
		r.logger.Info("schedule was not found", slog.String("operation", operation))
		return nil, ErrNotFound
	}

	schedule.TakingTimes = takingTimes
	return schedule, nil
}

func isPgUniqueViolation(err error) bool {
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint" ||
		err.Error() == "ERROR: duplicate key value violates unique constraint")
}
