// Package schema provides GraphQL subscription types.
package schema

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/graphql-go/graphql"
)

// BuildSubscriptionType defines the subscription root type.
func BuildSubscriptionType(res *resolver.Resolver) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"deviceUpdated": &graphql.Field{
				Type:        DeviceType,
				Description: "Subscribe to device update events",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "Optional device ID to filter updates",
					},
				},
			},
			"telemetryReceived": &graphql.Field{
				Type:        TelemetryEntryType,
				Description: "Subscribe to real-time telemetry from devices",
				Args: graphql.FieldConfigArgument{
					"deviceId": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "Device ID to subscribe to (omit for all devices)",
					},
				},
			},
			"commandStatusChanged": &graphql.Field{
				Type:        CommandType,
				Description: "Subscribe to command status changes",
				Args: graphql.FieldConfigArgument{
					"dispatchId": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "Dispatch ID to track (omit for all commands)",
					},
				},
			},
		},
	})
}
