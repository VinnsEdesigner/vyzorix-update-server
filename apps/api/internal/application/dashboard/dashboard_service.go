package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/logs"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/cache"
)

// Service handles dashboard operations.
type Service struct {
	deviceRepo  device.Repository
	commandRepo command.Repository
	logsRepo    logs.Repository
	statsCache  *cache.Section
}

// NewService creates a new dashboard service.
func NewService(deviceRepo device.Repository, commandRepo command.Repository, logsRepo logs.Repository) *Service {
	return &Service{
		deviceRepo:  deviceRepo,
		commandRepo: commandRepo,
		logsRepo:    logsRepo,
	}
}

// SetStatsCache wires the cache for dashboard stats responses.
func (s *Service) SetStatsCache(c *cache.Section) {
	s.statsCache = c
}

// GetDashboardStats retrieves aggregated dashboard statistics for an operator.
func (s *Service) GetDashboardStats(ctx context.Context, operatorID string) (*DashboardStatsResponse, error) {
	// Get device stats filtered by operator.
	allDevices, err := s.deviceRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	onlineCount := 0
	for _, d := range allDevices {
		if d.IsOnline() {
			onlineCount++
		}
	}

	// Get command stats.
	pendingCount, err := s.commandRepo.CountPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending commands: %w", err)
	}

	// Get last 24h time range.
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	// Count commands in last 24 hours.
	commandsLast24h, _, err := s.commandRepo.FindHistoryByDeviceID(ctx, "", "all", last24h, now, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands from last 24h: %w", err)
	}
	commandsLast24hCount := len(commandsLast24h)

	// Count failed commands in last 24 hours.
	failedCount := 0
	for _, cmd := range commandsLast24h {
		if cmd.Status == command.StatusFailed {
			failedCount++
		}
	}

	// Count registrations and deregistrations from device logs in last 24h.
	var registrations, deregistrations int
	if s.logsRepo != nil {
		registrations, err = s.logsRepo.CountLogs(ctx, "", logs.EventTypeRegistration, last24h, now)
		if err != nil {
			// Log error but don't fail the entire request.
			registrations = 0
		}

		deregistrations, err = s.logsRepo.CountLogs(ctx, "", logs.EventTypeDeregistration, last24h, now)
		if err != nil {
			deregistrations = 0
		}
	}

	return &DashboardStatsResponse{
		Devices: DevicesStats{
			Total:   len(allDevices),
			Online:  onlineCount,
			Offline: len(allDevices) - onlineCount,
		},
		Commands: CommandsStats{
			TotalToday: commandsLast24hCount,
			Pending:    pendingCount,
			Failed:     failedCount,
		},
		Activity: ActivityStats{
			Last24h: ActivityDetail{
				Commands:        commandsLast24hCount,
				Registrations:   registrations,
				Deregistrations: deregistrations,
			},
		},
	}, nil
}

// GetDashboardStatsByOrganization retrieves aggregated dashboard statistics for an organization.
func (s *Service) GetDashboardStatsByOrganization(ctx context.Context, orgID string) (*DashboardStatsResponse, error) {
	if s.statsCache != nil {
		if cached, ok := s.statsCache.Get(orgID); ok {
			if stats, ok := cached.(*DashboardStatsResponse); ok {
				return stats, nil
			}
		}
	}
	stats, err := s.getDashboardStatsByOrganizationNow(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if s.statsCache != nil {
		s.statsCache.Set(orgID, stats)
	}
	return stats, nil
}

func (s *Service) getDashboardStatsByOrganizationNow(ctx context.Context, orgID string) (*DashboardStatsResponse, error) {
	// Get device stats filtered by organization.
	allDevices, err := s.deviceRepo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	onlineCount := 0
	for _, d := range allDevices {
		if d.IsOnline() {
			onlineCount++
		}
	}

	// Get command stats (filtered by organization via device-anchored scoping).
	orgDeviceIDs := make([]string, 0, len(allDevices))
	orgDeviceSet := make(map[string]bool, len(allDevices))
	for _, d := range allDevices {
		orgDeviceIDs = append(orgDeviceIDs, d.ID)
		orgDeviceSet[d.ID] = true
	}

	pendingCount, err := s.commandRepo.CountPendingByDeviceIDs(ctx, orgDeviceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending commands: %w", err)
	}

	// Get last 24h time range.
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	// Count commands in last 24 hours (organization-filtered).
	commandsLast24h, _, err := s.commandRepo.FindHistoryByDeviceID(ctx, "", "all", last24h, now, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands from last 24h: %w", err)
	}

	// Filter commands by organization.
	var orgCommands []*command.Command
	for _, cmd := range commandsLast24h {
		if orgDeviceSet[cmd.DeviceID] {
			orgCommands = append(orgCommands, cmd)
		}
	}
	commandsLast24hCount := len(orgCommands)

	// Count failed commands in last 24 hours.
	failedCount := 0
	for _, cmd := range orgCommands {
		if cmd.Status == command.StatusFailed {
			failedCount++
		}
	}

	// Count registrations and deregistrations from device logs in last 24h (org-scoped).
	var registrations, deregistrations int
	if s.logsRepo != nil {
		registrations, err = s.logsRepo.CountLogsByDeviceIDs(ctx, orgDeviceIDs, logs.EventTypeRegistration, last24h, now)
		if err != nil {
			registrations = 0
		}

		deregistrations, err = s.logsRepo.CountLogsByDeviceIDs(ctx, orgDeviceIDs, logs.EventTypeDeregistration, last24h, now)
		if err != nil {
			deregistrations = 0
		}
	}

	return &DashboardStatsResponse{
		Devices: DevicesStats{
			Total:   len(allDevices),
			Online:  onlineCount,
			Offline: len(allDevices) - onlineCount,
		},
		Commands: CommandsStats{
			TotalToday: commandsLast24hCount,
			Pending:    pendingCount,
			Failed:     failedCount,
		},
		Activity: ActivityStats{
			Last24h: ActivityDetail{
				Commands:        commandsLast24hCount,
				Registrations:   registrations,
				Deregistrations: deregistrations,
			},
		},
	}, nil
}
