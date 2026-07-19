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

// IdentityInfoType represents device identity information.
var IdentityInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "IdentityInfo",
	Description: "Device identity information",
	Fields: graphql.Fields{
		"imei": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Device IMEI",
		},
		"deviceName": &graphql.Field{
			Type:        graphql.String,
			Description: "Device display name",
		},
		"model": &graphql.Field{
			Type:        graphql.String,
			Description: "Device model",
		},
		"manufacturer": &graphql.Field{
			Type:        graphql.String,
			Description: "Device manufacturer",
		},
	},
})

// SoftwareInfoType represents device software information.
var SoftwareInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SoftwareInfo",
	Description: "Device software information",
	Fields: graphql.Fields{
		"osVersion": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Operating system version",
		},
		"appVersion": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Application version",
		},
		"securityPatch": &graphql.Field{
			Type:        graphql.String,
			Description: "Security patch level",
		},
		"buildId": &graphql.Field{
			Type:        graphql.String,
			Description: "Build identifier",
		},
	},
})

// RegistrationInfoType represents device registration information.
var RegistrationInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "RegistrationInfo",
	Description: "Device registration information",
	Fields: graphql.Fields{
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(DeviceStatusEnum),
			Description: "Device registration status",
		},
		"registeredAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Registration timestamp",
		},
		"fcmTokenValid": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether FCM token is valid",
		},
		"fcmTokenRefreshedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "FCM token last refresh timestamp",
		},
		"commandSecretSet": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether command secret is configured",
		},
	},
})

// ConnectionInfoType represents device connection information.
var ConnectionInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ConnectionInfo",
	Description: "Device connection information",
	Fields: graphql.Fields{
		"webSocketStatus": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "WebSocket connection status",
		},
		"connectedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "WebSocket connection timestamp",
		},
		"fcmStatus": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "FCM status",
		},
		"lastSeen": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Last seen timestamp",
		},
		"clientIp": &graphql.Field{
			Type:        graphql.String,
			Description: "Client IP address",
		},
		"protocol": &graphql.Field{
			Type:        graphql.String,
			Description: "Connection protocol",
		},
	},
})

// TelemetryInfoType represents device telemetry statistics.
var TelemetryInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TelemetryInfo",
	Description: "Device telemetry statistics",
	Fields: graphql.Fields{
		"lastTimestamp": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "Last telemetry timestamp",
		},
		"framesToday": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of frames received today",
		},
		"avgLatencyMs": &graphql.Field{
			Type:        graphql.Int,
			Description: "Average WebSocket latency in milliseconds",
		},
		"totalBytesToday": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total bytes transferred today",
		},
		"sessionsToday": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Number of sessions today",
		},
	},
})

// DeviceInspectionType represents full device inspection data.
var DeviceInspectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceInspection",
	Description: "Full device inspection data for diagnostics",
	Fields: graphql.Fields{
		"identity": &graphql.Field{
			Type:        graphql.NewNonNull(IdentityInfoType),
			Description: "Device identity information",
		},
		"software": &graphql.Field{
			Type:        graphql.NewNonNull(SoftwareInfoType),
			Description: "Device software information",
		},
		"registration": &graphql.Field{
			Type:        graphql.NewNonNull(RegistrationInfoType),
			Description: "Device registration information",
		},
		"connection": &graphql.Field{
			Type:        graphql.NewNonNull(ConnectionInfoType),
			Description: "Device connection information",
		},
		"telemetry": &graphql.Field{
			Type:        graphql.NewNonNull(TelemetryInfoType),
			Description: "Device telemetry statistics",
		},
	},
})

// TimelineEventType represents a single event in the device timeline.
var TimelineEventType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TimelineEvent",
	Description: "A timeline event for a device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Event unique identifier",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(TimelineEventTypeEnum),
			Description: "Event type",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "Event timestamp",
		},
		"data": &graphql.Field{
			Type:        JSONScalar,
			Description: "Additional event data",
		},
	},
})

