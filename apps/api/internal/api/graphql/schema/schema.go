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
		Fields: mergeFields(settingsQueries(res), inboxQueries(res), deviceQueries(res), commandQueries(res), telemetryQueries(res),
			connectionQueries(res), dashboardQueries(res), updatesQueries(res), diagnosticsQueries(res),
			organizationQueries(res)),
	})
}

// settingsQueries returns GraphQL queries for settings management.
func settingsQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"mySettings": &graphql.Field{
			Type:        OperatorSettingsType,
			Description: "Get current operator's settings (client and notifications)",
			Resolve:    res.GetMySettings,
		},
		"deviceSettings": &graphql.Field{
			Type:        DeviceSettingsType,
			Description: "Get settings for a specific device including effective thresholds",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceImei":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Device IMEI"},
			},
			Resolve: res.GetDeviceSettings,
		},
		"organizationSettings": &graphql.Field{
			Type:        OrganizationSettingsType,
			Description: "Get settings for an organization including default thresholds",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
			},
			Resolve: res.GetOrganizationSettings,
		},
		"myNotifications": &graphql.Field{
			Type:        NotificationSettingsType,
			Description: "Get current operator's notification settings",
			Resolve:    res.GetMyNotifications,
		},
	}
}

func inboxQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"inbox": &graphql.Field{
			Type:        InboxListResponseType,
			Description: "Get paginated inbox entries for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"status":        &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "pending"},
				"page":          &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":         &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			},
			Resolve: res.GetInbox,
		},
		"inboxEntry": &graphql.Field{
			Type:        InboxEntryType,
			Description: "Get a single inbox entry by IMEI within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"imei":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: res.GetInboxEntry,
		},
	}
}

func deviceQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"device": &graphql.Field{
			Type:        DeviceType,
			Description: "Get a single device by ID within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"id":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetDevice,
		},
		"devices": &graphql.Field{
			Type:        graphql.NewList(DeviceType),
			Description: "List all devices for the authenticated operator in an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"limit":         &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
				"offset":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			},
			Resolve: res.GetDevices,
		},
		"deviceCount": &graphql.Field{
			Type:        graphql.Int,
			Description: "Get total device count for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
			},
			Resolve: res.GetDeviceCount,
		},
	}
}

func commandQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"command": &graphql.Field{
			Type:        CommandType,
			Description: "Get command status by dispatch ID within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"dispatchId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetCommand,
		},
		"pendingCommands": &graphql.Field{
			Type:        graphql.NewList(CommandType),
			Description: "Get pending commands for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetPendingCommands,
		},
	}
}

func telemetryQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"telemetryHistory": &graphql.Field{
			Type:        graphql.NewList(TelemetryEntryType),
			Description: "Query telemetry history for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"startTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":      &graphql.ArgumentConfig{Type: graphql.Int},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 100},
			},
			Resolve: res.GetTelemetryHistory,
		},
		"latestTelemetry": &graphql.Field{
			Type:        TelemetryEntryType,
			Description: "Get the latest telemetry entry for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetLatestTelemetry,
		},
		"telemetryStats": &graphql.Field{
			Type:        TelemetryStatsType,
			Description: "Get telemetry statistics for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetTelemetryStats,
		},
	}
}

func connectionQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"connectionStatus": &graphql.Field{
			Type:        ConnectionStatusType,
			Description: "Get connection status for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetConnectionStatus,
		},
		"allConnections": &graphql.Field{
			Type:        graphql.NewList(ConnectionStatusType),
			Description: "Get all device connection statuses for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
			},
			Resolve: res.GetAllConnections,
		},
	}
}

func dashboardQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"deviceMetrics": &graphql.Field{
			Type:        DeviceMetricsType,
			Description: "Get aggregated metrics for chart visualization within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"imei":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"range":        &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "6h"},
				"startTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":      &graphql.ArgumentConfig{Type: graphql.Int},
				"resolution":   &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: res.GetDeviceMetrics,
		},
		"deviceLogs": &graphql.Field{
			Type:        LogConnectionType,
			Description: "Get paginated device logs with cursor-based pagination for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"imei":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"type":         &graphql.ArgumentConfig{Type: graphql.String},
				"startTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":      &graphql.ArgumentConfig{Type: graphql.Int},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 100},
				"cursor":       &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: res.GetDeviceLogs,
		},
		"deviceCommandHistory": &graphql.Field{
			Type:        CommandHistoryType,
			Description: "Get paginated command history for a device within an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"imei":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"status":       &graphql.ArgumentConfig{Type: graphql.String},
				"page":         &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
				"startTime":    &graphql.ArgumentConfig{Type: graphql.Int},
				"endTime":      &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: res.GetDeviceCommandHistory,
		},
		"dashboardStats": &graphql.Field{
			Type:        DashboardStatsType,
			Description: "Get aggregated dashboard statistics for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
			},
			Resolve: res.GetDashboardStats,
		},
	}
}

func updatesQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"updatesStatus": &graphql.Field{
			Type:        UpdateStatusType,
			Description: "Get overall update system status for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceId":      &graphql.ArgumentConfig{Type: graphql.ID},
			},
			Resolve: res.GetUpdatesStatus,
		},
		"updatesVersions": &graphql.Field{
			Type:        UpdateVersionListType,
			Description: "List all available update versions for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"status":       &graphql.ArgumentConfig{Type: graphql.String},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
				"offset":       &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			},
			Resolve: res.GetUpdatesVersions,
		},
		"updatesChangelog": &graphql.Field{
			Type:        graphql.NewList(graphql.NewNonNull(ChangelogEntryType)),
			Description: "Get changelog entries for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"version":      &graphql.ArgumentConfig{Type: graphql.String},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
			},
			Resolve: res.GetUpdatesChangelog,
		},
		"updatesHistory": &graphql.Field{
			Type:        PushHistoryConnectionType,
			Description: "Get push history with pagination for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"status":       &graphql.ArgumentConfig{Type: graphql.String},
				"page":         &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 1},
				"limit":        &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			},
			Resolve: res.GetUpdatesHistory,
		},
		"updatesHistoryDetail": &graphql.Field{
			Type:        UpdatePushType,
			Description: "Get detailed push information for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"id":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: res.GetUpdatesHistoryDetail,
		},
		"updatesSyncStatus": &graphql.Field{
			Type:        SyncStatusType,
			Description: "Get GitHub sync status for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
			},
			Resolve: res.GetUpdatesSyncStatus,
		},
		// Spec-compliant aliases
		"updateHistory": &graphql.Field{
			Type:        UpdateHistoryType,
			Description: "Get a single update history record by ID (spec-compliant name)",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"id":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
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
			settingsMutations(res),
			inboxMutations(res),
			deviceMutations(res),
			commandMutations(res),
			updatesMutations(res),
			organizationMutations(res),
		),
	})
}

