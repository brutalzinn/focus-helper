package database

import (
	"database/sql"
	"fmt"
	"focus-helper/src/pkg/models"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func Init(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createStatements := `
	CREATE TABLE IF NOT EXISTS wellbeing_checks (
		id INTEGER PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		question TEXT NOT NULL,
		answer TEXT
	);
	CREATE TABLE IF NOT EXISTS hyperfocus_sessions (
		id INTEGER PRIMARY KEY,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		duration_seconds INTEGER NOT NULL,
		subject TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS hyperfocus_events (
		id INTEGER PRIMARY KEY,
		alert_level INTEGER NOT NULL,
		subject TEXT NOT NULL,
		duration_seconds INTEGER NOT NULL,
		createAt DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS mayday_events (
		id INTEGER PRIMARY KEY,
		timestamp DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		subject TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NULL,
		is_active BOOLEAN DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS current_session (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions (id)
	);`

	if _, err := db.Exec(createStatements); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully.")
	return db, nil
}

func LogWellbeingCheck(db *sql.DB, question, answer string) {
	stmt, err := db.Prepare("INSERT INTO wellbeing_checks(timestamp, question, answer) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement for wellbeing check: %v", err)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(time.Now(), question, answer); err != nil {
		log.Printf("Error inserting wellbeing check: %v", err)
	}
}

func LogHyperfocusSession(db *sql.DB, startTime, endTime time.Time, subject string) {
	duration := int(endTime.Sub(startTime).Seconds())
	log.Printf("Hyperfocus session logged: Subject='%s', Duration=%d seconds", subject, duration)
	stmt, err := db.Prepare("INSERT INTO hyperfocus_sessions(start_time, end_time, duration_seconds, subject) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement for hyperfocus session: %v", err)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(startTime, endTime, duration, subject); err != nil {
		log.Printf("Error inserting hyperfocus session: %v", err)
	}
}

func LogHyperfocusEvent(db *sql.DB, level int, startTime, endTime time.Time, subject string) {
	log.Printf("[DB]Insert Hyperfocus Event Level %d logged for subject: '%s'", level, subject)
	duration := int(endTime.Sub(startTime).Seconds())
	stmt, err := db.Prepare("INSERT INTO hyperfocus_events(alert_level, subject, createAt, duration_seconds) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement for hyperfocus event: %v", err)
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(level, subject, time.Now(), duration); err != nil {
		log.Printf("Error inserting hyperfocus event: %v", err)
	}
}

func LogMaydayEvent(db *sql.DB) {
	log.Println("!!! MAYDAY event logged in the database !!!")
	stmt, err := db.Prepare("INSERT INTO mayday_events(timestamp) VALUES(?)")
	if err != nil {
		log.Printf("Error preparing statement for Mayday event: %v", err)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(time.Now()); err != nil {
		log.Printf("Error inserting Mayday event: %v", err)
	}
}

func GetRecentHistorySummary(db *sql.DB) (string, error) {
	var summary strings.Builder
	var lastSubject string
	err := db.QueryRow("SELECT COALESCE(subject, 'Unknown') FROM hyperfocus_sessions ORDER BY end_time DESC LIMIT 1").Scan(&lastSubject)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if lastSubject != "" && lastSubject != "Unknown" {
		summary.WriteString(fmt.Sprintf("The user's last focused activity was '%s'. ", lastSubject))
	}
	var maxLevel sql.NullInt64
	err = db.QueryRow("SELECT MAX(alert_level) FROM hyperfocus_events WHERE createAt > ?", time.Now().Add(-1*time.Hour)).Scan(&maxLevel)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if maxLevel.Valid && maxLevel.Int64 > 0 {
		summary.WriteString(fmt.Sprintf("They have already reached Alert Level %d in the last hour. ", maxLevel.Int64))
	}
	if summary.Len() == 0 {
		return "", nil
	}
	return summary.String(), nil
}

func GetCurrentSession(db *sql.DB) (*models.Session, error) {
	var session models.Session
	var endTime sql.NullTime

	err := db.QueryRow(`
		SELECT id, subject, start_time, end_time, is_active 
		FROM sessions 
		WHERE is_active = 1 
		ORDER BY start_time DESC 
		LIMIT 1
	`).Scan(&session.ID, &session.Subject, &session.StartTime, &endTime, &session.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}
	if endTime.Valid {
		session.EndTime = endTime.Time
	}
	return &session, nil
}

func EndSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`
		UPDATE sessions 
		SET end_time = ?, is_active = 0 
		WHERE id = ?
	`, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}
	log.Printf("Session %s closed", sessionID)
	return nil
}

