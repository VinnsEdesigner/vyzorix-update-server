// Re-export shared utilities
export * from "./_shared";

// Export each module's entities directly
export * from "./admin";
export * from "./apikey";
export * from "./auth";
export * from "./clientcredentials";
export * from "./email";
export * from "./oauth";
export * from "./organization";
export * from "./session";

// Diagnostics module - selectively export to avoid conflict with realtime
export { type DeviceInspection, type TimelineEvent, type TimelineResult } from "./diagnostics/diagnostics-entity";
export { deviceInspectionFromRaw, timelineResultFromRaw } from "./diagnostics/diagnostics-mappers";

// Events module - selectively export to avoid conflicts
export { type Event, type EventResult } from "./events/events-entity";
export { eventFromRaw, eventsFromRaw } from "./events/events-mappers";

// Realtime module - selectively export to avoid conflict with diagnostics
export { type WSTelemetry, type WebSocketState } from "./realtime/realtime-entity";
export { telemetryFromRaw, eventFromRaw as wsEventFromRaw } from "./realtime/realtime-mappers";

// For modules with conflicting names, use selective exports
// Commands module
export { type Command, type CommandListItem, type CommandStatus, type CommandResult, type SendCommandRequest, type PresetCommandType, type CommandParams, COMMAND_STATUSES, PRESET_COMMANDS, isCommandTerminal, getCommandLabel, getStatusLabel } from "./commands/commands-entity";
export { commandFromRaw, commandListItemFromRaw, sendCommandRequestToRaw } from "./commands/commands-mappers";

// Device module  
export { type DeviceStatus, type Device, type DeviceListItem, type DeviceStats, type DeviceListResult, type DeviceConnection, formatDeviceName, isDeviceOnline } from "./device/device-entity";
export { deviceFromRaw, deviceListItemFromRaw, deviceStatsFromRaw } from "./device/device-mappers";

// Invitation module
export { type InvitationLifecycle, type Invitation, type InvitationApiResponse } from "./invitation/invitation-entity";

// Logs module
export { type LogEntry, type LogStats, type LogListResult } from "./logs/logs-entity";

// Metrics module
export { type MetricResolution, type TimeRange, type DashboardStats, type DeviceMetrics } from "./metrics/metrics-entity";

// Registration module
export { type CreateInboxRequest, type InboxEntry, type InboxStatus, type AcknowledgeAction, type CreateInboxResult, type ConfirmDeviceResult, type AckResult, type DeregisterResult } from "./registration/registration-entity";
export { inboxEntryFromRaw, deviceFromRaw as deviceFromRawReg, createInboxRequestToRaw, createInboxResultFromRaw, confirmDeviceResultFromRaw } from "./registration/registration-mappers";

// Settings module
export { type NotificationSettings, type SecuritySettings, type Thresholds, type ClientSettings, DEFAULT_THRESHOLDS, DEFAULT_CLIENT_SETTINGS, DEFAULT_SECURITY_SETTINGS } from "./settings/settings-entity";
export { thresholdsFromRaw, clientSettingsFromRaw, securitySettingsFromRaw, emailNotificationsFromRaw, pushNotificationsFromRaw, webhookNotificationsToRaw, notificationSettingsToRaw } from "./settings/settings-mappers";

// Telemetry module
export { type TelemetryFrame, type RawTelemetryFrame } from "./telemetry/telemetry-entity";

// Updates module
export { type Version as UpdateVersion, type UpdateStatus, type UpdatePush, type ChangelogEntry } from "./updates/updates-entity";
export { versionFromRaw, syncStateFromRaw, pushDevicesFromRaw, updatePushFromRaw, versionListResultFromRaw, updateHistoryResultFromRaw, changelogEntryFromRaw } from "./updates/updates-mappers";