// TimelineConnectionType represents paginated timeline results.
var TimelineConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TimelineConnection",
	Description: "Paginated timeline connection",
	Fields: graphql.Fields{
		"events": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(TimelineEventType))),
			Description: "Timeline events",
		},
		"hasMore": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether there are more results",
		},
		"nextCursor": &graphql.Field{
			Type:        graphql.String,
			Description: "Cursor for next page",
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

// ============================================================
// Updates API Types
// ============================================================

// UpdateVersionType represents an update version in the GraphQL schema.
var UpdateVersionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UpdateVersion",
	Description: "An available update version",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique version identifier",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version string (e.g. v1.2.0)",
		},
		"releaseType": &graphql.Field{
			Type:        graphql.NewNonNull(ReleaseTypeEnum),
			Description: "Release type (MAJOR, MINOR, PATCH)",
		},
		"releaseNotes": &graphql.Field{
			Type:        graphql.String,
			Description: "Release notes for this version",
		},
		"apkFilename": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "APK filename",
		},
		"apkSize": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "APK size in bytes",
		},
		"sha256": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "SHA256 hash of the APK",
		},
		"releasedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Release timestamp",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the version was added to the system",
		},
		"isLatest": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Whether this is the latest version",
		},
	},
})

// DeviceUpdateStatusType represents the device's update status.
var DeviceUpdateStatusType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceUpdateStatus",
	Description: "Device current version and update status",
	Fields: graphql.Fields{
		"currentVersion": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Current version installed on device",
		},
		"needsUpdate": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether device needs an update",
		},
	},
})

// UpdateVersionListType represents a paginated list of update versions.
var UpdateVersionListType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UpdateVersionList",
	Description: "Paginated list of update versions",
	Fields: graphql.Fields{
		"versions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(UpdateVersionType))),
			Description: "List of versions",
		},
		"pagination": &graphql.Field{
			Type:        PaginationType,
			Description: "Pagination information",
		},
	},
})

// CancelCommandResponseType represents the response from cancelling a command.
var CancelCommandResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CancelCommandResponse",
	Description: "Response when cancelling a command",
	Fields: graphql.Fields{
		"dispatchId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The dispatch ID of the cancelled command",
		},
		"cancelledAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Timestamp when the command was cancelled",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Status of the cancellation",
		},
	},
})

// ChangelogEntryType represents a changelog entry.
var ChangelogEntryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ChangelogEntry",
	Description: "A changelog entry for a release",
	Fields: graphql.Fields{
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version this changelog entry belongs to",
		},
		"date": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Release date",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Change type (added, changed, fixed, removed)",
		},
		"notes": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Change notes",
		},
	},
})

// SyncStatusType represents the GitHub sync status.
var SyncStatusType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SyncStatus",
	Description: "GitHub sync status information",
	Fields: graphql.Fields{
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Sync status: idle, syncing, synced, error",
		},
		"lastSyncAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the last sync completed",
		},
		"nextSyncAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the next scheduled sync will run",
		},
		"versionsFound": &graphql.Field{
			Type:        graphql.Int,
			Description: "Number of versions found in last sync",
		},
		"error": &graphql.Field{
			Type:        graphql.String,
			Description: "Error message if sync failed",
		},
	},
})

// UpdateStatusType represents the overall update system status.
var UpdateStatusType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UpdateStatus",
	Description: "Overall update system status",
	Fields: graphql.Fields{
		"sync": &graphql.Field{
			Type:        graphql.NewNonNull(SyncStatusType),
			Description: "GitHub sync status",
		},
		"latest": &graphql.Field{
			Type:        UpdateVersionType,
			Description: "Latest available version",
		},
		"device": &graphql.Field{
			Type:        graphql.NewNonNull(DeviceUpdateStatusType),
			Description: "Device current version and update status",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Device current version (alias for device)",
		},
		"apkFilename": &graphql.Field{
			Type:        graphql.String,
			Description: "APK filename for the update",
		},
		"sha256": &graphql.Field{
			Type:        graphql.String,
			Description: "SHA256 hash for the update",
		},
	},
})

