// Package schema provides GraphQL schema definitions.
package schema

import (
	"github.com/graphql-go/graphql"
)

// DeviceType represents a device in the system.
var DeviceType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Device",
	Description: "A registered device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique device identifier",
		},
		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "Device display name",
		},
		"online": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether device is currently connected",
		},
		"lastSeen": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Last time device sent telemetry",
		},
		"fcmToken": &graphql.Field{
			Type:        graphql.String,
			Description: "Firebase Cloud Messaging token",
		},
		"version": &graphql.Field{
			Type:        graphql.String,
			Description: "Device app version",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Device registration timestamp",
		},
	},
})

// TelemetryEntryType represents a telemetry data point.
var TelemetryEntryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TelemetryEntry",
	Description: "Telemetry data from a device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique entry identifier",
		},
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Device that sent this telemetry",
		},
		"receivedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When telemetry was received",
		},
		"riskScore": &graphql.Field{
			Type:        graphql.Int,
			Description: "Risk score (0-100)",
		},
		"bufferLevel": &graphql.Field{
			Type:        graphql.Int,
			Description: "Buffer level percentage",
		},
		"thermalTemp": &graphql.Field{
			Type:        graphql.Float,
			Description: "Device temperature in Celsius",
		},
		"payload": &graphql.Field{
			Type:        graphql.String,
			Description: "Raw telemetry payload",
		},
	},
})

// CommandType represents a command sent to a device.
var CommandType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Command",
	Description: "A command sent to a device",
	Fields: graphql.Fields{
		"dispatchId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique dispatch identifier",
		},
		"commandId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Internal command identifier",
		},
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target device",
		},
		"command": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Command name",
		},
		"args": &graphql.Field{
			Type:        JSONScalar,
			Description: "Command arguments",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(CommandStatusEnum),
			Description: "Current command status",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When command was created",
		},
		"deliveredAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When command was delivered",
		},
	},
})

// ConnectionStatusType represents a device's WebSocket connection status.
var ConnectionStatusType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ConnectionStatus",
	Description: "WebSocket connection status for a device",
	Fields: graphql.Fields{
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Device identifier",
		},
		"connected": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether device is connected",
		},
		"connectedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When connection was established",
		},
		"lastMessageAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Last message received from device",
		},
		"uptimeSeconds": &graphql.Field{
			Type:        graphql.Int,
			Description: "Connection uptime in seconds",
		},
	},
})

// CommandResultType represents the result of sending a command.
var CommandResultType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CommandResult",
	Description: "Result of sending a command to a device",
	Fields: graphql.Fields{
		"dispatchId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Dispatch identifier",
		},
		"commandId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Command identifier",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Delivery status: sent, queued, queued_fcm",
		},
		"deviceOnline": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether device was online at send time",
		},
	},
})

// TelemetryStatsType represents aggregated telemetry statistics.
var TelemetryStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TelemetryStats",
	Description: "Aggregated telemetry statistics",
	Fields: graphql.Fields{
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Device identifier",
		},
		"sampleCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of samples in calculation",
		},
		"riskScore": &graphql.Field{
			Type:        RiskScoreStatsType,
			Description: "Risk score statistics",
		},
		"bufferLevel": &graphql.Field{
			Type:        BufferLevelStatsType,
			Description: "Buffer level statistics",
		},
		"thermalTemp": &graphql.Field{
			Type:        ThermalTempStatsType,
			Description: "Temperature statistics",
		},
	},
})

// RiskScoreStatsType represents risk score statistics.
var RiskScoreStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "RiskScoreStats",
	Description: "Risk score statistics",
	Fields: (graphql.FieldsThunk)(func() graphql.Fields {
		return graphql.Fields{
			"avg": &graphql.Field{Type: graphql.Float},
			"min": &graphql.Field{Type: graphql.Int},
			"max": &graphql.Field{Type: graphql.Int},
		}
	}),
})

// BufferLevelStatsType represents buffer level statistics.
var BufferLevelStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "BufferLevelStats",
	Description: "Buffer level statistics",
	Fields: (graphql.FieldsThunk)(func() graphql.Fields {
		return graphql.Fields{
			"avg": &graphql.Field{Type: graphql.Float},
		}
	}),
})

