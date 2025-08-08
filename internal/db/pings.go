package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

func (db *DB) CreatePing(ping *models.Ping) error {
	ping.ID = uuid.New()
	ping.CreatedAt = time.Now()

	query := `
		INSERT INTO pings (id, endpoint_id, status_code, response_ms, success, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query, ping.ID, ping.EndpointID, ping.StatusCode, 
		ping.ResponseMs, ping.Success, ping.Error, ping.CreatedAt)
	return err
}

func (db *DB) GetPing(id uuid.UUID) (*models.Ping, error) {
	var ping models.Ping
	query := `SELECT * FROM pings WHERE id = $1`
	err := db.Get(&ping, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ping: %w", err)
	}
	return &ping, nil
}

func (db *DB) ListPings(endpointID uuid.UUID, limit int) ([]models.Ping, error) {
	var pings []models.Ping
	query := `SELECT * FROM pings WHERE endpoint_id = $1 ORDER BY created_at DESC LIMIT $2`
	err := db.Select(&pings, query, endpointID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pings: %w", err)
	}
	return pings, nil
}

func (db *DB) ListPingsByTimeRange(endpointID uuid.UUID, start, end time.Time) ([]models.Ping, error) {
	var pings []models.Ping
	query := `
		SELECT * FROM pings 
		WHERE endpoint_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
	`
	err := db.Select(&pings, query, endpointID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list pings by time range: %w", err)
	}
	return pings, nil
}

func (db *DB) DeleteOldPings(before time.Time) error {
	query := `DELETE FROM pings WHERE created_at < $1`
	_, err := db.Exec(query, before)
	return err
}