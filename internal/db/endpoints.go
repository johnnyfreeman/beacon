package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

func (db *DB) CreateEndpoint(endpoint *models.ServiceEndpoint) error {
	endpoint.ID = uuid.New()
	endpoint.CreatedAt = time.Now()
	endpoint.UpdatedAt = time.Now()

	query := `
		INSERT INTO service_endpoints 
		(id, service_id, name, url, method, headers, expected_code, timeout_ms, interval_sec, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := db.Exec(query, 
		endpoint.ID, endpoint.ServiceID, endpoint.Name, endpoint.URL, endpoint.Method,
		endpoint.Headers, endpoint.ExpectedCode, endpoint.TimeoutMs, endpoint.IntervalSec,
		endpoint.Enabled, endpoint.CreatedAt, endpoint.UpdatedAt)
	return err
}

func (db *DB) GetEndpoint(id uuid.UUID) (*models.ServiceEndpoint, error) {
	var endpoint models.ServiceEndpoint
	query := `SELECT id, service_id, name, url, method, headers, expected_code, timeout_ms, interval_sec, enabled, created_at, updated_at, deleted_at FROM service_endpoints WHERE id = $1 AND deleted_at IS NULL`
	err := db.Get(&endpoint, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}
	// Ensure headers is initialized
	if endpoint.Headers == nil {
		endpoint.Headers = make(models.JSONB)
	}
	return &endpoint, nil
}

func (db *DB) ListEndpoints(serviceID *uuid.UUID) ([]models.ServiceEndpoint, error) {
	var endpoints []models.ServiceEndpoint
	var query string
	var args []interface{}

	if serviceID != nil {
		query = `SELECT id, service_id, name, url, method, headers, expected_code, timeout_ms, interval_sec, enabled, created_at, updated_at, deleted_at FROM service_endpoints WHERE service_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
		args = append(args, *serviceID)
	} else {
		query = `SELECT id, service_id, name, url, method, headers, expected_code, timeout_ms, interval_sec, enabled, created_at, updated_at, deleted_at FROM service_endpoints WHERE deleted_at IS NULL ORDER BY created_at DESC`
	}

	err := db.Select(&endpoints, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}
	return endpoints, nil
}

func (db *DB) UpdateEndpoint(endpoint *models.ServiceEndpoint) error {
	endpoint.UpdatedAt = time.Now()
	query := `
		UPDATE service_endpoints 
		SET name = $2, url = $3, method = $4, headers = $5, expected_code = $6, 
		    timeout_ms = $7, interval_sec = $8, enabled = $9, updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := db.Exec(query, 
		endpoint.ID, endpoint.Name, endpoint.URL, endpoint.Method, endpoint.Headers,
		endpoint.ExpectedCode, endpoint.TimeoutMs, endpoint.IntervalSec, endpoint.Enabled,
		endpoint.UpdatedAt)
	return err
}

func (db *DB) DeleteEndpoint(id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE service_endpoints SET deleted_at = $2 WHERE id = $1`
	_, err := db.Exec(query, id, now)
	return err
}

func (db *DB) ListEnabledEndpoints() ([]models.ServiceEndpoint, error) {
	var endpoints []models.ServiceEndpoint
	query := `SELECT id, service_id, name, url, method, headers, expected_code, timeout_ms, interval_sec, enabled, created_at, updated_at, deleted_at FROM service_endpoints WHERE enabled = true AND deleted_at IS NULL`
	err := db.Select(&endpoints, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled endpoints: %w", err)
	}
	// Ensure all headers are initialized
	for i := range endpoints {
		if endpoints[i].Headers == nil {
			endpoints[i].Headers = make(models.JSONB)
		}
	}
	return endpoints, nil
}