package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// Service handles metrics operations.
type Service struct {
	metricsRepo        metrics.Repository
	deviceSettingsRepo device.DeviceSettingsRepository
	orgSettingsRepo    organization.OrganizationSettingsRepository
}

// NewService creates a new metrics service.
func NewService(
	metricsRepo metrics.Repository,
	deviceSettingsRepo device.DeviceSettingsRepository,
	orgSettingsRepo organization.OrganizationSettingsRepository,
) *Service {
	return &Service{
		metricsRepo:        metricsRepo,
		deviceSettingsRepo: deviceSettingsRepo,
		orgSettingsRepo:    orgSettingsRepo,
	}
}

// GetDeviceMetrics retrieves aggregated metrics for chart visualization.
func (s *Service) GetDeviceMetrics(ctx context.Context, req *GetMetricsRequest) (*GetMetricsResponse, error) {
	// Determine time range
	timeRange, resolution, startTime, endTime := s.parseTimeRange(req)

	// Get latest telemetry for current values
	latest, err := s.metricsRepo.GetLatestTelemetry(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest telemetry: %w", err)
	}

	// Get thresholds using hierarchical resolution: device → org → default
	thresholds := s.getResolvedThresholds(ctx, req.DeviceID, req.OrganizationID)

	// Get stats for each metric
	riskScoreStats, err := s.metricsRepo.GetMetricStats(ctx, req.DeviceID, "riskScore", startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score stats: %w", err)
	}

	thermalStats, err := s.metricsRepo.GetMetricStats(ctx, req.DeviceID, "thermalTemp", startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get thermal stats: %w", err)
	}

	bufferStats, err := s.metricsRepo.GetMetricStats(ctx, req.DeviceID, "bufferLevel", startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get buffer stats: %w", err)
	}

	// Get chart data for each metric
	riskScoreChart, err := s.metricsRepo.GetAggregatedMetrics(ctx, req.DeviceID, "riskScore", startTime, endTime, resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score chart: %w", err)
	}

	thermalChart, err := s.metricsRepo.GetAggregatedMetrics(ctx, req.DeviceID, "thermalTemp", startTime, endTime, resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to get thermal chart: %w", err)
	}

	bufferChart, err := s.metricsRepo.GetAggregatedMetrics(ctx, req.DeviceID, "bufferLevel", startTime, endTime, resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to get buffer chart: %w", err)
	}

	// Get threshold breach events
	events, err := s.metricsRepo.GetThresholdBreachEvents(ctx, req.DeviceID, startTime, endTime, thresholds)
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold events: %w", err)
	}

	// Get current values from latest telemetry
	currentRiskScore := float64(0)
	currentThermal := float64(0)
	currentBuffer := float64(0)
	currentUptime := int64(0)
	if latest != nil {
		currentRiskScore = latest.RiskScore
		currentThermal = latest.ThermalTemp
		currentBuffer = latest.BufferLevel
		currentUptime = latest.Uptime
	}

	// Build response
	return &GetMetricsResponse{
		Device: DeviceInfoResponse{
			IMEI: req.DeviceID,
		},
		TimeRange: TimeRangeResponse{
			Start:      startTime.UnixMilli(),
			End:        endTime.UnixMilli(),
			Range:      timeRange,
			Resolution: resolution,
		},
		Metrics: MetricsCollectionDTO{
			RiskScore: MetricDataDTO{
				Current: currentRiskScore,
				Avg:     riskScoreStats.Avg,
				Min:     riskScoreStats.Min,
				Max:     riskScoreStats.Max,
				Unit:    "%",
				Chart:   s.convertChartPoints(riskScoreChart),
				Threshold: ThresholdDTO{
					Warning:  thresholds.RiskScoreWarning,
					Critical: thresholds.RiskScoreCritical,
				},
			},
			ThermalTemp: MetricDataDTO{
				Current: currentThermal,
				Avg:     thermalStats.Avg,
				Min:     thermalStats.Min,
				Max:     thermalStats.Max,
				Unit:    "°C",
				Chart:   s.convertChartPoints(thermalChart),
				Threshold: ThresholdDTO{
					Warning:  thresholds.ThermalWarning,
					Critical: thresholds.ThermalCritical,
				},
			},
			BufferLevel: MetricDataDTO{
				Current: currentBuffer,
				Avg:     bufferStats.Avg,
				Min:     bufferStats.Min,
				Max:     bufferStats.Max,
				Unit:    "%",
				Chart:   s.convertChartPoints(bufferChart),
				Threshold: ThresholdDTO{
					Warning:  thresholds.BufferWarning,
					Critical: thresholds.BufferCritical,
				},
			},
			Uptime: MetricDataDTO{
				Current: float64(currentUptime),
				Avg:     0,
				Min:     0,
				Max:    0,
				Unit:    "s",
				Chart:   []MetricPointDTO{},
				Threshold: ThresholdDTO{
					Warning:  0,
					Critical: 0,
				},
			},
		},
		Events: s.convertThresholdEvents(events),
	}, nil
}

