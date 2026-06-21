// Package schema provides GraphQL schema definitions.
package schema

import (
	"github.com/graphql-go/graphql"
)

// Define scalar types

// JSONScalar is a custom scalar for arbitrary JSON values.
var JSONScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSON",
	Description: "Arbitrary JSON value",
	Serialize: func(value interface{}) interface{} {
		return value
	},
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	ParseLiteral: func(valueAST interface{}) interface{} {
		return valueAST
	},
})

// DateTimeScalar is a custom scalar for ISO 8601 datetime.
var DateTimeScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "DateTime",
	Description: "ISO 8601 datetime string",
	Serialize: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		return value
	},
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	ParseLiteral: func(valueAST interface{}) interface{} {
		return valueAST
	},
})

// Define enum types

// CommandStatusEnum represents the status of a command.
var CommandStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "CommandStatus",
	Description: "Status of a device command",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: "pending",
			Description: "Command is pending delivery",
		},
		"DELIVERED": &graphql.EnumValueConfig{
			Value: "delivered",
			Description: "Command was delivered to device",
		},
		"FAILED": &graphql.EnumValueConfig{
			Value: "failed",
			Description: "Command delivery failed",
		},
		"CANCELLED": &graphql.EnumValueConfig{
			Value: "cancelled",
			Description: "Command was cancelled",
		},
	},
})

// Define object types

// DeviceType represents a device in the system.
var DeviceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Device",
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
	Name: "TelemetryEntry",
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
	Name: "Command",
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
	Name: "ConnectionStatus",
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
	Name: "CommandResult",
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
	Name: "TelemetryStats",
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
	Name: "RiskScoreStats",
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
	Name: "BufferLevelStats",
	Description: "Buffer level statistics",
	Fields: (graphql.FieldsThunk)(func() graphql.Fields {
		return graphql.Fields{
			"avg": &graphql.Field{Type: graphql.Float},
		}
	}),
})

// ThermalTempStatsType represents temperature statistics.
var ThermalTempStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ThermalTempStats",
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
	Name: "Pagination",
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