// PushDeviceType represents a device in an update push.
var PushDeviceType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "PushDevice",
	Description: "A device included in an update push",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Push device entry ID",
		},
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Device identifier",
		},
		"deviceName": &graphql.Field{
			Type:        graphql.String,
			Description: "Device name (if available)",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(DevicePushStatusEnum),
			Description: "Push status for this device",
		},
		"sentAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the update was sent to this device",
		},
		"acknowledgedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the device acknowledged the update",
		},
		"error": &graphql.Field{
			Type:        graphql.String,
			Description: "Error message if push failed for this device",
		},
	},
})

// UpdatePushType represents an update push.
var UpdatePushType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UpdatePush",
	Description: "An update push to devices",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique push identifier",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version being pushed",
		},
		"installType": &graphql.Field{
			Type:        graphql.NewNonNull(InstallTypeEnum),
			Description: "Install type (immediate or scheduled)",
		},
		"scheduledAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push is scheduled (null for immediate)",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(UpdateStatusEnum),
			Description: "Push status",
		},
		"initiatedBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Operator ID who initiated the push",
		},
		"initiatedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push was initiated",
		},
		"completedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push completed",
		},
		"cancelledAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push was cancelled",
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of devices in this push",
		},
		"devices": &graphql.Field{
			Type:        graphql.NewList(graphql.NewNonNull(PushDeviceType)),
			Description: "Devices in this push",
		},
	},
})

// PushHistoryEntryType represents a single push history entry.
var PushHistoryEntryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "PushHistoryEntry",
	Description: "A single push history entry",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Push identifier",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version pushed",
		},
		"installType": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Install type",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Push status",
		},
		"initiatedBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Operator who initiated",
		},
		"initiatedAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When initiated (Unix ms)",
		},
		"completedAt": &graphql.Field{
			Type:        graphql.Int,
			Description: "When completed (Unix ms)",
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total device count",
		},
		"pending": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Devices still pending",
		},
		"acknowledged": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Devices that acknowledged",
		},
		"failed": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Devices that failed",
		},
	},
})

// PushHistoryConnectionType represents paginated push history.
var PushHistoryConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "PushHistoryConnection",
	Description: "Paginated push history",
	Fields: graphql.Fields{
		"pushes": &graphql.Field{
			Type:        graphql.NewList(graphql.NewNonNull(PushHistoryEntryType)),
			Description: "Push history entries",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(PaginationType),
			Description: "Pagination info",
		},
	},
})

// UpdateHistoryType represents a single update push history record.
var UpdateHistoryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UpdateHistory",
	Description: "An update push history record",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique push identifier",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version that was pushed",
		},
		"installType": &graphql.Field{
			Type:        graphql.NewNonNull(InstallTypeEnum),
			Description: "Install type (immediate or scheduled)",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(UpdateStatusEnum),
			Description: "Push status",
		},
		"initiatedBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Operator who initiated the push",
		},
		"initiatedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push was initiated",
		},
		"completedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push completed",
		},
		"cancelledAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the push was cancelled",
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total devices in this push",
		},
		"devices": &graphql.Field{
			Type:        graphql.NewList(graphql.NewNonNull(PushDeviceType)),
			Description: "Devices in this push",
		},
	},
})

// PushUpdateResponseType represents the response from pushing an update.
var PushUpdateResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "PushUpdateResponse",
	Description: "Response from pushing an update",
	Fields: graphql.Fields{
		"pushId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Created push ID",
		},
		"version": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Version being pushed",
		},
		"installType": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Install type",
		},
		"scheduledAt": &graphql.Field{
			Type:        graphql.Int,
			Description: "When scheduled (Unix ms, null for immediate)",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Push status",
		},
		"initiatedBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Operator who initiated",
		},
		"initiatedAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When initiated (Unix ms)",
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total devices",
		},
	},
})

// SyncResponseType represents the response from triggering a sync.
var SyncResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SyncResponse",
	Description: "Response from triggering a GitHub sync",
	Fields: graphql.Fields{
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Sync status",
		},
		"startedAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When sync started (Unix ms)",
		},
		"message": &graphql.Field{
			Type:        graphql.String,
			Description: "Status message",
		},
		"versionsFound": &graphql.Field{
			Type:        graphql.Int,
			Description: "Versions found in this sync",
		},
	},
})

