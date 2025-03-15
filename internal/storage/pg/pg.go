package pg

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/models"
	"time"
)

type StorageRepository interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	NextTakings(interval time.Duration, id int64) ([]models.Taking, error)
	GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type Storage struct {
	db *sql.DB
}

func New(dbCfg config.DB) (StorageRepository, error) {
	const operation = "storage.pg.new"

	DSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbCfg.Username,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.Name)
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	err = createTable(db, createSchedulesQuery, operation)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = createTable(db, createTakingsQuery, operation)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateSchedule(schedule models.ScheduleRequest) (int64, error) {
	const operation = "storage.pg.CreateSchedule"
	if schedule.Duration == 0 {
		id, err := createInfinitySchedule(s, schedule, operation)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", operation, err)
		}
		return id, nil
	} else {
		id, err := createEndingSchedule(s, schedule, operation)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", operation, err)
		}
		return id, nil
	}
}

func (s *Storage) GetSchedulesIDs(userID int64) ([]int64, error) {
	const operation = "storage.pg.GetSchedulesIDs"
	stmt, err := s.db.Prepare(getSchedulesQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var schedulesIDs []int64
	for rows.Next() {
		var scheduleID int64

		err := rows.Scan(&scheduleID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		schedulesIDs = append(schedulesIDs, scheduleID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return schedulesIDs, nil
}

func (s *Storage) NextTakings(interval time.Duration, id int64) ([]models.Taking, error) {
	const operation = "storage.pg.NextTakings"

	stmt, err := s.db.Prepare(getNextTakingsQuery)
	if err != nil {
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(time.Now().Format("15:04"), time.Now().Add(interval).Format("15:04"), id)
	if err != nil {
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var takings []models.Taking
	for rows.Next() {
		var taking models.Taking
		err = rows.Scan(&taking.MedicineName, &taking.TakingTime)
		if err != nil {
			return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
		}
		takings = append(takings, taking)
	}

	if err := rows.Err(); err != nil {
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}

	return takings, nil
}

func (s *Storage) GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error) {
	const operation = "storage.pg.GetSchedule"

	stmt, err := s.db.Prepare(getScheduleQuery)
	if err != nil {
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, scheduleID)
	if err != nil {
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var schedule models.ScheduleResponse
	for rows.Next() {
		var rawTakingTime time.Time
		err := rows.Scan(&schedule.MedicineName, &schedule.StartDate, &schedule.EndDate, &schedule.UserID, &rawTakingTime)
		if err != nil {
			return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
		}

		takingTime := rawTakingTime.Format("15:04")

		schedule.TakingTime = append(schedule.TakingTime, takingTime)
	}

	if err := rows.Err(); err != nil {
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}

	return schedule, nil

}
