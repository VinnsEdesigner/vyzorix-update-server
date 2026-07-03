// Package schema provides GraphQL schema definitions.
package schema

import (
	"github.com/graphql-go/graphql"
)

// InboxStatusEnum represents the status of an inbox entry.
var InboxStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "InboxStatus",
	Description: "Status of an inbox entry",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Registration request is pending",
		},
		"APPROVED": &graphql.EnumValueConfig{
			Value:       "approved",
			Description: "Registration request was approved",
		},
		"REJECTED": &graphql.EnumValueConfig{
			Value:       "rejected",
			Description: "Registration request was rejected",
		},
	},
})

// DeviceStatusEnum represents the status of a device.
var DeviceStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "DeviceStatus",
	Description: "Status of a device",
	Values: graphql.EnumValueConfigMap{
		"ONLINE": &graphql.EnumValueConfig{
			Value:       "online",
			Description: "Device is currently online",
		},
		"OFFLINE": &graphql.EnumValueConfig{
			Value:       "offline",
			Description: "Device is offline",
		},
		"DEREGISTERED": &graphql.EnumValueConfig{
			Value:       "deregistered",
			Description: "Device has been deregistered",
		},
	},
})

// AckActionEnum represents the action for acknowledging an inbox entry.
var AckActionEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "AckAction",
	Description: "Action to take on an inbox entry",
	Values: graphql.EnumValueConfigMap{
		"APPROVE": &graphql.EnumValueConfig{
			Value:       "approve",
			Description: "Approve the registration request",
		},
		"REJECT": &graphql.EnumValueConfig{
			Value:       "reject",
			Description: "Reject the registration request",
		},
	},
})

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
		"COMPLETED": &graphql.EnumValueConfig{
			Value:       "completed",
			Description: "Command completed successfully",
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

// InstallTypeEnum represents the install type for an update push.
var InstallTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "InstallType",
	Description: "Install type for update push",
	Values: graphql.EnumValueConfigMap{
		"IMMEDIATE": &graphql.EnumValueConfig{
			Value:       "immediate",
			Description: "Install immediately",
		},
		"SCHEDULED": &graphql.EnumValueConfig{
			Value:       "scheduled",
			Description: "Install at scheduled time",
		},
	},
})

// ReleaseTypeEnum represents the release type of an update version.
var ReleaseTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "ReleaseType",
	Description: "Release type for an update version",
	Values: graphql.EnumValueConfigMap{
		"MAJOR": &graphql.EnumValueConfig{
			Value:       "major",
			Description: "Major release",
		},
		"MINOR": &graphql.EnumValueConfig{
			Value:       "minor",
			Description: "Minor release",
		},
		"PATCH": &graphql.EnumValueConfig{
			Value:       "patch",
			Description: "Patch release",
		},
	},
})

// UpdateStatusEnum represents the status of an update push.
var UpdateStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "UpdateStatus",
	Description: "Status of an update push",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Push is pending",
		},
		"IN_PROGRESS": &graphql.EnumValueConfig{
			Value:       "in_progress",
			Description: "Push is in progress",
		},
		"COMPLETED": &graphql.EnumValueConfig{
			Value:       "completed",
			Description: "Push completed successfully",
		},
		"FAILED": &graphql.EnumValueConfig{
			Value:       "failed",
			Description: "Push failed",
		},
		"CANCELLED": &graphql.EnumValueConfig{
			Value:       "cancelled",
			Description: "Push was cancelled",
		},
	},
})

// DevicePushStatusEnum represents the status of a device push.
var DevicePushStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "DevicePushStatus",
	Description: "Status of a device push",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Device push is pending",
		},
		"SENT": &graphql.EnumValueConfig{
			Value:       "sent",
			Description: "Update command sent to device",
		},
		"ACKNOWLEDGED": &graphql.EnumValueConfig{
			Value:       "acknowledged",
			Description: "Device acknowledged the update",
		},
		"FAILED": &graphql.EnumValueConfig{
			Value:       "failed",
			Description: "Device push failed",
		},
	},
})