// CancelPushResponseType represents the response from cancelling a push.
var CancelPushResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CancelPushResponse",
	Description: "Response from cancelling a push",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Push ID",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "New status (cancelled)",
		},
		"cancelledAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When cancelled (Unix ms)",
		},
		"cancelledBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Who cancelled",
		},
	},
})

// InboxEntryType represents a device registration request in the inbox.
var InboxEntryType = graphql.NewObject(graphql.ObjectConfig{
Name:        "InboxEntry",
Description: "A device registration request",
Fields: graphql.Fields{
"id": &graphql.Field{
Type:        graphql.NewNonNull(graphql.ID),
Description: "Unique inbox entry ID",
},
"imei": &graphql.Field{
Type:        graphql.NewNonNull(graphql.String),
Description: "Device IMEI",
},
"model": &graphql.Field{
Type:        graphql.String,
Description: "Device model",
},
"manufacturer": &graphql.Field{
Type:        graphql.String,
Description: "Device manufacturer",
},
"osVersion": &graphql.Field{
Type:        graphql.String,
Description: "Device OS version",
},
"appVersion": &graphql.Field{
Type:        graphql.String,
Description: "App version",
},
"firebaseInstallId": &graphql.Field{
Type:        graphql.String,
Description: "Firebase Install ID",
},
"status": &graphql.Field{
Type:        graphql.NewNonNull(graphql.String),
Description: "Status: pending, approved, rejected",
},
"notes": &graphql.Field{
Type:        graphql.String,
Description: "Operator notes",
},
"operatorId": &graphql.Field{
Type:        graphql.String,
Description: "Handling operator ID",
},
"createdAt": &graphql.Field{
Type:        graphql.NewNonNull(graphql.Int),
Description: "When created (Unix ms)",
},
"approvedAt": &graphql.Field{
Type:        graphql.Int,
Description: "When approved (Unix ms)",
},
"rejectedAt": &graphql.Field{
Type:        graphql.Int,
Description: "When rejected (Unix ms)",
},
},
})

// InboxListResponseType represents a paginated inbox list response.
var InboxListResponseType = graphql.NewObject(graphql.ObjectConfig{
Name:        "InboxListResponse",
Description: "Paginated inbox list response",
Fields: graphql.Fields{
"requests": &graphql.Field{
Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(InboxEntryType))),
Description: "Inbox entries",
},
"pagination": &graphql.Field{
Type:        graphql.NewNonNull(PaginationType),
Description: "Pagination info",
},
},
})

// RegistrationLogType represents an audit log entry for registration actions.
var RegistrationLogType = graphql.NewObject(graphql.ObjectConfig{
Name:        "RegistrationLog",
Description: "Audit log entry for registration actions",
Fields: graphql.Fields{
"id": &graphql.Field{
Type:        graphql.NewNonNull(graphql.ID),
Description: "Log entry ID",
},
"deviceId": &graphql.Field{
Type:        graphql.String,
Description: "Device ID",
},
"imei": &graphql.Field{
Type:        graphql.String,
Description: "Device IMEI",
},
"action": &graphql.Field{
Type:        graphql.NewNonNull(graphql.String),
Description: "Action type",
},
"operatorId": &graphql.Field{
Type:        graphql.String,
Description: "Operator ID",
},
"clientIp": &graphql.Field{
Type:        graphql.String,
Description: "Client IP address",
},
"userAgent": &graphql.Field{
Type:        graphql.String,
Description: "User agent string",
},
"details": &graphql.Field{
Type:        graphql.String,
Description: "Additional details",
},
"timestamp": &graphql.Field{
Type:        graphql.NewNonNull(graphql.Int),
Description: "When the action occurred (Unix ms)",
},
},
})

// DeviceConnectionType represents WebSocket connection info for a device.
var DeviceConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceConnection",
	Description: "WebSocket connection information for a device",
	Fields: graphql.Fields{
		"webSocketStatus": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "WebSocket connection status",
		},
		"connectedAt": &graphql.Field{
			Type:        graphql.Int,
			Description: "When connection was established (Unix ms)",
		},
		"protocol": &graphql.Field{
			Type:        graphql.String,
			Description: "Connection protocol (WSS/WEBSOCKET)",
		},
		"clientIp": &graphql.Field{
			Type:        graphql.String,
			Description: "Client IP address",
		},
	},
})