func CreateSession(db *sql.DB, subject string) (string, error) {
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	_, err := db.Exec(`
		INSERT INTO sessions (id, subject, start_time, is_active) 
		VALUES (?, ?, ?, 1)
	`, sessionID, subject, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	log.Printf("Created new session: %s for subject: %s", sessionID, subject)
	return sessionID, nil
}

func GetCurrentSessionWithTimeout(db *sql.DB, timeout time.Duration) (*models.Session, error) {
	var session models.Session
	var endTime sql.NullTime

	timeoutHours := int(timeout.Hours())
	if timeoutHours < 1 {
		timeoutHours = 1 // Minimum 1 hour
	}

	err := db.QueryRow(`
		SELECT id, subject, start_time, end_time, is_active 
		FROM sessions 
		WHERE is_active = 1 
		AND start_time > datetime('now', '-`+fmt.Sprintf("%d", timeoutHours)+` hours')
		ORDER BY start_time DESC 
		LIMIT 1
	`).Scan(&session.ID, &session.Subject, &session.StartTime, &endTime, &session.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}
	if endTime.Valid {
		session.EndTime = endTime.Time
	}
	return &session, nil
}

func UpdateSessionSubject(db *sql.DB, sessionID, subject string) error {
	_, err := db.Exec(`
		UPDATE sessions 
		SET subject = ? 
		WHERE id = ? AND is_active = 1
	`, subject, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session subject: %w", err)
	}
	log.Printf("Updated session %s subject to: %s", sessionID, subject)
	return nil
}

func CleanupOldSessions(db *sql.DB, timeout time.Duration) error {

	timeoutHours := int(timeout.Hours())
	if timeoutHours < 1 {
		timeoutHours = 1 // Minimum 1 hour
	}

	_, err := db.Exec(`
		UPDATE sessions 
		SET is_active = 0, end_time = datetime('now')
		WHERE is_active = 1 
		AND start_time <= datetime('now', '-` + fmt.Sprintf("%d", timeoutHours) + ` hours')
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup old sessions: %w", err)
	}
	log.Printf("Cleaned up old sessions (older than %v)", timeout)
	return nil
}

func GetSessions(db *sql.DB, limit, offset int) ([]models.Session, error) {
	query := `
		SELECT id, subject, start_time, end_time, is_active 
		FROM sessions 
		ORDER BY start_time DESC 
		LIMIT ? OFFSET ?
	`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		var endTime sql.NullTime
		err := rows.Scan(&session.ID, &session.Subject, &session.StartTime, &endTime, &session.IsActive)
		if err != nil {
			return nil, err
		}
		if endTime.Valid {
			session.EndTime = endTime.Time
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func SaveWellbeingResponse(db *sql.DB, questionID, response, sessionID string) error {
	query := `
		INSERT INTO wellbeing_checks (timestamp, question, answer, session_id) 
		VALUES (datetime('now'), ?, ?, ?)
	`
	_, err := db.Exec(query, questionID, response, sessionID)
	return err
}

func GetStatistics(db *sql.DB) (map[string]any, error) {
	stats := make(map[string]any)

	var totalSessions int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	if err != nil {
		return nil, err
	}
	stats["total_sessions"] = totalSessions

	var activeSessions int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE is_active = 1").Scan(&activeSessions)
	if err != nil {
		return nil, err
	}
	stats["active_sessions"] = activeSessions

	var hyperfocusCount int
	err = db.QueryRow("SELECT COUNT(*) FROM hyperfocus_sessions").Scan(&hyperfocusCount)
	if err != nil {
		return nil, err
	}
	stats["hyperfocus_count"] = hyperfocusCount

	var wellbeingResponses int
	err = db.QueryRow("SELECT COUNT(*) FROM wellbeing_checks").Scan(&wellbeingResponses)
	if err != nil {
		return nil, err
	}
	stats["wellbeing_responses"] = wellbeingResponses

	return stats, nil
}

func GetTotalSessions(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	return count, err
}

func GetActiveSessions(db *sql.DB) ([]models.Session, error) {
	query := `
		SELECT id, subject, start_time, end_time, is_active 
		FROM sessions 
		WHERE is_active = 1
		ORDER BY start_time DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		var endTime sql.NullTime
		err := rows.Scan(&session.ID, &session.Subject, &session.StartTime, &endTime, &session.IsActive)
		if err != nil {
			return nil, err
		}
		if endTime.Valid {
			session.EndTime = endTime.Time
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func GetHyperfocusCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM hyperfocus_sessions").Scan(&count)
	return count, err
}
