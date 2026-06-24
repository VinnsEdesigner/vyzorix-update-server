// Package schema provides GraphQL schema definitions.
package schema

import (
	"github.com/graphql-go/graphql"
)

// CommandStatusEnum represents the status of a command.
var CommandStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "CommandStatus",
	Description: "Status of a device command",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Command is pending delivery",
		},
		"DELIVERED": &graphql.EnumValueConfig{
			Value:       "delivered",
			Description: "Command was delivered to device",
		},
		"FAILED": &graphql.EnumValueConfig{
			Value:       "failed",
			Description: "Command delivery failed",
		},
		"CANCELLED": &graphql.EnumValueConfig{
			Value:       "cancelled",
			Description: "Command was cancelled",
		},
	},
})