// GetTelemetry retrieves raw telemetry frames.
func (s *Service) GetTelemetry(ctx context.Context, req *GetTelemetryRequest) (*GetTelemetryResponse, error) {
	// Apply defaults
	if req.Limit <= 0 {
		req.Limit = 500
	}
	if req.Limit > 10000 {
		req.Limit = 10000
	}

	// Calculate time range
	endTime := time.Now()
	if req.EndTime > 0 {
		endTime = time.UnixMilli(req.EndTime)
	}

	startTime := endTime.Add(-6 * time.Hour) // Default: last 6 hours
	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
	}

	// Get telemetry frames
	frames, err := s.metricsRepo.GetTelemetryFrames(ctx, req.DeviceID, startTime, endTime, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry frames: %w", err)
	}

	// Build response
	response := &GetTelemetryResponse{
		Frames: make([]TelemetryFrameDTO, 0, len(frames)),
		Stats: TelemetryStatsDTO{
			RiskScore:   MetricStatsDTO{},
			ThermalTemp: MetricStatsDTO{},
			BufferLevel: MetricStatsDTO{},
		},
	}

	var riskScores, thermals, buffers []float64

	for _, frame := range frames {
		response.Frames = append(response.Frames, TelemetryFrameDTO{
			Timestamp:   frame.Timestamp.UnixMilli(),
			RiskScore:   frame.RiskScore,
			ThermalTemp: frame.ThermalTemp,
			BufferLevel: frame.BufferLevel,
			Uptime:      frame.Uptime,
		})

		riskScores = append(riskScores, frame.RiskScore)
		thermals = append(thermals, frame.ThermalTemp)
		buffers = append(buffers, frame.BufferLevel)
	}

	// Calculate stats
	if len(riskScores) > 0 {
		response.Stats.RiskScore = s.calculateStats(riskScores)
		response.Stats.ThermalTemp = s.calculateStats(thermals)
		response.Stats.BufferLevel = s.calculateStats(buffers)
	}

	return response, nil
}

// parseTimeRange determines the time range, resolution, and time bounds.
func (s *Service) parseTimeRange(req *GetMetricsRequest) (rangeStr, resolution string, startTime, endTime time.Time) {
	endTime = time.Now()
	resolution = req.Resolution

	// Handle explicit time range
	if req.StartTime > 0 && req.EndTime > 0 {
		startTime = time.UnixMilli(req.StartTime)
		endTime = time.UnixMilli(req.EndTime)
		rangeStr = "custom"
		if resolution == "" {
			resolution = ResolutionAuto
		}
		return rangeStr, resolution, startTime, endTime
	}

	// Handle range preset
	rangeStr = req.Range
	if rangeStr == "" {
		rangeStr = TimeRange6h // Default
	}

	switch rangeStr {
	case TimeRange1h:
		startTime = endTime.Add(-1 * time.Hour)
		if resolution == "" || resolution == ResolutionAuto {
			resolution = Resolution1m
		}
	case TimeRange6h:
		startTime = endTime.Add(-6 * time.Hour)
		if resolution == "" || resolution == ResolutionAuto {
			resolution = Resolution5m
		}
	case TimeRange24h:
		startTime = endTime.Add(-24 * time.Hour)
		if resolution == "" || resolution == ResolutionAuto {
			resolution = Resolution15m
		}
	case TimeRange7d:
		startTime = endTime.Add(-7 * 24 * time.Hour)
		if resolution == "" || resolution == ResolutionAuto {
			resolution = Resolution1h
		}
	default:
		rangeStr = TimeRange6h
		startTime = endTime.Add(-6 * time.Hour)
		if resolution == "" || resolution == ResolutionAuto {
			resolution = Resolution5m
		}
	}

	return rangeStr, resolution, startTime, endTime
}