// DeviceListConnectionType represents a paginated list of devices.
var DeviceListConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceListConnection",
	Description: "Paginated list of devices",
	Fields: graphql.Fields{
		"devices": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "List of devices",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(PaginationType),
			Description: "Pagination information",
		},
	},
})

// AckResultType represents the result of acknowledging an inbox entry.
var AckResultType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "AckResult",
	Description: "Result of acknowledging an inbox entry",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Inbox entry ID",
		},
		"imei": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Device IMEI",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(InboxStatusEnum),
			Description: "New status after acknowledgement",
		},
		"approvedAt": &graphql.Field{
			Type:        graphql.Int,
			Description: "When approved (Unix ms)",
		},
		"rejectedAt": &graphql.Field{
			Type:        graphql.Int,
			Description: "When rejected (Unix ms)",
		},
		"commandSecret": &graphql.Field{
			Type:        graphql.String,
			Description: "Generated command secret (if approved)",
		},
		"fcmPushSent": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Whether FCM push was sent",
		},
		"notes": &graphql.Field{
			Type:        graphql.String,
			Description: "Operator notes",
		},
	},
})

// DeregisterResultType represents the result of deregistering a device.
var DeregisterResultType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeregisterResult",
	Description: "Result of deregistering a device",
	Fields: graphql.Fields{
		"imei": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Device IMEI",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Deregistration status",
		},
		"deregisteredAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When deregistered (Unix ms)",
		},
		"retentionUntil": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "When device data will be permanently deleted (Unix ms)",
		},
	},
})

// ClientSettingsType represents client settings.
var ClientSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ClientSettings",
	Description: "Client settings for dashboard behavior",
	Fields: graphql.Fields{
		"serverUrl": &graphql.Field{
			Type:        graphql.String,
			Description: "Server URL for device connection",
		},
		"deviceId": &graphql.Field{
			Type:        graphql.String,
			Description: "Default device ID",
		},
		"requestTimeoutMs": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Request timeout in milliseconds",
		},
		"autoReconnect": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Auto reconnect on disconnect",
		},
		"strictHmac": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Require strict HMAC validation",
		},
		"logBufferLimit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Maximum log buffer size",
		},
		"signalHistoryLimit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Signal history retention limit",
		},
	},
})

// ThresholdsType represents alert thresholds.
var ThresholdsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Thresholds",
	Description: "Alert thresholds for device telemetry",
	Fields: graphql.Fields{
		"riskWarn": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Risk warning threshold (0-100)",
		},
		"riskCrit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Risk critical threshold (0-100)",
		},
		"thermalWarn": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Thermal warning threshold (0-100)",
		},
		"thermalCrit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Thermal critical threshold (0-100)",
		},
		"bufferWarn": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Buffer warning threshold (0-100)",
		},
		"bufferCrit": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Buffer critical threshold (0-100)",
		},
	},
})

// NotificationTypesType represents notification type flags.
var NotificationTypesType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "NotificationTypes",
	Description: "Notification type preferences",
	Fields: graphql.Fields{
		"thresholdBreach": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify on threshold breach",
		},
		"deviceOffline": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify when device goes offline",
		},
		"deviceOnline": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify when device comes online",
		},
		"updateAvailable": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify when update is available",
		},
		"commandFailed": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify when command fails",
		},
		"registrationRequest": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Notify on registration requests",
		},
	},
})

// WebhookSettingsType represents webhook notification settings.
var WebhookSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "WebhookSettings",
	Description: "Webhook notification configuration",
	Fields: graphql.Fields{
		"enabled": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether webhook is enabled",
		},
		"url": &graphql.Field{
			Type:        graphql.String,
			Description: "Webhook URL",
		},
		"secret": &graphql.Field{
			Type:        graphql.String,
			Description: "Webhook secret (hidden)",
		},
		"types": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "Notification types to send",
		},
	},
})

