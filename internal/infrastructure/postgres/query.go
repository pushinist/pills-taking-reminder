package postgres

const (
	createSchedulesQuery = `
	CREATE TABLE IF NOT EXISTS schedules(
	    id SERIAL PRIMARY KEY,
	    medicine_name TEXT,
	    start_date DATE NOT NULL,
	    end_date DATE,
	    user_id INTEGER,
	    UNIQUE(medicine_name, user_id)
	)`

	createTakingsQuery = `
CREATE TABLE IF NOT EXISTS takings(
	    id SERIAL PRIMARY KEY,
	    schedule_id INTEGER NOT NULL,
	    taking_time TIME NOT NULL,
	    FOREIGN KEY(schedule_id) REFERENCES schedules(id)
	)`

	addInfiniteScheduleQuery = `
		INSERT INTO schedules(medicine_name, start_date, user_id)
		VALUES ($1, $2, $3)
		RETURNING id
		`

	addTemporaryScheduleQuery = `
		INSERT INTO schedules(medicine_name, start_date, end_date, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`

	addTakingTimeQuery = `
INSERT INTO takings(schedule_id, taking_time)
VALUES ($1, $2)`

	getNextTakingsQuery = `
		SELECT s.medicine_name, TO_CHAR(t.taking_time, 'HH24:MI') AS time
		FROM takings t
		JOIN schedules s ON t.schedule_id = s.id
		WHERE s.user_id = $1 
		  AND t.taking_time::time BETWEEN $2::time AND $3::time
		  AND (s.end_date > $4 OR s.end_date IS NULL)
		ORDER BY t.taking_time
	`

	getScheduleQuery = `
		SELECT s.id, s.medicine_name, s.start_date, s.end_date, s.user_id, t.taking_time 
		FROM schedules s
		JOIN takings t ON s.id = t.schedule_id
		WHERE s.user_id = $1 AND s.id = $2
	`

	getSchedulesQuery = `
		SELECT id FROM schedules
		WHERE user_id = $1 AND (end_date > $2 or end_date IS NULL)
		`
)
