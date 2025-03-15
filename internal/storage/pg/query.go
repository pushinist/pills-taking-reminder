package pg

const (
	createSchedulesQuery = `
	CREATE TABLE IF NOT EXISTS schedules(
	    id SERIAL PRIMARY KEY,
	    medicine_name TEXT,
	    start_date DATE NOT NULL,
	    end_date DATE,
	    user_id INTEGER	    
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
		SELECT schedules.medicine_name, takings.taking_time FROM takings
		JOIN schedules ON takings.schedule_id = schedules.id
		WHERE (takings.taking_time BETWEEN $1 AND $2) AND (schedules.user_id = $3)
		`

	getScheduleQuery = `
		SELECT schedules.medicine_name, schedules.start_date, schedules.end_date, schedules.user_id, takings.taking_time FROM schedules
		JOIN takings ON schedules.id = takings.schedule_id
		WHERE (schedules.user_id = $1) AND (schedules.id = $2)
		`

	getSchedulesQuery = `
		SELECT id FROM schedules
		WHERE user_id = $1 AND (end_date > $2 or end_date IS NULL)
		`
)
