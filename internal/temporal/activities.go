package temporal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/models"
	"github.com/google/uuid"
)

type Activities struct {
	DB *db.DB
}

type PingResult struct {
	EndpointID uuid.UUID
	StatusCode int
	ResponseMs int
	Success    bool
	Error      string
}

func (a *Activities) PingEndpoint(ctx context.Context, endpointID uuid.UUID) (*PingResult, error) {
	endpoint, err := a.DB.GetEndpoint(endpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(endpoint.TimeoutMs) * time.Millisecond,
	}

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, endpoint.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Safely handle headers - they might be nil or empty
	if endpoint.Headers != nil && len(endpoint.Headers) > 0 {
		for key, value := range endpoint.Headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	responseMs := int(time.Since(start).Milliseconds())

	result := &PingResult{
		EndpointID: endpointID,
		ResponseMs: responseMs,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.StatusCode = 0
	} else {
		defer resp.Body.Close()
		result.StatusCode = resp.StatusCode
		result.Success = resp.StatusCode == endpoint.ExpectedCode
		if !result.Success {
			result.Error = fmt.Sprintf("Expected status %d but got %d", endpoint.ExpectedCode, resp.StatusCode)
		}
	}

	ping := &models.Ping{
		EndpointID: endpointID,
		StatusCode: result.StatusCode,
		ResponseMs: result.ResponseMs,
		Success:    result.Success,
	}
	if result.Error != "" {
		ping.Error = &result.Error
	}

	if err := a.DB.CreatePing(ping); err != nil {
		return nil, fmt.Errorf("failed to save ping: %w", err)
	}

	return result, nil
}

func (a *Activities) CheckIncidentStatus(ctx context.Context, endpointID uuid.UUID, success bool) error {
	incident, err := a.DB.GetOpenIncident(endpointID)
	
	if success {
		if err == nil && incident != nil {
			if err := a.DB.ResolveIncident(incident.ID); err != nil {
				return fmt.Errorf("failed to resolve incident: %w", err)
			}
			
			endpoint, _ := a.DB.GetEndpoint(endpointID)
			if endpoint != nil {
				if err := a.TriggerWebhooks(ctx, endpoint.ServiceID, "incident_resolved", incident); err != nil {
					return fmt.Errorf("failed to trigger webhooks: %w", err)
				}
			}
		}
	} else {
		if err != nil {
			endpoint, _ := a.DB.GetEndpoint(endpointID)
			incident = &models.Incident{
				EndpointID: endpointID,
				StartedAt:  time.Now(),
				Status:     "open",
				Message:    fmt.Sprintf("Endpoint %s is down", endpoint.Name),
			}
			if err := a.DB.CreateIncident(incident); err != nil {
				return fmt.Errorf("failed to create incident: %w", err)
			}
			
			if endpoint != nil {
				if err := a.TriggerWebhooks(ctx, endpoint.ServiceID, "incident_start", incident); err != nil {
					return fmt.Errorf("failed to trigger webhooks: %w", err)
				}
			}
		}
	}
	
	return nil
}

func (a *Activities) TriggerWebhooks(ctx context.Context, serviceID uuid.UUID, event string, incident *models.Incident) error {
	webhooks, err := a.DB.ListEnabledWebhooks(serviceID, event)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, webhook := range webhooks {
		payload := map[string]interface{}{
			"event":       event,
			"incident_id": incident.ID,
			"endpoint_id": incident.EndpointID,
			"started_at":  incident.StartedAt,
			"message":     incident.Message,
		}
		if incident.ResolvedAt != nil {
			payload["resolved_at"] = incident.ResolvedAt
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payloadBytes))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		for key, value := range webhook.Headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}

		client.Do(req)
	}

	return nil
}

func (a *Activities) AggregateMetrics(ctx context.Context, endpointID uuid.UUID, windowStart, windowEnd time.Time) error {
	pings, err := a.DB.ListPingsByTimeRange(endpointID, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to list pings: %w", err)
	}

	if len(pings) == 0 {
		return nil
	}

	window := &models.PingWindow{
		EndpointID:  endpointID,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		TotalPings:  len(pings),
	}

	var totalResponseMs int
	minResponseMs := pings[0].ResponseMs
	maxResponseMs := pings[0].ResponseMs

	for _, ping := range pings {
		if ping.Success {
			window.SuccessPings++
		}
		totalResponseMs += ping.ResponseMs
		if ping.ResponseMs < minResponseMs {
			minResponseMs = ping.ResponseMs
		}
		if ping.ResponseMs > maxResponseMs {
			maxResponseMs = ping.ResponseMs
		}
	}

	window.AvgResponseMs = totalResponseMs / len(pings)
	window.MinResponseMs = minResponseMs
	window.MaxResponseMs = maxResponseMs

	if err := a.DB.CreatePingWindow(window); err != nil {
		return fmt.Errorf("failed to create ping window: %w", err)
	}

	return nil
}

func (a *Activities) CleanupOldData(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	
	if err := a.DB.DeleteOldPings(cutoffTime); err != nil {
		return fmt.Errorf("failed to delete old pings: %w", err)
	}
	
	if err := a.DB.DeleteOldPingWindows(cutoffTime); err != nil {
		return fmt.Errorf("failed to delete old ping windows: %w", err)
	}
	
	return nil
}

func (a *Activities) GetEnabledEndpoints(ctx context.Context) ([]uuid.UUID, error) {
	endpoints, err := a.DB.ListEnabledEndpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled endpoints: %w", err)
	}

	ids := make([]uuid.UUID, len(endpoints))
	for i, ep := range endpoints {
		ids[i] = ep.ID
	}
	return ids, nil
}