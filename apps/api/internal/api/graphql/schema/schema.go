// Package schema provides the GraphQL schema definition and configuration.
package schema

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/graphql-go/graphql"
)

// BuildSchema creates the complete GraphQL schema with the given resolver.
func BuildSchema(res *resolver.Resolver) (graphql.Schema, error) {
	queryType := buildQueryType(res)
	mutationType := buildMutationType(res)

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

// buildQueryType creates the Query type with all query fields.
func buildQueryType(res *resolver.Resolver) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: mergeFields(deviceQueries(res), commandQueries(res), telemetryQueries(res),
			connectionQueries(res), dashboardQueries(res)),
	})
}

func deviceQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"device": &graphql.Field{
			Type:        DeviceType,
			Description: "Get a single device by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetDevice,
		},
		"devices": &graphql.Field{
			Type:        graphql.NewList(DeviceType),
			Description: "List all devices for the authenticated operator",
			Args: graphql.FieldConfigArgument{
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
				"offset": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			},
			Resolve: res.GetDevices,
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.Int,
			Description: "Get total device count",
			Resolve:     res.GetDeviceCount,
		},
	}
}

func commandQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"command": &graphql.Field{
			Type:        CommandType,
			Description: "Get command status by dispatch ID",
			Args: graphql.FieldConfigArgument{
				"dispatchId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetCommand,
		},
		"pendingCommands": &graphql.Field{
			Type:        graphql.NewList(CommandType),
			Description: "Get pending commands for a device",
			Args: graphql.FieldConfigArgument{
				"deviceId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetPendingCommands,
		},
	}
}

func telemetryQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"telemetryHistory": &graphql.Field{
			Type:        graphql.NewList(TelemetryEntryType),
			Description: "Query telemetry history for a device",
			Args: graphql.FieldConfigArgument{
				"deviceId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"startTime":  &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"limit":      &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 100},
			},
			Resolve: res.GetTelemetryHistory,
		},
		"latestTelemetry": &graphql.Field{
			Type:        TelemetryEntryType,
			Description: "Get the latest telemetry entry for a device",
			Args: graphql.FieldConfigArgument{
				"deviceId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetLatestTelemetry,
		},
		"telemetryStats": &graphql.Field{
			Type:        TelemetryStatsType,
			Description: "Get telemetry statistics for a device",
			Args: graphql.FieldConfigArgument{
				"deviceId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetTelemetryStats,
		},
	}
}

func connectionQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"connectionStatus": &graphql.Field{
			Type:        ConnectionStatusType,
			Description: "Get connection status for a device",
			Args: graphql.FieldConfigArgument{
				"deviceId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetConnectionStatus,
		},
		"allConnections": &graphql.Field{
			Type:        graphql.NewList(ConnectionStatusType),
			Description: "Get all device connection statuses",
			Resolve:     res.GetAllConnections,
		},
	}
}

func dashboardQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"deviceMetrics": &graphql.Field{
			Type:        DeviceMetricsType,
			Description: "Get aggregated metrics for chart visualization",
			Args: graphql.FieldConfigArgument{
				"imei":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"range":      &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "6h"},
				"startTime":  &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"resolution": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: res.GetDeviceMetrics,
		},
		"deviceLogs": &graphql.Field{
			Type:        LogConnectionType,
			Description: "Get paginated device logs with cursor-based pagination",
			Args: graphql.FieldConfigArgument{
				"imei":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"type":      &graphql.ArgumentConfig{Type: graphql.String},
				"startTime": &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":   &graphql.ArgumentConfig{Type: graphql.Int},
				"limit":     &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 100},
				"cursor":    &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: res.GetDeviceLogs,
		},
		"deviceCommandHistory": &graphql.Field{
			Type:        CommandHistoryType,
			Description: "Get paginated command history for a device",
			Args: graphql.FieldConfigArgument{
				"imei":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"status":    &graphql.ArgumentConfig{Type: graphql.String},
				"page":      &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":     &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
				"startTime": &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":   &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: res.GetDeviceCommandHistory,
		},
		"dashboardStats": &graphql.Field{
			Type:        DashboardStatsType,
			Description: "Get aggregated dashboard statistics",
			Resolve:     res.GetDashboardStats,
		},
	}
}

func mergeFields(maps ...graphql.Fields) graphql.Fields {
	result := make(graphql.Fields)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// buildMutationType creates the Mutation type with all mutation fields.
func buildMutationType(res *resolver.Resolver) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			// Device mutations
			"updateFCMToken": &graphql.Field{
				Type:        DeviceType,
				Description: "Update FCM token for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"token": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: res.UpdateFCMToken,
			},
			"deleteDevice": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Delete a device",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.DeleteDevice,
			},
			// Command mutations
			"sendCommand": &graphql.Field{
				Type:        CommandResultType,
				Description: "Send a command to a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"command": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"args": &graphql.ArgumentConfig{
						Type: JSONScalar,
					},
				},
				Resolve: res.SendCommand,
			},
			"retryCommand": &graphql.Field{
				Type:        CommandType,
				Description: "Retry a failed command",
				Args: graphql.FieldConfigArgument{
					"dispatchId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.RetryCommand,
			},
			"cancelCommand": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Cancel a pending command",
				Args: graphql.FieldConfigArgument{
					"dispatchId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.CancelCommand,
			},
			// Device control mutations
			"disconnectDevice": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Force disconnect a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.DisconnectDevice,
			},
		},
	})
}
