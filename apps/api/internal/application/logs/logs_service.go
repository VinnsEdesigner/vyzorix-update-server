package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/logs"
)

// Service handles device logs operations.
type Service struct {
	logsRepo logs.Repository
	logger  *slog.Logger
}

// NewService creates a new logs service.
func NewService(logsRepo logs.Repository, logger *slog.Logger) *Service {
	return &Service{logsRepo: logsRepo, logger: logger}
}

// GetDeviceLogs retrieves paginated device logs with cursor-based pagination.
func (s *Service) GetDeviceLogs(ctx context.Context, req *ListLogsRequest) (*ListLogsResponse, error) {
	// Apply defaults.
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	// Calculate time range.
	endTime := time.Now()
	if req.EndTime > 0 {
		endTime = time.UnixMilli(req.EndTime)
	}

	startTime := endTime.Add(-24 * time.Hour) // Default: last 24 hours.
	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
	}

	// Validate event type if provided.
	if req.EventType != "" && !logs.IsValidEventType(req.EventType) {
		return nil, logs.ErrInvalidEventType
	}

	// Fetch logs.
	logList, nextCursor, err := s.logsRepo.ListLogs(ctx, req.DeviceID, req.EventType, startTime, endTime, req.Limit, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list device logs: %w", err)
	}

	// Convert to response format.
	events := make([]LogEvent, 0, len(logList))
	for _, log := range logList {
		var data map[string]interface{}
		if log.Data != nil {
			if err := json.Unmarshal(log.Data, &data); err != nil {
				s.logger.Warn("Failed to unmarshal log data, using empty map",
					"logID", log.ID, "error", err)
				data = make(map[string]interface{})
			}
		}

		events = append(events, LogEvent{
			ID:        log.ID,
			Type:      log.EventType,
			Timestamp: log.Timestamp.UnixMilli(),
			Data:      data,
		})
	}

	return &ListLogsResponse{
		Events: events,
		Pagination: CursorPagination{
			Limit:      req.Limit,
			HasMore:    nextCursor != "",
			NextCursor: nextCursor,
		},
	}, nil
}

// CreateLog creates a new device log entry.
func (s *Service) CreateLog(ctx context.Context, deviceID, eventType string, data map[string]interface{}) error {
	var dataJSON json.RawMessage
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal log data: %w", err)
		}
	}

	log := &logs.DeviceLog{
		DeviceID:  deviceID,
		EventType: eventType,
		Timestamp: time.Now(),
		Data:     dataJSON,
	}

	return s.logsRepo.CreateLog(ctx, log)
}
