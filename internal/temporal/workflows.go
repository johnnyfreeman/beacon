package temporal

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MonitorEndpointWorkflow(ctx workflow.Context, endpointID uuid.UUID, intervalSec int) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for {
		var result PingResult
		err := workflow.ExecuteActivity(ctx, "PingEndpoint", endpointID).Get(ctx, &result)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to ping endpoint", "error", err)
		} else {
			err = workflow.ExecuteActivity(ctx, "CheckIncidentStatus", endpointID, result.Success).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to check incident status", "error", err)
			}
		}
		
		// Use workflow.Sleep instead of time.Sleep to properly handle cancellation
		err = workflow.Sleep(ctx, time.Duration(intervalSec)*time.Second)
		if err != nil {
			// Workflow was cancelled
			workflow.GetLogger(ctx).Info("Monitoring workflow cancelled", "endpoint", endpointID)
			return err
		}
	}
}

func AggregateMetricsWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for {
		now := workflow.Now(ctx)
		windowEnd := now.Truncate(5 * time.Minute)
		windowStart := windowEnd.Add(-5 * time.Minute)
		
		var endpoints []uuid.UUID
		err := workflow.ExecuteActivity(ctx, "GetEnabledEndpoints").Get(ctx, &endpoints)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to get enabled endpoints", "error", err)
		} else {
			for _, endpointID := range endpoints {
				err = workflow.ExecuteActivity(ctx, "AggregateMetrics", endpointID, windowStart, windowEnd).Get(ctx, nil)
				if err != nil {
					workflow.GetLogger(ctx).Error("Failed to aggregate metrics", "endpoint", endpointID, "error", err)
				}
			}
		}
		
		err = workflow.Sleep(ctx, 5*time.Minute)
		if err != nil {
			// Workflow was cancelled
			workflow.GetLogger(ctx).Info("Aggregate metrics workflow cancelled")
			return err
		}
	}
}

func CleanupWorkflow(ctx workflow.Context, retentionDays int) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for {
		err := workflow.ExecuteActivity(ctx, "CleanupOldData", retentionDays).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to cleanup old data", "error", err)
		}
		
		err = workflow.Sleep(ctx, 24*time.Hour)
		if err != nil {
			// Workflow was cancelled
			workflow.GetLogger(ctx).Info("Cleanup workflow cancelled")
			return err
		}
	}
}