// convertChartPoints converts domain chart points to DTO format.
func (s *Service) convertChartPoints(points []*metrics.MetricDataPoint) []MetricPointDTO {
	result := make([]MetricPointDTO, 0, len(points))
	for _, p := range points {
		result = append(result, MetricPointDTO{
			Timestamp: p.Timestamp.UnixMilli(),
			Value:     p.Value,
		})
	}
	return result
}

// convertThresholdEvents converts domain events to DTO format.
func (s *Service) convertThresholdEvents(events []*metrics.MetricThresholdEvent) []ThresholdEventDTO {
	result := make([]ThresholdEventDTO, 0, len(events))
	for _, e := range events {
		result = append(result, ThresholdEventDTO{
			Timestamp: e.Timestamp.UnixMilli(),
			Type:      e.Type,
			Metric:    e.Metric,
			Value:     e.Value,
			Threshold: e.Threshold,
		})
	}
	return result
}

// getResolvedThresholds retrieves thresholds using hierarchical resolution:
// device settings → organization settings → defaults
func (s *Service) getResolvedThresholds(ctx context.Context, deviceID, orgID string) *metrics.ThresholdPreset {
	result := defaultThresholds()

	// Get organization thresholds
	if orgID != "" && s.orgSettingsRepo != nil {
		orgSettings, err := s.orgSettingsRepo.FindByOrganizationID(ctx, orgID)
		if err == nil && orgSettings != nil && orgSettings.DefaultThresholds != nil {
			result.RiskScoreWarning = float64(orgSettings.DefaultThresholds.RiskWarn)
			result.RiskScoreCritical = float64(orgSettings.DefaultThresholds.RiskCrit)
			result.ThermalWarning = float64(orgSettings.DefaultThresholds.ThermalWarn)
			result.ThermalCritical = float64(orgSettings.DefaultThresholds.ThermalCrit)
			result.BufferWarning = float64(orgSettings.DefaultThresholds.BufferWarn)
			result.BufferCritical = float64(orgSettings.DefaultThresholds.BufferCrit)
		}
	}

	// Override with device-specific thresholds
	if deviceID != "" && s.deviceSettingsRepo != nil {
		deviceSettings, err := s.deviceSettingsRepo.FindByDeviceIMEI(ctx, deviceID)
		if err == nil && deviceSettings != nil && deviceSettings.HasThresholds() {
			if deviceSettings.Thresholds.RiskWarn != 0 {
				result.RiskScoreWarning = float64(deviceSettings.Thresholds.RiskWarn)
			}
			if deviceSettings.Thresholds.RiskCrit != 0 {
				result.RiskScoreCritical = float64(deviceSettings.Thresholds.RiskCrit)
			}
			if deviceSettings.Thresholds.ThermalWarn != 0 {
				result.ThermalWarning = float64(deviceSettings.Thresholds.ThermalWarn)
			}
			if deviceSettings.Thresholds.ThermalCrit != 0 {
				result.ThermalCritical = float64(deviceSettings.Thresholds.ThermalCrit)
			}
			if deviceSettings.Thresholds.BufferWarn != 0 {
				result.BufferWarning = float64(deviceSettings.Thresholds.BufferWarn)
			}
			if deviceSettings.Thresholds.BufferCrit != 0 {
				result.BufferCritical = float64(deviceSettings.Thresholds.BufferCrit)
			}
		}
	}

	return result
}

// defaultThresholds returns default threshold values.
func defaultThresholds() *metrics.ThresholdPreset {
	return &metrics.ThresholdPreset{
		RiskScoreWarning:  70,
		RiskScoreCritical: 85,
		ThermalWarning:    45,
		ThermalCritical:   50,
		BufferWarning:     30,
		BufferCritical:    15,
	}
}

// calculateStats computes statistics from a slice of values.
func (s *Service) calculateStats(values []float64) MetricStatsDTO {
	if len(values) == 0 {
		return MetricStatsDTO{}
	}

	stats := MetricStatsDTO{
		Current: values[0],
		Min:     values[0],
		Max:     values[0],
	}

	var sum float64
	for _, v := range values {
		sum += v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	stats.Avg = sum / float64(len(values))

	return stats
}
