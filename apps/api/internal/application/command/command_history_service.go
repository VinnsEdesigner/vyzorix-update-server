package command

import (
	"context"
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
	DeviceID  string
	Status    string
	Page      int
	Limit     int
	StartTime int64
	EndTime   int64
}

// HistoryResponse represents paginated command history.
type HistoryResponse struct {
	Commands   []CommandEntry `json:"commands"`
	Pagination PaginationInfo `json:"pagination"`
}

// CommandEntry represents a command in history.
type CommandEntry struct {
	ID            string `json:"id"`
	DispatchID    string `json:"dispatchId"`
	Command       string `json:"command"`
	Status        string `json:"status"`
	FailureReason string `json:"failureReason,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
	DeliveredAt   int64  `json:"deliveredAt,omitempty"`
	CompletedAt   int64  `json:"completedAt,omitempty"`
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
	// Validate device
	if _, err := s.devRepo.FindByID(ctx, req.DeviceID); err != nil {
		return nil, err
	}

	// Apply defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Default time range: last 30 days
	startTime := time.Now().Add(-30 * 24 * time.Hour)
	endTime := time.Now()

	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
	}
	if req.EndTime > 0 {
		endTime = time.UnixMilli(req.EndTime)
	}

	offset := (req.Page - 1) * req.Limit

	// Get command history
	commands, total, err := s.commandRepo.FindHistoryByDeviceID(
		ctx, req.DeviceID, req.Status, startTime, endTime, req.Limit, offset,
	)
	if err != nil {
		return nil, err
	}

	// Build response
	entries := make([]CommandEntry, 0, len(commands))
	for _, cmd := range commands {
		entry := CommandEntry{
			ID:            cmd.ID,
			DispatchID:    cmd.DispatchID,
			Command:       string(cmd.Command),
			Status:        string(cmd.Status),
			CreatedAt:     cmd.CreatedAt.UnixMilli(),
			FailureReason: cmd.FailureReason,
		}

		if cmd.DeliveredAt != nil {
			entry.DeliveredAt = *cmd.DeliveredAt
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
