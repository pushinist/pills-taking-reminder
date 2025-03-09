package pg

import (
	"database/sql"
	"fmt"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func createTable(db *sql.DB, statement, operation string) error {
	stmt, err := db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}

func New(dbCfg config.DB) (*Storage, error) {
	const operation = "storage.pg.new"

	const createSchedulesStatement = `
	CREATE TABLE IF NOT EXISTS schedules(
	    id SERIAL PRIMARY KEY,
	    medicine_name TEXT,
	    frequency TEXT,
	    start_date TIMESTAMP NOT NULL,
	    end_date TIMESTAMP,
	    user_id INTEGER	    
	)`

	const createTakingsStatement = `
CREATE TABLE IF NOT EXISTS takings(
	    id SERIAL PRIMARY KEY,
	    schedule_id INTEGER,
	    taking_time TIMESTAMP NOT NULL,
	    FOREIGN KEY(schedule_id) REFERENCES schedules(id)
	)`

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

	err = createTable(db, createSchedulesStatement, operation)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = createTable(db, createTakingsStatement, operation)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateSchedule(schedule models.Schedule) (int64, error) {
	const operation = "storage.pg.createSchedule"
	stmt, err := s.db.Prepare(`
		INSERT INTO schedules(medicine_name, frequency, start_date, end_date, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	var id int64
	err = stmt.QueryRow(schedule.MedicineName, schedule.Frequency, schedule.StartDate, schedule.EndDate, schedule.UserID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}
