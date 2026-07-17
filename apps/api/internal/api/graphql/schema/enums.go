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

// DeviceLifecycleEnum represents the lifecycle state of a device.
// Maps to domain/device.Lifecycle: Pending → Registered → Deregistered
var DeviceLifecycleEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "DeviceLifecycle",
	Description: "Lifecycle state of a device registration",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Device is pending approval",
		},
		"REGISTERED": &graphql.EnumValueConfig{
			Value:       "registered",
			Description: "Device is registered and active",
		},
		"DEREGISTERED": &graphql.EnumValueConfig{
			Value:       "deregistered",
			Description: "Device has been deregistered",
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

// TimelineEventTypeEnum for timeline event types.
var TimelineEventTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "TimelineEventType",
	Description: "Type of device timeline event",
	Values: graphql.EnumValueConfigMap{
		"TELEMETRY": &graphql.EnumValueConfig{Value: "TELEMETRY"},
		"COMMAND_SENT": &graphql.EnumValueConfig{Value: "COMMAND_SENT"},
		"COMMAND_ACK": &graphql.EnumValueConfig{Value: "COMMAND_ACK"},
		"COMMAND_FAILED": &graphql.EnumValueConfig{Value: "COMMAND_FAILED"},
		"CONNECTION_OPEN": &graphql.EnumValueConfig{Value: "CONNECTION_OPEN"},
		"CONNECTION_LOST": &graphql.EnumValueConfig{Value: "CONNECTION_LOST"},
		"FCM_FALLBACK": &graphql.EnumValueConfig{Value: "FCM_FALLBACK"},
		"RECONNECTED": &graphql.EnumValueConfig{Value: "RECONNECTED"},
		"THRESHOLD_BREACH": &graphql.EnumValueConfig{Value: "THRESHOLD_BREACH"},
		"REGISTERED": &graphql.EnumValueConfig{Value: "REGISTERED"},
		"DEREGISTERED": &graphql.EnumValueConfig{Value: "DEREGISTERED"},
		"ERROR": &graphql.EnumValueConfig{Value: "ERROR"},
	},
})

// OrganizationLifecycleEnum represents the lifecycle state of an organization.
// Maps to domain/organization.OrganizationLifecycle: Active → Inactive/Archived
var OrganizationLifecycleEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "OrganizationLifecycle",
	Description: "Lifecycle state of an organization",
	Values: graphql.EnumValueConfigMap{
		"ACTIVE": &graphql.EnumValueConfig{
			Value:       "active",
			Description: "Organization is active",
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value:       "inactive",
			Description: "Organization is inactive (suspended)",
		},
		"ARCHIVED": &graphql.EnumValueConfig{
			Value:       "archived",
			Description: "Organization has been archived (soft-deleted)",
		},
	},
})

// MemberLifecycleEnum represents the lifecycle state of an organization member.
// Maps to domain/organization.MemberLifecycle: Invited → Active → Suspended/Removed
var MemberLifecycleEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "MemberLifecycle",
	Description: "Lifecycle state of an organization member",
	Values: graphql.EnumValueConfigMap{
		"INVITED": &graphql.EnumValueConfig{
			Value:       "invited",
			Description: "Member has been invited but hasn't joined",
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value:       "active",
			Description: "Member is active in the organization",
		},
		"SUSPENDED": &graphql.EnumValueConfig{
			Value:       "suspended",
			Description: "Member's access has been suspended",
		},
		"REMOVED": &graphql.EnumValueConfig{
			Value:       "removed",
			Description: "Member has been removed from the organization",
		},
	},
})

// OrgRoleEnum represents the role of an organization member.
// Maps to domain/organization.OrganizationRole: super_admin, admin, operator, viewer
var OrgRoleEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "OrgRole",
	Description: "Role of an organization member",
	Values: graphql.EnumValueConfigMap{
		"SUPER_ADMIN": &graphql.EnumValueConfig{
			Value:       "super_admin",
			Description: "Super admin - full control including deletion",
		},
		"ADMIN": &graphql.EnumValueConfig{
			Value:       "admin",
			Description: "Admin - can manage members and devices",
		},
		"OPERATOR": &graphql.EnumValueConfig{
			Value:       "operator",
			Description: "Operator - can send commands and manage devices",
		},
		"VIEWER": &graphql.EnumValueConfig{
			Value:       "viewer",
			Description: "Viewer - read-only access",
		},
	},
})

// InvitationStatusEnum represents the status of an invitation.
// Maps to domain/organization.InvitationStatus: pending, accepted, rejected, expired
var InvitationStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name:        "InvitationStatus",
	Description: "Status of an organization invitation",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value:       "pending",
			Description: "Invitation is pending",
		},
		"ACCEPTED": &graphql.EnumValueConfig{
			Value:       "accepted",
			Description: "Invitation was accepted",
		},
		"REJECTED": &graphql.EnumValueConfig{
			Value:       "rejected",
			Description: "Invitation was rejected",
		},
		"EXPIRED": &graphql.EnumValueConfig{
			Value:       "expired",
			Description: "Invitation has expired",
		},
	},
})
