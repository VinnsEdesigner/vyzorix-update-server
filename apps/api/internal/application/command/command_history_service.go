package command

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// HistoryService handles command history operations.
type HistoryService struct {
	commandRepo command.Repository
	devRepo     device.Repository
}

// NewHistoryService creates a new command history service.
func NewHistoryService(commandRepo command.Repository, devRepo device.Repository) *HistoryService {
	return &HistoryService{
		commandRepo: commandRepo,
		devRepo:     devRepo,
	}
}

// GetHistoryRequest represents a request for command history.
type GetHistoryRequest struct {
	DeviceID       string
	OrganizationID string
	Status         string
	Page           int
	Limit          int
	StartTime      int64
	EndTime        int64
}

// HistoryResponse represents paginated command history.
type HistoryResponse struct {
	Commands   []CommandEntry `json:"commands"`
	Pagination PaginationInfo `json:"pagination"`
}

// CommandEntry represents a command in history.
type CommandEntry struct {
	ID            string `json:"id,omitempty"`
	DispatchID    string `json:"dispatchId"`
	DeviceID      string `json:"deviceId"`
	Command       string `json:"command"`
	Status        string `json:"status"`
	FailureReason string `json:"failureReason,omitempty"`
	SentAt        int64  `json:"sentAt"`
	DeliveredAt   int64  `json:"deliveredAt,omitempty"`
	CompletedAt   int64  `json:"completedAt,omitempty"`
	LatencyMs     int64  `json:"latencyMs,omitempty"`
}

// PaginationInfo represents pagination metadata.
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasMore    bool `json:"hasMore"`
}

// GetHistory retrieves paginated command history for a device.
func (s *HistoryService) GetHistory(ctx context.Context, req *GetHistoryRequest) (*HistoryResponse, error) {
	// Validate device belongs to organization.
	if _, err := s.devRepo.FindByIDAndOrganization(ctx, req.DeviceID, req.OrganizationID); err != nil {
		if errors.Is(err, device.ErrNotFound) {
			return nil, err
		}
		return nil, err
	}

	// Apply defaults.
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Default time range: last 30 days.
	startTime := time.Now().Add(-30 * 24 * time.Hour)
	endTime := time.Now()

	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
	}
	if req.EndTime > 0 {
		endTime = time.UnixMilli(req.EndTime)
	}

	offset := (req.Page - 1) * req.Limit

	// Get command history.
	commands, total, err := s.commandRepo.FindHistoryByDeviceID(
		ctx, req.DeviceID, req.Status, startTime, endTime, req.Limit, offset,
	)
	if err != nil {
		return nil, err
	}

	// Build response.
	entries := make([]CommandEntry, 0, len(commands))
	for _, cmd := range commands {
		sentAt := cmd.CreatedAt.UnixMilli()
		entry := CommandEntry{
			ID:            cmd.ID,
			DispatchID:    cmd.DispatchID,
			DeviceID:      req.DeviceID,
			Command:       string(cmd.Command),
			Status:        string(cmd.Status),
			SentAt:        sentAt,
			FailureReason: cmd.FailureReason,
		}

		if cmd.DeliveredAt != nil {
			entry.DeliveredAt = *cmd.DeliveredAt
			// Calculate latency: time from sent to delivered.
			entry.LatencyMs = *cmd.DeliveredAt - sentAt
		}
		if cmd.CompletedAt != nil {
			entry.CompletedAt = *cmd.CompletedAt
		}

		entries = append(entries, entry)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + req.Limit - 1) / req.Limit
	}

	return &HistoryResponse{
		Commands: entries,
		Pagination: PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    req.Page < totalPages,
		},
	}, nil
}