// ThermalTempStatsType represents temperature statistics.
var ThermalTempStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ThermalTempStats",
	Description: "Temperature statistics",
	Fields: (graphql.FieldsThunk)(func() graphql.Fields {
		return graphql.Fields{
			"avg": &graphql.Field{Type: graphql.Float},
			"min": &graphql.Field{Type: graphql.Float},
			"max": &graphql.Field{Type: graphql.Float},
		}
	}),
})

// PaginationType represents pagination information.
var PaginationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Pagination",
	Description: "Pagination information for list queries",
	Fields: graphql.Fields{
		"total": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of items",
		},
		"limit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Items per page",
		},
		"offset": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Current offset",
		},
		"hasMore": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether more items exist",
		},
	},
})

// ============================================================
// Dashboard Commands & Logs Types
// ============================================================

// MetricChartPointType represents a single data point in a chart.
var MetricChartPointType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "MetricChartPoint",
	Description: "A single data point in a metric chart",
	Fields: graphql.Fields{
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Unix timestamp in milliseconds",
		},
		"value": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Metric value at this point",
		},
	},
})

// ThresholdType represents warning and critical thresholds.
var ThresholdType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Threshold",
	Description: "Warning and critical thresholds for a metric",
	Fields: graphql.Fields{
		"warning": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Warning threshold value",
		},
		"critical": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Critical threshold value",
		},
	},
})

// MetricDataType represents a complete metric with stats, chart data, and thresholds.
var MetricDataType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "MetricData",
	Description: "Complete metric data including current, average, min, max, chart and thresholds",
	Fields: graphql.Fields{
		"current": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Current metric value",
		},
		"avg": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Average metric value over the time range",
		},
		"min": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Minimum metric value",
		},
		"max": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Maximum metric value",
		},
		"unit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unit of measurement",
		},
		"chart": &graphql.Field{
			Type:        graphql.NewList(MetricChartPointType),
			Description: "Chart data points",
		},
		"threshold": &graphql.Field{
			Type:        ThresholdType,
			Description: "Warning and critical thresholds",
		},
	},
})

// MetricsCollectionType represents all metric types for a device.
var MetricsCollectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "MetricsCollection",
	Description: "Collection of all metrics for a device",
	Fields: graphql.Fields{
		"riskScore": &graphql.Field{
			Type:        graphql.NewNonNull(MetricDataType),
			Description: "Risk score metric (0-100%)",
		},
		"thermalTemp": &graphql.Field{
			Type:        graphql.NewNonNull(MetricDataType),
			Description: "Device temperature in Celsius",
		},
		"bufferLevel": &graphql.Field{
			Type:        graphql.NewNonNull(MetricDataType),
			Description: "Buffer level percentage",
		},
		"uptime": &graphql.Field{
			Type:        graphql.NewNonNull(MetricDataType),
			Description: "Device uptime in seconds",
		},
	},
})

// TimeRangeType represents the time range information.
var TimeRangeType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TimeRange",
	Description: "Time range information for metrics queries",
	Fields: graphql.Fields{
		"start": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Start timestamp in milliseconds",
		},
		"end": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "End timestamp in milliseconds",
		},
		"range": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Range preset (1h, 6h, 24h, 7d)",
		},
		"resolution": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Data resolution (1m, 5m, 15m, 1h)",
		},
	},
})

// ThresholdEventType represents an event when a threshold is breached.
var ThresholdEventType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ThresholdEvent",
	Description: "Event when a metric exceeds a threshold",
	Fields: graphql.Fields{
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When the threshold was breached",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Event type (threshold_breach)",
		},
		"metric": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Which metric was breached",
		},
		"value": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Value at time of breach",
		},
		"threshold": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Float),
			Description: "Threshold that was breached",
		},
	},
})