// settingsMutations returns GraphQL mutations for settings management.
func settingsMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"updateMyNotifications": &graphql.Field{
			Type:        NotificationSettingsType,
			Description: "Update current operator's notification settings",
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
						Name: "UpdateNotificationsInput",
						Fields: graphql.InputObjectConfigFieldMap{
							"enabled": &graphql.InputObjectFieldConfig{
								Type: graphql.Boolean,
							},
							"channels": &graphql.InputObjectFieldConfig{
								Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
							},
							"email": &graphql.InputObjectFieldConfig{
								Type: graphql.NewInputObject(graphql.InputObjectConfig{
									Name: "EmailNotificationInput",
									Fields: graphql.InputObjectConfigFieldMap{
										"thresholdBreach":     &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"deviceOffline":      &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"deviceOnline":       &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"updateAvailable":    &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"commandFailed":      &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"registrationRequest": &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
									},
								}),
							},
							"push": &graphql.InputObjectFieldConfig{
								Type: graphql.NewInputObject(graphql.InputObjectConfig{
									Name: "PushNotificationInput",
									Fields: graphql.InputObjectConfigFieldMap{
										"thresholdBreach":     &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"deviceOffline":      &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"deviceOnline":       &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"updateAvailable":    &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"commandFailed":      &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"registrationRequest": &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
									},
								}),
							},
							"webhook": &graphql.InputObjectFieldConfig{
								Type: graphql.NewInputObject(graphql.InputObjectConfig{
									Name: "WebhookNotificationInput",
									Fields: graphql.InputObjectConfigFieldMap{
										"enabled": &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
										"url":     &graphql.InputObjectFieldConfig{Type: graphql.String},
										"types":   &graphql.InputObjectFieldConfig{Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
									},
								}),
							},
						},
					})),
				},
			},
			Resolve: res.UpdateMyNotifications,
		},
		"updateDeviceSettings": &graphql.Field{
			Type:        DeviceSettingsType,
			Description: "Update settings for a specific device",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"deviceImei":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Device IMEI"},
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
						Name: "UpdateDeviceSettingsInput",
						Fields: graphql.InputObjectConfigFieldMap{
							"customName": &graphql.InputObjectFieldConfig{Type: graphql.String},
							"location":   &graphql.InputObjectFieldConfig{Type: graphql.String},
							"metadata":   &graphql.InputObjectFieldConfig{Type: graphql.NewList(graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
								Name: "MetadataInput",
								Fields: graphql.InputObjectConfigFieldMap{
									"key":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
									"value": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
								},
							})))},
							"thresholds": &graphql.InputObjectFieldConfig{
								Type: graphql.NewInputObject(graphql.InputObjectConfig{
									Name: "DeviceThresholdsInput",
									Fields: graphql.InputObjectConfigFieldMap{
										"riskWarn":    &graphql.InputObjectFieldConfig{Type: graphql.Int},
										"riskCrit":    &graphql.InputObjectFieldConfig{Type: graphql.Int},
										"thermalWarn": &graphql.InputObjectFieldConfig{Type: graphql.Int},
										"thermalCrit": &graphql.InputObjectFieldConfig{Type: graphql.Int},
										"bufferWarn":  &graphql.InputObjectFieldConfig{Type: graphql.Int},
										"bufferCrit":  &graphql.InputObjectFieldConfig{Type: graphql.Int},
									},
								}),
							},
						},
					})),
				},
			},
			Resolve: res.UpdateDeviceSettings,
		},
		"updateOrganizationSettings": &graphql.Field{
			Type:        OrganizationSettingsType,
			Description: "Update settings for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "Organization ID"},
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
						Name: "UpdateOrganizationSettingsInput",
						Fields: graphql.InputObjectConfigFieldMap{
							"timezone":              &graphql.InputObjectFieldConfig{Type: graphql.String},
							"dateFormat":            &graphql.InputObjectFieldConfig{Type: graphql.String},
							"alertCooldownMinutes": &graphql.InputObjectFieldConfig{Type: graphql.Int},
							"defaultThresholds": &graphql.InputObjectFieldConfig{
								Type: graphql.NewInputObject(graphql.InputObjectConfig{
									Name: "OrgDefaultThresholdsInput",
									Fields: graphql.InputObjectConfigFieldMap{
										"riskWarn":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
										"riskCrit":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
										"thermalWarn": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
										"thermalCrit": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
										"bufferWarn":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
										"bufferCrit":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
									},
								}),
							},
						},
					})),
				},
			},
			Resolve: res.UpdateOrganizationSettings,
		},
	}
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

