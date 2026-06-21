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
		Name: "Query",
		Fields: graphql.Fields{
			// Device queries
			"device": &graphql.Field{
				Type:        DeviceType,
				Description: "Get a single device by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetDevice,
			},
			"devices": &graphql.Field{
				Type:        graphql.NewList(DeviceType),
				Description: "List all devices for the authenticated operator",
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 50,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: res.GetDevices,
			},
			"deviceCount": &graphql.Field{
				Type:        graphql.Int,
				Description: "Get total device count",
				Resolve:     res.GetDeviceCount,
			},
			// Command queries
			"command": &graphql.Field{
				Type:        CommandType,
				Description: "Get command status by dispatch ID",
				Args: graphql.FieldConfigArgument{
					"dispatchId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetCommand,
			},
			"pendingCommands": &graphql.Field{
				Type:        graphql.NewList(CommandType),
				Description: "Get pending commands for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetPendingCommands,
			},
			// Telemetry queries
			"telemetryHistory": &graphql.Field{
				Type:        graphql.NewList(TelemetryEntryType),
				Description: "Query telemetry history for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"startTime": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"endTime": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 100,
					},
				},
				Resolve: res.GetTelemetryHistory,
			},
			"latestTelemetry": &graphql.Field{
				Type:        TelemetryEntryType,
				Description: "Get the latest telemetry entry for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetLatestTelemetry,
			},
			"telemetryStats": &graphql.Field{
				Type:        TelemetryStatsType,
				Description: "Get telemetry statistics for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetTelemetryStats,
			},
			// Connection status queries
			"connectionStatus": &graphql.Field{
				Type:        ConnectionStatusType,
				Description: "Get connection status for a device",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: res.GetConnectionStatus,
			},
			"allConnections": &graphql.Field{
				Type:        graphql.NewList(ConnectionStatusType),
				Description: "Get all device connection statuses",
				Resolve:     res.GetAllConnections,
			},
		},
	})
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