// NotificationSettingsType represents notification settings.
var NotificationSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "NotificationSettings",
	Description: "Notification preferences",
	Fields: graphql.Fields{
		"enabled": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether notifications are enabled",
		},
		"channels": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "Enabled notification channels",
		},
		"email": &graphql.Field{
			Type:        graphql.NewNonNull(NotificationTypesType),
			Description: "Email notification preferences",
		},
		"push": &graphql.Field{
			Type:        graphql.NewNonNull(NotificationTypesType),
			Description: "Push notification preferences",
		},
		"webhook": &graphql.Field{
			Type:        graphql.NewNonNull(WebhookSettingsType),
			Description: "Webhook notification settings",
		},
	},
})

// OperatorSettingsType represents all operator settings.
var OperatorSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OperatorSettings",
	Description: "Complete operator settings",
	Fields: graphql.Fields{
		"client": &graphql.Field{
			Type:        graphql.NewNonNull(ClientSettingsType),
			Description: "Client settings",
		},
		"notifications": &graphql.Field{
			Type:        graphql.NewNonNull(NotificationSettingsType),
			Description: "Notification settings",
		},
	},
})

// DeviceSettingsType represents settings for a specific device.
var DeviceSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DeviceSettings",
	Description: "Settings for a device including custom name, location, and threshold overrides",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique settings identifier",
		},
		"deviceImei": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Device IMEI",
		},
		"customName": &graphql.Field{
			Type:        graphql.String,
			Description: "Custom display name for the device",
		},
		"location": &graphql.Field{
			Type:        graphql.String,
			Description: "Device location description",
		},
		"metadata": &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
				Name: "MetadataEntry",
				Fields: graphql.Fields{
					"key": &graphql.Field{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Metadata key",
					},
					"value": &graphql.Field{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Metadata value",
					},
				},
			}))),
			Description: "Custom metadata key-value pairs",
		},
		"thresholds": &graphql.Field{
			Type:        ThresholdsType,
			Description: "Device-specific threshold overrides (null = use org defaults)",
		},
		"effectiveThresholds": &graphql.Field{
			Type:        graphql.NewNonNull(ThresholdsType),
			Description: "Effective thresholds after applying hierarchy (device → org → default)",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When settings were created",
		},
		"updatedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When settings were last updated",
		},
	},
})

// OrganizationSettingsType represents settings for an organization.
var OrganizationSettingsType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OrganizationSettings",
	Description: "Settings for an organization including timezone and default thresholds",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique settings identifier",
		},
		"organizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Organization ID",
		},
		"timezone": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Organization timezone (e.g., UTC, America/New_York)",
		},
		"dateFormat": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Date format (e.g., YYYY-MM-DD)",
		},
		"alertCooldownMinutes": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Minimum minutes between repeated alerts",
		},
		"defaultThresholds": &graphql.Field{
			Type:        graphql.NewNonNull(ThresholdsType),
			Description: "Default thresholds for devices without custom thresholds",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When settings were created",
		},
		"updatedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When settings were last updated",
		},
	},
})

// WebhookTestResultType represents the result of testing a webhook.
var WebhookTestResultType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "WebhookTestResult",
	Description: "Result of testing a webhook endpoint",
	Fields: graphql.Fields{
		"success": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether the webhook test was successful",
		},
		"statusCode": &graphql.Field{
			Type:        graphql.Int,
			Description: "HTTP status code from webhook",
		},
		"responseTime": &graphql.Field{
			Type:        graphql.Int,
			Description: "Response time in milliseconds",
		},
		"error": &graphql.Field{
			Type:        graphql.String,
			Description: "Error message if test failed",
		},
	},
})

// OrganizationType represents an organization in the system.
var OrganizationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Organization",
	Description: "An organization for multi-tenant device management",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique organization identifier",
		},
		"name": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Organization name",
		},
		"lifecycle": &graphql.Field{
			Type:        graphql.NewNonNull(OrganizationLifecycleEnum),
			Description: "Organization lifecycle state",
		},
		"maxMembers": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Maximum number of members allowed",
		},
		"memberCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Current number of active members",
		},
		"createdAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Organization creation timestamp",
		},
		"updatedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Organization last update timestamp",
		},
		"deletedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "Organization deletion timestamp (soft delete)",
		},
		"createdBy": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "ID of the operator who created this organization",
		},
	},
})

