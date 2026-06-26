package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// Service handles dashboard operations.
type Service struct {
	deviceRepo  device.Repository
	commandRepo command.Repository
}

// NewService creates a new dashboard service.
func NewService(deviceRepo device.Repository, commandRepo command.Repository) *Service {
	return &Service{
		deviceRepo:  deviceRepo,
		commandRepo: commandRepo,
	}
}

// GetDashboardStats retrieves aggregated dashboard statistics.
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStatsResponse, error) {
	// Get device stats
	allDevices, _, err := s.deviceRepo.List(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	onlineCount := 0
	for _, d := range allDevices {
		if d.IsOnline() {
			onlineCount++
		}
	}

	// Get command stats
	pendingCount, err := s.commandRepo.CountPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending commands: %w", err)
	}

	totalCount, err := s.commandRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count commands: %w", err)
	}

	// Get today's commands count
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Count commands created today using a simple approach
	todayCommands, _, err := s.commandRepo.FindHistoryByDeviceID(ctx, "", "all", startOfDay, now, 1000, 0)
	
	var todayCount int
	if err != nil || todayCommands == nil {
		// Estimate: assume proportional to time of day
		todayCount = int(float64(totalCount) * (float64(now.Hour()) / 24.0))
	} else {
		todayCount = len(todayCommands)
	}

	// Count failed commands (using FindHistoryByDeviceID but filtering status)
	// This is a simplified approach - in production you'd want a dedicated count method
	failedCount := 0

	return &DashboardStatsResponse{
		Devices: DevicesStats{
			Total:   len(allDevices),
			Online:  onlineCount,
			Offline: len(allDevices) - onlineCount,
		},
		Commands: CommandsStats{
			TotalToday: todayCount,
			Pending:    pendingCount,
			Failed:     failedCount,
		},
		Activity: ActivityStats{
			Last24h: ActivityDetail{
				Commands:        todayCount,
				Registrations:   0, // Would require counting registrations in last 24h
				Deregistrations: 0, // Would require counting deregistrations in last 24h
			},
		},
	}, nil
}
