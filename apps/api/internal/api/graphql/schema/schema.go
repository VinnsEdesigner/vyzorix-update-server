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
		Fields: mergeFields(inboxQueries(res), deviceQueries(res), commandQueries(res), telemetryQueries(res),
			connectionQueries(res), dashboardQueries(res), updatesQueries(res), diagnosticsQueries(res)),
	})
}

func inboxQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"inbox": &graphql.Field{
			Type:        InboxListResponseType,
			Description: "Get paginated inbox entries",
			Args: graphql.FieldConfigArgument{
				"status": &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "pending"},
				"page":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			},
			Resolve: res.GetInbox,
		},
		"inboxEntry": &graphql.Field{
			Type:        InboxEntryType,
			Description: "Get a single inbox entry by IMEI",
			Args: graphql.FieldConfigArgument{
				"imei": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: res.GetInboxEntry,
		},
	}
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

func updatesQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"updatesStatus": &graphql.Field{
			Type:        UpdateStatusType,
			Description: "Get overall update system status",
			Args: graphql.FieldConfigArgument{
				"deviceId": &graphql.ArgumentConfig{Type: graphql.ID},
			},
			Resolve: res.GetUpdatesStatus,
		},
		"updatesVersions": &graphql.Field{
			Type:        UpdateVersionListType,
			Description: "List all available update versions",
			Args: graphql.FieldConfigArgument{
				"status": &graphql.ArgumentConfig{Type: graphql.String},
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
				"offset": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			},
			Resolve: res.GetUpdatesVersions,
		},
		"updatesChangelog": &graphql.Field{
			Type:        graphql.NewList(graphql.NewNonNull(ChangelogEntryType)),
			Description: "Get changelog entries",
			Args: graphql.FieldConfigArgument{
				"version": &graphql.ArgumentConfig{Type: graphql.String},
				"limit":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
			},
			Resolve: res.GetUpdatesChangelog,
		},
		"updatesHistory": &graphql.Field{
			Type:        PushHistoryConnectionType,
			Description: "Get push history with pagination",
			Args: graphql.FieldConfigArgument{
				"status": &graphql.ArgumentConfig{Type: graphql.String},
				"page":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			},
			Resolve: res.GetUpdatesHistory,
		},
		"updatesHistoryDetail": &graphql.Field{
			Type:        UpdatePushType,
			Description: "Get detailed push information",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetUpdatesHistoryDetail,
		},
		"updatesSyncStatus": &graphql.Field{
			Type:        SyncStatusType,
			Description: "Get GitHub sync status",
			Resolve:     res.GetUpdatesSyncStatus,
		},
		// Spec-compliant aliases
		"updateHistory": &graphql.Field{
			Type:        UpdateHistoryType,
			Description: "Get a single update history record by ID (spec-compliant name)",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetUpdatesHistoryDetail,
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
		Fields: mergeMutationFields(
			inboxMutations(res),
			deviceMutations(res),
			commandMutations(res),
			updatesMutations(res),
		),
	})
}

func inboxMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"ackInbox": &graphql.Field{
			Type:        AckResultType,
			Description: "Acknowledge (approve/reject) an inbox entry",
			Args: graphql.FieldConfigArgument{
				"imei": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"action": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(AckActionEnum),
				},
				"notes": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: res.AckInbox,
		},
		"deregisterDevice": &graphql.Field{
			Type:        DeregisterResultType,
			Description: "Deregister a device (soft delete with 30-day retention)",
			Args: graphql.FieldConfigArgument{
				"imei": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"hard": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
				},
			},
			Resolve: res.DeregisterDeviceGraphQL,
		},
	}
}

func deviceMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
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
	}
}

func commandMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
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
			Type:        CancelCommandResponseType,
			Description: "Cancel a pending command",
			Args: graphql.FieldConfigArgument{
				"dispatchId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
				},
			},
			Resolve: res.CancelCommand,
		},
	}
}

func updatesMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"pushUpdate": &graphql.Field{
			Type:        PushUpdateResponseType,
			Description: "Push an update to devices",
			Args: graphql.FieldConfigArgument{
				"version": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"deviceIds": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.ID))),
				},
				"installType": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"scheduledAt": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: res.PushUpdate,
		},
		"cancelUpdate": &graphql.Field{
			Type:        CancelPushResponseType,
			Description: "Cancel a pending update push",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
				},
			},
			Resolve: res.CancelUpdate,
		},
		"syncFromGitHub": &graphql.Field{
			Type:        SyncResponseType,
			Description: "Trigger a GitHub sync",
			Resolve:     res.SyncFromGitHub,
		},
		// Spec-compliant alias
		"syncUpdates": &graphql.Field{
			Type:        SyncResponseType,
			Description: "Sync updates from GitHub (spec-compliant name)",
			Resolve:     res.SyncFromGitHub,
		},
	}
}

func mergeMutationFields(maps ...graphql.Fields) graphql.Fields {
	result := make(graphql.Fields)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func diagnosticsQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"deviceInspection": &graphql.Field{
			Type:        DeviceInspectionType,
			Description: "Get full device inspection data for diagnostics",
			Args: graphql.FieldConfigArgument{
				"imei": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: res.GetDeviceInspection,
		},
		"deviceTimeline": &graphql.Field{
			Type:        TimelineConnectionType,
			Description: "Get chronological event timeline for a device",
			Args: graphql.FieldConfigArgument{
				"imei":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"eventType": &graphql.ArgumentConfig{Type: TimelineEventTypeEnum},
				"startTime":  &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"limit":      &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
				"cursor":     &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: res.GetDeviceTimeline,
		},
	}
}
