package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

func (db *DB) CreatePingWindow(window *models.PingWindow) error {
	window.ID = uuid.New()
	window.CreatedAt = time.Now()

	query := `
		INSERT INTO ping_windows 
		(id, endpoint_id, window_start, window_end, total_pings, success_pings, 
		 avg_response_ms, min_response_ms, max_response_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := db.Exec(query, 
		window.ID, window.EndpointID, window.WindowStart, window.WindowEnd,
		window.TotalPings, window.SuccessPings, window.AvgResponseMs,
		window.MinResponseMs, window.MaxResponseMs, window.CreatedAt)
	return err
}

func (db *DB) GetPingWindow(id uuid.UUID) (*models.PingWindow, error) {
	var window models.PingWindow
	query := `SELECT * FROM ping_windows WHERE id = $1`
	err := db.Get(&window, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ping window: %w", err)
	}
	return &window, nil
}

func (db *DB) ListPingWindows(endpointID uuid.UUID, limit int) ([]models.PingWindow, error) {
	var windows []models.PingWindow
	query := `
		SELECT * FROM ping_windows 
		WHERE endpoint_id = $1 
		ORDER BY window_start DESC 
		LIMIT $2
	`
	err := db.Select(&windows, query, endpointID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ping windows: %w", err)
	}
	return windows, nil
}

func (db *DB) ListPingWindowsByTimeRange(endpointID uuid.UUID, start, end time.Time) ([]models.PingWindow, error) {
	var windows []models.PingWindow
	query := `
		SELECT * FROM ping_windows 
		WHERE endpoint_id = $1 AND window_start >= $2 AND window_end <= $3
		ORDER BY window_start DESC
	`
	err := db.Select(&windows, query, endpointID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list ping windows by time range: %w", err)
	}
	return windows, nil
}

func (db *DB) DeleteOldPingWindows(before time.Time) error {
	query := `DELETE FROM ping_windows WHERE created_at < $1`
	_, err := db.Exec(query, before)
	return err
}