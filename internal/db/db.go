package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Things 3 reference date (2001-01-01 00:00:00 UTC) - Core Data epoch
var thingsReferenceDate = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

// dateToThingsDays converts a "YYYY-MM-DD" date to Things 3 days format
func dateToThingsDays(dateStr string) (int64, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, err
	}
	days := t.Sub(thingsReferenceDate).Hours() / 24
	return int64(days), nil
}

// ThingsDaysToDate converts Things 3 days to a time.Time
func ThingsDaysToDate(days int64) time.Time {
	return thingsReferenceDate.AddDate(0, 0, int(days))
}

// ThingsTimestampToTime converts Things 3 timestamp (seconds since 2001-01-01) to time.Time
func ThingsTimestampToTime(ts float64) time.Time {
	return thingsReferenceDate.Add(time.Duration(ts * float64(time.Second)))
}

// DefaultDBPath returns the default path to the Things 3 database
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Things 3 stores data in a versioned folder like "ThingsData-XXXXX"
	containerPath := filepath.Join(home, "Library", "Group Containers", "JLMPQHK86H.com.culturedcode.ThingsMac")

	// Find the ThingsData-* directory
	entries, err := os.ReadDir(containerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Things container: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 10 && entry.Name()[:10] == "ThingsData" {
			return filepath.Join(containerPath, entry.Name(), "Things Database.thingsdatabase", "main.sqlite"), nil
		}
	}

	// Fallback to old path structure (pre-versioned)
	return filepath.Join(containerPath, "Things Database.thingsdatabase", "main.sqlite"), nil
}

// TaskType represents the type of a task in Things 3
type TaskType int

const (
	TaskTypeTask    TaskType = 0
	TaskTypeProject TaskType = 1
	TaskTypeHeading TaskType = 2
)

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskStatusOpen      TaskStatus = 0
	TaskStatusCanceled  TaskStatus = 2
	TaskStatusCompleted TaskStatus = 3
)

// StartType represents the start classification of a task
type StartType int

const (
	StartTypeNotStarted StartType = 0
	StartTypeToday      StartType = 1
	StartTypeSomeday    StartType = 2
)

// Task represents a task, project, or heading from the TMTask table
type Task struct {
	UUID         string
	Title        string
	Notes        sql.NullString
	Type         TaskType
	Status       TaskStatus
	Start        StartType
	StartDate    sql.NullInt64
	Deadline     sql.NullInt64
	Project      sql.NullString
	Area         sql.NullString
	Heading      sql.NullString
	Trashed      bool
	CreationDate sql.NullFloat64
}

// Area represents an area from the TMArea table
type Area struct {
	UUID  string
	Title string
}

// Tag represents a tag from the TMTag table
type Tag struct {
	UUID  string
	Title string
}

// DB provides read-only access to the Things 3 database
type DB struct {
	conn *sql.DB
}

// Open connects to the Things 3 SQLite database in read-only mode
func (db *DB) Open(path string) error {
	// Open in read-only mode with immutable flag for safety
	dsn := fmt.Sprintf("file:%s?mode=ro&immutable=1", path)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection works
	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.conn = conn
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying database connection for executing queries
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// baseTaskQuery is the common SELECT clause for task queries
const baseTaskQuery = `
SELECT uuid, title, notes, type, status, start, startDate, deadline, project, area, heading, trashed, creationDate
FROM TMTask
`

// scanTask scans a row into a Task struct
func scanTask(row interface{ Scan(...interface{}) error }) (*Task, error) {
	var t Task
	var trashed int
	err := row.Scan(
		&t.UUID,
		&t.Title,
		&t.Notes,
		&t.Type,
		&t.Status,
		&t.Start,
		&t.StartDate,
		&t.Deadline,
		&t.Project,
		&t.Area,
		&t.Heading,
		&trashed,
		&t.CreationDate,
	)
	if err != nil {
		return nil, err
	}
	t.Trashed = trashed == 1
	return &t, nil
}

// scanTasks scans multiple rows into a slice of Tasks
func (db *DB) scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return tasks, nil
}

