package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

func (db *DB) CreateIncident(incident *models.Incident) error {
	incident.ID = uuid.New()
	incident.CreatedAt = time.Now()
	incident.UpdatedAt = time.Now()

	query := `
		INSERT INTO incidents 
		(id, endpoint_id, started_at, status, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query, 
		incident.ID, incident.EndpointID, incident.StartedAt,
		incident.Status, incident.Message, incident.CreatedAt, incident.UpdatedAt)
	return err
}

func (db *DB) GetIncident(id uuid.UUID) (*models.Incident, error) {
	var incident models.Incident
	query := `SELECT * FROM incidents WHERE id = $1`
	err := db.Get(&incident, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}
	return &incident, nil
}

func (db *DB) ListIncidents(endpointID *uuid.UUID, status string) ([]models.Incident, error) {
	var incidents []models.Incident
	var query string
	var args []interface{}

	if endpointID != nil && status != "" {
		query = `SELECT * FROM incidents WHERE endpoint_id = $1 AND status = $2 ORDER BY started_at DESC`
		args = append(args, *endpointID, status)
	} else if endpointID != nil {
		query = `SELECT * FROM incidents WHERE endpoint_id = $1 ORDER BY started_at DESC`
		args = append(args, *endpointID)
	} else if status != "" {
		query = `SELECT * FROM incidents WHERE status = $1 ORDER BY started_at DESC`
		args = append(args, status)
	} else {
		query = `SELECT * FROM incidents ORDER BY started_at DESC`
	}

	err := db.Select(&incidents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	return incidents, nil
}

func (db *DB) GetOpenIncident(endpointID uuid.UUID) (*models.Incident, error) {
	var incident models.Incident
	query := `SELECT * FROM incidents WHERE endpoint_id = $1 AND status = 'open' ORDER BY started_at DESC LIMIT 1`
	err := db.Get(&incident, query, endpointID)
	if err != nil {
		return nil, err // May return sql.ErrNoRows
	}
	return &incident, nil
}

func (db *DB) ResolveIncident(id uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE incidents 
		SET status = 'resolved', resolved_at = $2, updated_at = $3
		WHERE id = $1
	`
	_, err := db.Exec(query, id, now, now)
	return err
}

func (db *DB) UpdateIncident(incident *models.Incident) error {
	incident.UpdatedAt = time.Now()
	query := `
		UPDATE incidents 
		SET message = $2, updated_at = $3
		WHERE id = $1
	`
	_, err := db.Exec(query, incident.ID, incident.Message, incident.UpdatedAt)
	return err
}