// DeviceMetricsType represents aggregated metrics for a device.
var DeviceMetricsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceMetrics",
	Description: "Complete metrics data for a device",
	Fields: graphql.Fields{
		"device": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
				Name:   "DeviceInfo",
				Fields: graphql.Fields{
					"imei":       &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
					"deviceName": &graphql.Field{Type: graphql.String},
				},
			})),
			Description: "Device identification",
		},
		"timeRange": &graphql.Field{
			Type:        graphql.NewNonNull(TimeRangeType),
			Description: "Query time range and resolution",
		},
		"metrics": &graphql.Field{
			Type:        graphql.NewNonNull(MetricsCollectionType),
			Description: "All metric data",
		},
		"events": &graphql.Field{
			Type:        graphql.NewList(ThresholdEventType),
			Description: "Threshold breach events",
		},
	},
})

// LogEntryType represents a device log entry.
var LogEntryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "LogEntry",
	Description: "A log entry from a device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique log entry identifier",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Log type (connection, command, telemetry, error, warning)",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Unix timestamp in milliseconds",
		},
		"data": &graphql.Field{
			Type:        JSONScalar,
			Description: "Additional log data as JSON",
		},
	},
})

// LogConnectionType represents a paginated connection of log entries.
var LogConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "LogConnection",
	Description: "Paginated log entries with cursor-based pagination",
	Fields: graphql.Fields{
		"events": &graphql.Field{
			Type:        graphql.NewList(LogEntryType),
			Description: "Log entries",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(LogPaginationType),
			Description: "Pagination information",
		},
	},
})

// LogPaginationType represents cursor-based pagination for logs.
var LogPaginationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "LogPagination",
	Description: "Cursor-based pagination for logs",
	Fields: graphql.Fields{
		"limit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Items per page",
		},
		"hasMore": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether more entries exist",
		},
		"nextCursor": &graphql.Field{
			Type:        graphql.String,
			Description: "Cursor for next page",
		},
	},
})

// CommandHistoryType represents paginated command history.
var CommandHistoryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CommandHistory",
	Description: "Paginated command history for a device",
	Fields: graphql.Fields{
		"commands": &graphql.Field{
			Type:        graphql.NewList(CommandType),
			Description: "Command entries",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(CommandPaginationType),
			Description: "Pagination information",
		},
	},
})

// CommandPaginationType represents pagination for command history.
var CommandPaginationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CommandPagination",
	Description: "Pagination for command history",
	Fields: graphql.Fields{
		"page": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Current page number",
		},
		"limit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Items per page",
		},
		"total": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of commands",
		},
		"totalPages": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of pages",
		},
		"hasMore": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether more commands exist",
		},
	},
})

// DashboardStatsType represents aggregated dashboard statistics.
var DashboardStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DashboardStats",
	Description: "Aggregated statistics for the dashboard",
	Fields: graphql.Fields{
		"devices": &graphql.Field{
			Type:        graphql.NewNonNull(DevicesStatsType),
			Description: "Device statistics",
		},
		"commands": &graphql.Field{
			Type:        graphql.NewNonNull(CommandsStatsType),
			Description: "Command statistics",
		},
		"activity": &graphql.Field{
			Type:        graphql.NewNonNull(ActivityStatsType),
			Description: "Activity in the last 24 hours",
		},
	},
})

// DevicesStatsType represents device statistics.
var DevicesStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DevicesStats",
	Description: "Statistics about devices",
	Fields: graphql.Fields{
		"total": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of devices",
		},
		"online": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of online devices",
		},
		"offline": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of offline devices",
		},
	},
})

// CommandsStatsType represents command statistics.
var CommandsStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CommandsStats",
	Description: "Statistics about commands",
	Fields: graphql.Fields{
		"totalToday": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total commands sent today",
		},
		"pending": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of pending commands",
		},
		"failed": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of failed commands",
		},
	},
})

// ActivityDetailType represents detailed activity.
var ActivityDetailType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ActivityDetail",
	Description: "Detailed activity information",
	Fields: graphql.Fields{
		"commands": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Commands in the period",
		},
		"registrations": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Device registrations in the period",
		},
		"deregistrations": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Device deregistrations in the period",
		},
	},
})

// ActivityStatsType represents activity in the last 24 hours.
var ActivityStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ActivityStats",
	Description: "Activity statistics",
	Fields: graphql.Fields{
		"last24h": &graphql.Field{
			Type:        graphql.NewNonNull(ActivityDetailType),
			Description: "Activity in the last 24 hours",
		},
	},
})