// todayStartDate calculates today's startDate value in Things 3 format.
// Things 3 encodes dates as: (year << 16) + (day_of_year + 32) * 128
func todayStartDate() int64 {
	now := time.Now()
	year := int64(now.Year())
	dayOfYear := int64(now.YearDay())
	return (year << 16) + (dayOfYear+32)*128
}

// GetToday returns all tasks scheduled for today
func (db *DB) GetToday() ([]Task, error) {
	todayValue := todayStartDate()
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type IN (0, 1)
  AND start = 1
  AND startDate = ?
ORDER BY todayIndex ASC
`
	rows, err := db.conn.Query(query, todayValue)
	if err != nil {
		return nil, fmt.Errorf("failed to query today tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetInbox returns all tasks in the inbox (no project, no area, not scheduled)
func (db *DB) GetInbox() ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND start = 0
  AND project IS NULL
  AND heading IS NULL
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbox tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetUpcoming returns all tasks with a scheduled start date
func (db *DB) GetUpcoming() ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND startDate IS NOT NULL
ORDER BY startDate ASC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetAnytime returns all tasks in the anytime list (not today, not someday)
func (db *DB) GetAnytime() ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND start = 0
  AND startDate IS NULL
  AND (project IS NOT NULL OR area IS NOT NULL OR heading IS NOT NULL)
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query anytime tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetSomeday returns all tasks in the someday list
func (db *DB) GetSomeday() ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND start = 2
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query someday tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetProjects returns all active projects
func (db *DB) GetProjects() ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 1
ORDER BY title ASC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetAreas returns all areas
func (db *DB) GetAreas() ([]Area, error) {
	query := `SELECT uuid, title FROM TMArea ORDER BY title ASC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query areas: %w", err)
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		var a Area
		if err := rows.Scan(&a.UUID, &a.Title); err != nil {
			return nil, fmt.Errorf("failed to scan area: %w", err)
		}
		areas = append(areas, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating areas: %w", err)
	}
	return areas, nil
}