// MembershipType represents an organization membership.
var MembershipType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Membership",
	Description: "An organization membership",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique membership identifier",
		},
		"organizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Organization ID",
		},
		"operatorId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Operator ID",
		},
		"role": &graphql.Field{
			Type:        graphql.NewNonNull(OrgRoleEnum),
			Description: "Member role in the organization",
		},
		"lifecycle": &graphql.Field{
			Type:        graphql.NewNonNull(MemberLifecycleEnum),
			Description: "Membership lifecycle state",
		},
		"invitedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the invitation was sent",
		},
		"joinedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the operator joined",
		},
		"removedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the member was removed",
		},
		"suspendedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the member was suspended",
		},
		"operator": &graphql.Field{
			Type:        OperatorType,
			Description: "The operator associated with this membership",
		},
		"organization": &graphql.Field{
			Type:        OrganizationType,
			Description: "The organization",
		},
	},
})

// InvitationType represents an organization invitation.
var InvitationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Invitation",
	Description: "An organization invitation",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique invitation identifier",
		},
		"organizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Organization ID",
		},
		"organizationName": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Organization name at time of invitation",
		},
		"email": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Email address the invitation was sent to",
		},
		"role": &graphql.Field{
			Type:        graphql.NewNonNull(OrgRoleEnum),
			Description: "Role being invited to",
		},
		"status": &graphql.Field{
			Type:        graphql.NewNonNull(InvitationStatusEnum),
			Description: "Invitation status",
		},
		"token": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unique invitation token",
		},
		"inviterId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "ID of the operator who sent the invitation",
		},
		"inviterName": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Name of the inviter",
		},
		"inviteeId": &graphql.Field{
			Type:        graphql.ID,
			Description: "ID of the invited operator (if accepted)",
		},
		"inviteeNotes": &graphql.Field{
			Type:        graphql.String,
			Description: "Notes from the invitee",
		},
		"inviterNotes": &graphql.Field{
			Type:        graphql.String,
			Description: "Notes from the inviter",
		},
		"createdAt": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "When the invitation was created",
		},
		"expiresAt": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "When the invitation expires",
		},
		"respondedAt": &graphql.Field{
			Type:        DateTimeScalar,
			Description: "When the invitation was responded to",
		},
		"organization": &graphql.Field{
			Type:        OrganizationType,
			Description: "The organization",
		},
	},
})

// OrganizationMemberType represents a member within an organization context.
var OrganizationMemberType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OrganizationMember",
	Description: "A member of an organization with their membership details",
	Fields: graphql.Fields{
		"membership": &graphql.Field{
			Type:        graphql.NewNonNull(MembershipType),
			Description: "Membership details",
		},
		"operator": &graphql.Field{
			Type:        OperatorType,
			Description: "Operator information",
		},
	},
})

// CreateOrganizationPayloadType represents the result of creating an organization.
var CreateOrganizationPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CreateOrganizationPayload",
	Description: "Result of creating an organization",
	Fields: graphql.Fields{
		"organization": &graphql.Field{
			Type:        graphql.NewNonNull(OrganizationType),
			Description: "The created organization",
		},
		"membership": &graphql.Field{
			Type:        graphql.NewNonNull(MembershipType),
			Description: "The creator's membership as super_admin",
		},
	},
})

// TransferDevicePayloadType represents the result of transferring a device.
var TransferDevicePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TransferDevicePayload",
	Description: "Result of transferring a device between organizations",
	Fields: graphql.Fields{
		"success": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether the transfer was successful",
		},
		"deviceId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The transferred device ID",
		},
		"sourceOrganizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Source organization ID",
		},
		"targetOrganizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target organization ID",
		},
	},
})

// PaginationType represents pagination metadata.
var PaginationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Pagination",
	Description: "Pagination metadata for list queries",
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
			Description: "Total number of items",
		},
		"totalPages": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total number of pages",
		},
		"hasMore": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether there are more pages",
		},
	},
})

