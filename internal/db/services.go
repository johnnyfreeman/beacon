package db

import (
	"fmt"
	"time"

	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

func (db *DB) CreateService(service *models.Service) error {
	service.ID = uuid.New()
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	query := `
		INSERT INTO services (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(query, service.ID, service.Name, service.Description, service.CreatedAt, service.UpdatedAt)
	return err
}

func (db *DB) GetService(id uuid.UUID) (*models.Service, error) {
	var service models.Service
	query := `SELECT * FROM services WHERE id = $1 AND deleted_at IS NULL`
	err := db.Get(&service, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return &service, nil
}

func (db *DB) ListServices() ([]models.Service, error) {
	var services []models.Service
	query := `SELECT * FROM services WHERE deleted_at IS NULL ORDER BY created_at DESC`
	err := db.Select(&services, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return services, nil
}

func (db *DB) UpdateService(service *models.Service) error {
	service.UpdatedAt = time.Now()
	query := `
		UPDATE services 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := db.Exec(query, service.ID, service.Name, service.Description, service.UpdatedAt)
	return err
}

func (db *DB) DeleteService(id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE services SET deleted_at = $2 WHERE id = $1`
	_, err := db.Exec(query, id, now)
	return err
}