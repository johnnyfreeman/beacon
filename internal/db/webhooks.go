package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (db *DB) CreateWebhook(webhook *models.Webhook) error {
	webhook.ID = uuid.New()
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	query := `
		INSERT INTO webhooks 
		(id, service_id, name, url, events, headers, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := db.Exec(query, 
		webhook.ID, webhook.ServiceID, webhook.Name, webhook.URL,
		pq.Array(webhook.Events), webhook.Headers, webhook.Enabled,
		webhook.CreatedAt, webhook.UpdatedAt)
	return err
}

func (db *DB) GetWebhook(id uuid.UUID) (*models.Webhook, error) {
	var webhook models.Webhook
	query := `SELECT * FROM webhooks WHERE id = $1 AND deleted_at IS NULL`
	err := db.Get(&webhook, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &webhook, nil
}

func (db *DB) ListWebhooks(serviceID *uuid.UUID) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	var query string
	var args []interface{}

	if serviceID != nil {
		query = `SELECT * FROM webhooks WHERE service_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
		args = append(args, *serviceID)
	} else {
		query = `SELECT * FROM webhooks WHERE deleted_at IS NULL ORDER BY created_at DESC`
	}

	err := db.Select(&webhooks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return webhooks, nil
}

func (db *DB) ListEnabledWebhooks(serviceID uuid.UUID, event string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	query := `
		SELECT * FROM webhooks 
		WHERE service_id = $1 AND enabled = true AND $2 = ANY(events) AND deleted_at IS NULL
	`
	err := db.Select(&webhooks, query, serviceID, event)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled webhooks: %w", err)
	}
	return webhooks, nil
}

func (db *DB) UpdateWebhook(webhook *models.Webhook) error {
	webhook.UpdatedAt = time.Now()
	query := `
		UPDATE webhooks 
		SET name = $2, url = $3, events = $4, headers = $5, enabled = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := db.Exec(query, 
		webhook.ID, webhook.Name, webhook.URL, pq.Array(webhook.Events),
		webhook.Headers, webhook.Enabled, webhook.UpdatedAt)
	return err
}

func (db *DB) DeleteWebhook(id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE webhooks SET deleted_at = $2 WHERE id = $1`
	_, err := db.Exec(query, id, now)
	return err
}