func organizationQueries(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"organization": &graphql.Field{
			Type:        OrganizationType,
			Description: "Get an organization by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
			},
			Resolve: res.GetOrganization,
		},
		"organizations": &graphql.Field{
			Type:        OrganizationListResponseType,
			Description: "List all organizations for the authenticated operator",
			Args: graphql.FieldConfigArgument{
				"page": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 1,
					Description:  "Page number",
				},
				"limit": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 50,
					Description:  "Maximum number of organizations to return (max 100)",
				},
			},
			Resolve: res.GetOrganizations,
		},
		"myMemberships": &graphql.Field{
			Type:        MemberListResponseType,
			Description: "Get all organization memberships for the authenticated operator",
			Args: graphql.FieldConfigArgument{
				"page": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 1,
					Description:  "Page number",
				},
				"limit": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 50,
					Description:  "Maximum number of memberships to return (max 100)",
				},
			},
			Resolve: res.GetMyMemberships,
		},
		"organizationMembers": &graphql.Field{
			Type:        MemberListResponseType,
			Description: "List all active members of an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"page": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 1,
					Description:  "Page number",
				},
				"limit": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 50,
					Description:  "Maximum number of members to return (max 100)",
				},
			},
			Resolve: res.GetOrganizationMembers,
		},
		"organizationInvitations": &graphql.Field{
			Type:        InvitationListResponseType,
			Description: "List all pending invitations for an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"page": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 1,
					Description:  "Page number",
				},
				"limit": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 50,
					Description:  "Maximum number of invitations to return (max 100)",
				},
			},
			Resolve: res.GetOrganizationInvitations,
		},
		"myInvitations": &graphql.Field{
			Type:        graphql.NewList(InvitationType),
			Description: "Get all pending invitations for the authenticated operator",
			Resolve: res.GetMyInvitations,
		},
		"invitation": &graphql.Field{
			Type:        InvitationType,
			Description: "Get an invitation by token",
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Invitation token",
				},
			},
			Resolve: res.GetInvitationByToken,
		},
	}
}

func organizationMutations(res *resolver.Resolver) graphql.Fields {
	return graphql.Fields{
		"createOrganization": &graphql.Field{
			Type:        CreateOrganizationPayloadType,
			Description: "Create a new organization",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Organization name",
				},
				"maxMembers": &graphql.ArgumentConfig{
					Type: graphql.Int,
					DefaultValue: 100,
					Description: "Maximum number of members",
				},
			},
			Resolve: res.CreateOrganization,
		},
		"updateOrganization": &graphql.Field{
			Type:        OrganizationType,
			Description: "Update an organization",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
					Description: "New organization name",
				},
				"maxMembers": &graphql.ArgumentConfig{
					Type: graphql.Int,
					Description: "Maximum number of members",
				},
				"isActive": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
					Description: "Whether the organization is active",
				},
			},
			Resolve: res.UpdateOrganization,
		},
		"deleteOrganization": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete an organization (soft delete)",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
			},
			Resolve: res.DeleteOrganization,
		},
		"inviteMember": &graphql.Field{
			Type:        InvitationType,
			Description: "Invite a member to an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Email address to invite",
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(OrgRoleEnum),
					Description: "Role to assign",
				},
				"notes": &graphql.ArgumentConfig{
					Type: graphql.String,
					Description: "Optional notes for the invitation",
				},
			},
			Resolve: res.InviteMember,
		},
		"removeMember": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Remove a member from an organization",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"memberId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Membership ID to remove",
				},
			},
			Resolve: res.RemoveMember,
		},
		"updateMemberRole": &graphql.Field{
			Type:        MembershipType,
			Description: "Update a member's role",
			Args: graphql.FieldConfigArgument{
				"organizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Organization ID",
				},
				"memberId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Membership ID to update",
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(OrgRoleEnum),
					Description: "New role",
				},
			},
			Resolve: res.UpdateMemberRole,
		},
		"acceptInvitation": &graphql.Field{
			Type:        MembershipType,
			Description: "Accept an invitation to join an organization",
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Invitation token",
				},
				"notes": &graphql.ArgumentConfig{
					Type: graphql.String,
					Description: "Optional notes",
				},
			},
			Resolve: res.AcceptInvitation,
		},
		"rejectInvitation": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Reject an invitation",
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Invitation token",
				},
				"notes": &graphql.ArgumentConfig{
					Type: graphql.String,
					Description: "Reason for rejection",
				},
			},
			Resolve: res.RejectInvitation,
		},
		"cancelInvitation": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Cancel a pending invitation",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Invitation ID to cancel",
				},
			},
			Resolve: res.CancelInvitation,
		},
		"transferDevice": &graphql.Field{
			Type:        TransferDevicePayloadType,
			Description: "Transfer a device to another organization",
			Args: graphql.FieldConfigArgument{
				"imei": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "Device IMEI",
				},
				"sourceOrganizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Source organization ID",
				},
				"targetOrganizationId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
					Description: "Target organization ID",
				},
			},
			Resolve: res.TransferDevice,
		},
	}
}