// GetTags returns all tags
func (db *DB) GetTags() ([]Tag, error) {
	query := `SELECT uuid, title FROM TMTag ORDER BY title ASC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.UUID, &t.Title); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}
	return tags, nil
}

// GetTask returns a single task by UUID
func (db *DB) GetTask(uuid string) (*Task, error) {
	query := baseTaskQuery + `WHERE uuid = ?`
	row := db.conn.QueryRow(query, uuid)
	task, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", uuid)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// GetTasksInProject returns all tasks in a specific project
func (db *DB) GetTasksInProject(projectUUID string) ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND project = ?
ORDER BY todayIndex ASC, creationDate ASC
`
	rows, err := db.conn.Query(query, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetTasksInArea returns all tasks in a specific area
func (db *DB) GetTasksInArea(areaUUID string) ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND area = ?
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query, areaUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query area tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// Search searches for tasks by title
func (db *DB) Search(query string) ([]Task, error) {
	sqlQuery := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND title LIKE ?
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetTaskTags returns all tags for a specific task
func (db *DB) GetTaskTags(taskUUID string) ([]Tag, error) {
	query := `
SELECT t.uuid, t.title
FROM TMTag t
JOIN TMTaskTag tt ON t.uuid = tt.tags
WHERE tt.tasks = ?
ORDER BY t.title ASC
`
	rows, err := db.conn.Query(query, taskUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.UUID, &t.Title); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}
	return tags, nil
}

// GetAreaByUUID returns an area by its UUID
func (db *DB) GetAreaByUUID(uuid string) (*Area, error) {
	query := `SELECT uuid, title FROM TMArea WHERE uuid = ?`
	row := db.conn.QueryRow(query, uuid)
	var a Area
	if err := row.Scan(&a.UUID, &a.Title); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("area not found: %s", uuid)
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return &a, nil
}

// GetTasksByTag returns all open tasks with a specific tag
func (db *DB) GetTasksByTag(tagTitle string) ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 0
  AND trashed = 0
  AND type = 0
  AND uuid IN (
    SELECT tt.tasks
    FROM TMTaskTag tt
    JOIN TMTag t ON tt.tags = t.uuid
    WHERE t.title = ?
  )
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query, tagTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by tag: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetTagByTitle returns a tag by its title
func (db *DB) GetTagByTitle(title string) (*Tag, error) {
	query := `SELECT uuid, title FROM TMTag WHERE title = ?`
	row := db.conn.QueryRow(query, title)
	var t Tag
	if err := row.Scan(&t.UUID, &t.Title); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found: %s", title)
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return &t, nil
}

// GetLogbook returns completed tasks, optionally filtered by date range
// startDate and endDate should be in "YYYY-MM-DD" format, or empty for no filter
func (db *DB) GetLogbook(startDate, endDate string) ([]Task, error) {
	query := baseTaskQuery + `
WHERE status = 3
  AND trashed = 0
  AND type = 0
`
	var args []interface{}

	// Things 3 stores dates as days since 2001-01-01 (Core Data reference date)
	if startDate != "" {
		query += " AND stopDate >= ?"
		days, err := dateToThingsDays(startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date: %w", err)
		}
		args = append(args, days)
	}
	if endDate != "" {
		query += " AND stopDate <= ?"
		days, err := dateToThingsDays(endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}
		args = append(args, days)
	}

	query += " ORDER BY stopDate DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logbook: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// GetTrashed returns trashed tasks
func (db *DB) GetTrashed() ([]Task, error) {
	query := baseTaskQuery + `
WHERE trashed = 1
  AND type = 0
ORDER BY creationDate DESC
`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query trashed tasks: %w", err)
	}
	defer rows.Close()
	return db.scanTasks(rows)
}

// Stats holds aggregated task statistics
type Stats struct {
	Inbox        int `json:"inbox"`
	Today        int `json:"today"`
	Upcoming     int `json:"upcoming"`
	Anytime      int `json:"anytime"`
	Someday      int `json:"someday"`
	Completed    int `json:"completed"`
	Projects     int `json:"projects"`
	Areas        int `json:"areas"`
	Tags         int `json:"tags"`
}

// GetStats returns aggregated counts for all lists
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get counts using efficient queries
	todayValue := todayStartDate()
	queries := []struct {
		query string
		dest  *int
	}{
		{"SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type = 0 AND start = 0 AND project IS NULL AND heading IS NULL", &stats.Inbox},
		{fmt.Sprintf("SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type IN (0,1) AND start = 1 AND startDate = %d", todayValue), &stats.Today},
		{"SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type = 0 AND startDate IS NOT NULL", &stats.Upcoming},
		{"SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type = 0 AND start = 0 AND startDate IS NULL AND (project IS NOT NULL OR area IS NOT NULL OR heading IS NOT NULL)", &stats.Anytime},
		{"SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type = 0 AND start = 2", &stats.Someday},
		{"SELECT COUNT(*) FROM TMTask WHERE status = 3 AND trashed = 0 AND type = 0", &stats.Completed},
		{"SELECT COUNT(*) FROM TMTask WHERE status = 0 AND trashed = 0 AND type = 1", &stats.Projects},
		{"SELECT COUNT(*) FROM TMArea", &stats.Areas},
		{"SELECT COUNT(*) FROM TMTag", &stats.Tags},
	}

	for _, q := range queries {
		if err := db.conn.QueryRow(q.query).Scan(q.dest); err != nil {
			return nil, fmt.Errorf("failed to get count: %w", err)
		}
	}

	return stats, nil
}