// OrganizationListResponseType represents a paginated list of organizations.
var OrganizationListResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OrganizationListResponse",
	Description: "Paginated list of organizations",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(OrganizationType))),
			Description: "List of organizations",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(PaginationType),
			Description: "Pagination metadata",
		},
	},
})

// MemberListResponseType represents a paginated list of organization members.
var MemberListResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "MemberListResponse",
	Description: "Paginated list of organization members",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(MembershipType))),
			Description: "List of members",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(PaginationType),
			Description: "Pagination metadata",
		},
	},
})

// InvitationListResponseType represents a paginated list of invitations.
var InvitationListResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "InvitationListResponse",
	Description: "Paginated list of invitations",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(InvitationType))),
			Description: "List of invitations",
		},
		"pagination": &graphql.Field{
			Type:        graphql.NewNonNull(PaginationType),
			Description: "Pagination metadata",
		},
	},
})

// OrganizationEventTypeEnum represents the type of organization event.
var OrganizationEventTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "OrganizationEventType",
	Description: "Type of organization event",
	Values: graphql.EnumValueConfigMap{
		"CREATED": &graphql.EnumValueConfig{
			Value:       "created",
			Description: "Organization was created",
		},
		"UPDATED": &graphql.EnumValueConfig{
			Value:       "updated",
			Description: "Organization was updated",
		},
		"DELETED": &graphql.EnumValueConfig{
			Value:       "deleted",
			Description: "Organization was deleted",
		},
		"ACTIVATED": &graphql.EnumValueConfig{
			Value:       "activated",
			Description: "Organization was activated",
		},
		"DEACTIVATED": &graphql.EnumValueConfig{
			Value:       "deactivated",
			Description: "Organization was deactivated",
		},
	},
})

// MemberEventTypeEnum represents the type of member event.
var MemberEventTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "MemberEventType",
	Description: "Type of member event",
	Values: graphql.EnumValueConfigMap{
		"MEMBER_JOINED": &graphql.EnumValueConfig{
			Value:       "member_joined",
			Description: "A member joined the organization",
		},
		"MEMBER_INVITED": &graphql.EnumValueConfig{
			Value:       "member_invited",
			Description: "A member was invited",
		},
		"MEMBER_REMOVED": &graphql.EnumValueConfig{
			Value:       "member_removed",
			Description: "A member was removed",
		},
		"MEMBER_SUSPENDED": &graphql.EnumValueConfig{
			Value:       "member_suspended",
			Description: "A member was suspended",
		},
		"MEMBER_REACTIVATED": &graphql.EnumValueConfig{
			Value:       "member_reactivated",
			Description: "A member was reactivated",
		},
		"ROLE_CHANGED": &graphql.EnumValueConfig{
			Value:       "role_changed",
			Description: "A member's role was changed",
		},
	},
})

// OrganizationEventType represents an organization event.
var OrganizationEventType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OrganizationEvent",
	Description: "An event related to an organization",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique event identifier",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(OrganizationEventTypeEnum),
			Description: "Event type",
		},
		"organizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Organization ID",
		},
		"operatorId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Operator who triggered the event",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "Event timestamp",
		},
		"data": &graphql.Field{
			Type:        JSONScalar,
			Description: "Additional event data",
		},
		"organization": &graphql.Field{
			Type:        OrganizationType,
			Description: "The organization",
		},
	},
})

// MemberEventType represents a member event.
var MemberEventType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "MemberEvent",
	Description: "An event related to an organization member",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Unique event identifier",
		},
		"type": &graphql.Field{
			Type:        graphql.NewNonNull(MemberEventTypeEnum),
			Description: "Event type",
		},
		"organizationId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Organization ID",
		},
		"memberId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Membership ID affected",
		},
		"operatorId": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Operator who triggered the event",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.NewNonNull(DateTimeScalar),
			Description: "Event timestamp",
		},
		"data": &graphql.Field{
			Type:        JSONScalar,
			Description: "Additional event data (e.g., old/new role)",
		},
		"membership": &graphql.Field{
			Type:        MembershipType,
			Description: "The membership",
		},
		"organization": &graphql.Field{
			Type:        OrganizationType,
			Description: "The organization",
		},
	},
})
