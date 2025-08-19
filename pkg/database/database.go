package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Init initializes the database connection and ensures all necessary tables are created.
func Init(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQL statements to create tables if they don't exist.
	// This makes the database setup self-contained and automatic.
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
	`

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
		return "No significant recent history available.", nil
	}
	return summary.String(), nil
}
