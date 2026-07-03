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

	// Get last 24h time range
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	// Count commands in last 24 hours
	commandsLast24h, _, err := s.commandRepo.FindHistoryByDeviceID(ctx, "", "all", last24h, now, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands from last 24h: %w", err)
	}
	commandsLast24hCount := len(commandsLast24h)

	// Count failed commands in last 24 hours
	failedCount := 0
	for _, cmd := range commandsLast24h {
		if cmd.Status == command.StatusFailed {
			failedCount++
		}
	}

	// Count registrations and deregistrations from device events in last 24h
	// This would require access to device events or logs - for now use placeholder
	// In production, query device_events table for registration/deregistration events
	registrations := 0
	deregistrations := 0

	return &DashboardStatsResponse{
		Devices: DevicesStats{
			Total:   len(allDevices),
			Online:  onlineCount,
			Offline: len(allDevices) - onlineCount,
		},
		Commands: CommandsStats{
			TotalToday: commandsLast24hCount,
			Pending:   pendingCount,
			Failed:    failedCount,
		},
		Activity: ActivityStats{
			Last24h: ActivityDetail{
				Commands:        commandsLast24hCount,
				Registrations:  registrations,
				Deregistrations: deregistrations,
			},
		},
	}, nil
}
