package pg

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/models"
	"time"
)

type Storage struct {
	db *sql.DB
}

func New(dbCfg config.DB) (*Storage, error) {
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

func (s *Storage) GetSchedules(userID int64) ([]models.ScheduleResponse, error) {
	const operation = "storage.pg.GetSchedules"
	stmt, err := s.db.Prepare(getSchedulesQuery)
	if err != nil {
		return []models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}

	rows, err := stmt.Query(userID)
	if err != nil {
		return []models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}

	var schedules []models.ScheduleResponse
	for rows.Next() {
		var schedule models.ScheduleResponse

		err := rows.Scan(&schedule.ID, &schedule.MedicineName, &schedule.StartDate, &schedule.EndDate, &schedule.UserID)
		if err != nil {
			return []models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (s *Storage) NextTakings(interval time.Duration, id int64) ([]models.Taking, error) {
	const operation = "storage.pg.NextTakings"

	stmt, err := s.db.Prepare(getNextTakingsQuery)
	if err != nil {
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	rows, err := stmt.Query(time.Now().Format("15:04"), time.Now().Add(interval).Format("15:04"), id)
	if err != nil {
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	var takings []models.Taking
	for rows.Next() {
		var taking models.Taking
		err = rows.Scan(&taking.MedicineName, &taking.TakingTime)
		if err != nil {
			return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
		}
		takings = append(takings, taking)
	}

	return takings, nil
}

func (s *Storage) GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, []string, error) {
	const operation = "storage.pg.GetSchedule"

	stmt, err := s.db.Prepare(getScheduleQuery)
	if err != nil {
		return models.ScheduleResponse{}, []string{}, fmt.Errorf("%s: %w", operation, err)
	}

	rows, err := stmt.Query(userID, scheduleID)
	if err != nil {
		return models.ScheduleResponse{}, []string{}, fmt.Errorf("%s: %w", operation, err)
	}

	var schedule models.ScheduleResponse
	var takingTimes []string
	for rows.Next() {
		var rawTakingTime time.Time
		err := rows.Scan(&schedule.MedicineName, &schedule.StartDate, &schedule.EndDate, &schedule.UserID, &rawTakingTime)
		if err != nil {
			return models.ScheduleResponse{}, []string{}, fmt.Errorf("%s: %w", operation, err)
		}
		if schedule.EndDate.Time.IsZero() {

		}
		takingTime := rawTakingTime.Format("15:04")

		takingTimes = append(takingTimes, takingTime)
	}

	return schedule, takingTimes, nil

}
