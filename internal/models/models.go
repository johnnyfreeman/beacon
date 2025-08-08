package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type ServiceEndpoint struct {
	ID            uuid.UUID  `db:"id" json:"ID"`
	ServiceID     uuid.UUID  `db:"service_id" json:"ServiceID"`
	Name          string     `db:"name" json:"Name"`
	URL           string     `db:"url" json:"URL"`
	Method        string     `db:"method" json:"Method"`
	Headers       JSONB      `db:"headers" json:"Headers"`
	ExpectedCode  int        `db:"expected_code" json:"ExpectedCode"`
	TimeoutMs     int        `db:"timeout_ms" json:"TimeoutMs"`
	IntervalSec   int        `db:"interval_sec" json:"IntervalSec"`
	Enabled       bool       `db:"enabled" json:"Enabled"`
	CreatedAt     time.Time  `db:"created_at" json:"CreatedAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"UpdatedAt"`
	DeletedAt     *time.Time `db:"deleted_at" json:"DeletedAt"`
}

type Ping struct {
	ID         uuid.UUID  `db:"id"`
	EndpointID uuid.UUID  `db:"endpoint_id"`
	StatusCode int        `db:"status_code"`
	ResponseMs int        `db:"response_ms"`
	Success    bool       `db:"success"`
	Error      *string    `db:"error"`
	CreatedAt  time.Time  `db:"created_at"`
}

type PingWindow struct {
	ID           uuid.UUID  `db:"id"`
	EndpointID   uuid.UUID  `db:"endpoint_id"`
	WindowStart  time.Time  `db:"window_start"`
	WindowEnd    time.Time  `db:"window_end"`
	TotalPings   int        `db:"total_pings"`
	SuccessPings int        `db:"success_pings"`
	AvgResponseMs int       `db:"avg_response_ms"`
	MinResponseMs int       `db:"min_response_ms"`
	MaxResponseMs int       `db:"max_response_ms"`
	CreatedAt    time.Time  `db:"created_at"`
}

type Incident struct {
	ID         uuid.UUID  `db:"id"`
	EndpointID uuid.UUID  `db:"endpoint_id"`
	StartedAt  time.Time  `db:"started_at"`
	ResolvedAt *time.Time `db:"resolved_at"`
	Status     string     `db:"status"` // open, resolved
	Message    string     `db:"message"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

type Webhook struct {
	ID        uuid.UUID  `db:"id"`
	ServiceID uuid.UUID  `db:"service_id"`
	Name      string     `db:"name"`
	URL       string     `db:"url"`
	Events    []string   `db:"events"` // incident_start, incident_resolved
	Headers   JSONB      `db:"headers"`
	Enabled   bool       `db:"enabled"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	// Initialize to empty map by default
	*j = make(map[string]interface{})
	
	if value == nil {
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		// If it's already a map, try to cast it
		if m, ok := value.(map[string]interface{}); ok {
			*j = JSONB(m)
			return nil
		}
		return fmt.Errorf("cannot scan type %T into JSONB", value)
	}
	
	// If empty data, keep the initialized empty map
	if len(data) == 0 || string(data) == "{}" || string(data) == "null" {
		return nil
	}
	
	// Only unmarshal if we have actual data
	if err := json.Unmarshal(data, j); err != nil {
		// If unmarshal fails, keep the empty map
		return nil
	}
	
	return nil
}

// MarshalJSON implements json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONB: UnmarshalJSON on nil pointer")
	}
	
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = JSONB(m)
	return nil
}