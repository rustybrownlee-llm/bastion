package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

type Logger struct {
	db *sql.DB
}

func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

func (l *Logger) Log(eventType, userID string, details map[string]interface{}, ipAddress string) {
	var detailsJSON interface{}
	var err error

	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			log.Printf("failed to marshal audit details: %v", err)
			return
		}
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	_, err = l.db.Exec(
		`INSERT INTO audit_log (event_type, user_id, details, ip_address)
		 VALUES ($1, $2, $3, $4)`,
		eventType, userIDPtr, detailsJSON, ipAddress,
	)

	if err != nil {
		log.Printf("failed to insert audit log: %v", err)
	}
}

func (l *Logger) LogError(eventType string, err error, ipAddress string) {
	details := map[string]interface{}{
		"error": err.Error(),
	}
	l.Log(eventType, "", details, ipAddress)
}

func (l *Logger) LogWithUser(eventType, userID string, ipAddress string) {
	l.Log(eventType, userID, nil, ipAddress)
}

func (l *Logger) GetRecentEvents(limit int) ([]map[string]interface{}, error) {
	rows, err := l.db.Query(
		`SELECT id, event_type, user_id, details, ip_address, created_at
		 FROM audit_log
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit log: %w", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, eventType, ipAddress string
		var userID sql.NullString
		var details sql.RawBytes
		var createdAt string

		if err := rows.Scan(&id, &eventType, &userID, &details, &ipAddress, &createdAt); err != nil {
			continue
		}

		event := map[string]interface{}{
			"id":         id,
			"event_type": eventType,
			"ip_address": ipAddress,
			"created_at": createdAt,
		}

		if userID.Valid {
			event["user_id"] = userID.String
		}

		if len(details) > 0 {
			var detailsMap map[string]interface{}
			if err := json.Unmarshal(details, &detailsMap); err == nil {
				event["details"] = detailsMap
			}
		}

		events = append(events, event)
	}

	return events, nil
}
