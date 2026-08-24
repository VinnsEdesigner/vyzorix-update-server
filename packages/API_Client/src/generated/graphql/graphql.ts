/* eslint-disable */
/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
/** Action to take on an inbox entry */
export type AckAction =
  /** Approve the registration request */
  | 'APPROVE'
  /** Reject the registration request */
  | 'REJECT';

/** Status of a device command */
export type CommandStatus =
  /** Command was cancelled */
  | 'CANCELLED'
  /** Command completed successfully */
  | 'COMPLETED'
  /** Command was delivered to device */
  | 'DELIVERED'
  /** Command delivery failed */
  | 'FAILED'
  /** Command is pending delivery */
  | 'PENDING';

/** Status of a device push */
export type DevicePushStatus =
  /** Device acknowledged the update */
  | 'ACKNOWLEDGED'
  /** Device push failed */
  | 'FAILED'
  /** Device push is pending */
  | 'PENDING'
  /** Update command sent to device */
  | 'SENT';

/** Status of a device */
export type DeviceStatus =
  /** Device has been deregistered */
  | 'DEREGISTERED'
  /** Device is offline */
  | 'OFFLINE'
  /** Device is currently online */
  | 'ONLINE';

/** Device alert thresholds (risk/thermal/buffer warn+crit) */
export type DeviceThresholdsInput = {
  bufferCrit?: number | null | undefined;
  bufferWarn?: number | null | undefined;
  riskCrit?: number | null | undefined;
  riskWarn?: number | null | undefined;
  thermalCrit?: number | null | undefined;
  thermalWarn?: number | null | undefined;
};

/** Email notification channel settings */
export type EmailNotificationInput = {
  commandFailed?: boolean | null | undefined;
  deviceOffline?: boolean | null | undefined;
  deviceOnline?: boolean | null | undefined;
  registrationRequest?: boolean | null | undefined;
  thresholdBreach?: boolean | null | undefined;
  updateAvailable?: boolean | null | undefined;
};

/** Status of an inbox entry */
export type InboxStatus =
  /** Registration request was approved */
  | 'APPROVED'
  /** Registration request is pending */
  | 'PENDING'
  /** Registration request was rejected */
  | 'REJECTED';

/** Install type for update push */
export type InstallType =
  /** Install immediately */
  | 'IMMEDIATE'
  /** Install at scheduled time */
  | 'SCHEDULED';

/** Type of member event */
export type MemberEventType =
  /** A member was invited */
  | 'MEMBER_INVITED'
  /** A member joined the organization */
  | 'MEMBER_JOINED'
  /** A member was reactivated */
  | 'MEMBER_REACTIVATED'
  /** A member was removed */
  | 'MEMBER_REMOVED'
  /** A member was suspended */
  | 'MEMBER_SUSPENDED'
  /** A member's role was changed */
  | 'ROLE_CHANGED';

/** Device metadata entries (key/value) */
export type MetadataInput = {
  key: string;
  value: string;
};

/** Organization default alert thresholds applied to devices */
export type OrgDefaultThresholdsInput = {
  bufferCrit: number;
  bufferWarn: number;
  riskCrit: number;
  riskWarn: number;
  thermalCrit: number;
  thermalWarn: number;
};

/** Type of organization event */
export type OrganizationEventType =
  /** Organization was activated */
  | 'ACTIVATED'
  /** Organization was created */
  | 'CREATED'
  /** Organization was deactivated */
  | 'DEACTIVATED'
  /** Organization was deleted */
  | 'DELETED'
  /** Organization was updated */
  | 'UPDATED';

/** Push notification channel settings */
export type PushNotificationInput = {
  commandFailed?: boolean | null | undefined;
  deviceOffline?: boolean | null | undefined;
  deviceOnline?: boolean | null | undefined;
  registrationRequest?: boolean | null | undefined;
  thresholdBreach?: boolean | null | undefined;
  updateAvailable?: boolean | null | undefined;
};

/** Release type for an update version */
export type ReleaseType =
  /** Major release */
  | 'MAJOR'
  /** Minor release */
  | 'MINOR'
  /** Patch release */
  | 'PATCH';

/** Type of device timeline event */
export type TimelineEventType =
  | 'COMMAND_ACK'
  | 'COMMAND_FAILED'
  | 'COMMAND_SENT'
  | 'CONNECTION_LOST'
  | 'CONNECTION_OPEN'
  | 'DEREGISTERED'
  | 'ERROR'
  | 'FCM_FALLBACK'
  | 'RECONNECTED'
  | 'REGISTERED'
  | 'TELEMETRY'
  | 'THRESHOLD_BREACH';

/** Per-device settings update (custom name, location, metadata, thresholds) */
export type UpdateDeviceSettingsInput = {
  customName?: string | null | undefined;
  location?: string | null | undefined;
  metadata?: Array<MetadataInput> | null | undefined;
  thresholds?: DeviceThresholdsInput | null | undefined;
};

/** Notification settings update (per-channel toggles + webhook config) */
export type UpdateNotificationsInput = {
  channels?: Array<string> | null | undefined;
  email?: EmailNotificationInput | null | undefined;
  enabled?: boolean | null | undefined;
  push?: PushNotificationInput | null | undefined;
  webhook?: WebhookNotificationInput | null | undefined;
};

/** Organization settings update (timezone, date format, alert cooldown, default thresholds) */
export type UpdateOrganizationSettingsInput = {
  alertCooldownMinutes?: number | null | undefined;
  dateFormat?: string | null | undefined;
  defaultThresholds?: OrgDefaultThresholdsInput | null | undefined;
  timezone?: string | null | undefined;
};

/** Status of an update push */
export type UpdatePushStatus =
  /** Push was cancelled */
  | 'CANCELLED'
  /** Push completed successfully */
  | 'COMPLETED'
  /** Push failed */
  | 'FAILED'
  /** Push is in progress */
  | 'IN_PROGRESS'
  /** Push is pending */
  | 'PENDING';

/** Webhook notification channel settings (URL + signing secret) */
export type WebhookNotificationInput = {
  enabled?: boolean | null | undefined;
  types?: Array<string> | null | undefined;
  url?: string | null | undefined;
};

export type SendCommandMutationVariables = Exact<{
  deviceId: string | number;
  command: string;
  args?: unknown;
}>;


export type SendCommandMutation = { sendCommand: { dispatchId: string, commandId: string, status: string, deviceOnline: boolean } | null };

export type RetryCommandMutationVariables = Exact<{
  dispatchId: string | number;
}>;


export type RetryCommandMutation = { retryCommand: { dispatchId: string, commandId: string, deviceId: string, command: string, status: CommandStatus, createdAt: string | null, deliveredAt: string | null } | null };

export type CancelCommandMutationVariables = Exact<{
  dispatchId: string | number;
}>;


export type CancelCommandMutation = { cancelCommand: { dispatchId: string, cancelledAt: number, status: string } | null };

export type GetPendingCommandsQueryVariables = Exact<{
  organizationId: string;
  deviceId: string | number;
}>;


export type GetPendingCommandsQuery = { pendingCommands: Array<{ dispatchId: string, commandId: string, deviceId: string, command: string, args: unknown, status: CommandStatus, createdAt: string | null, deliveredAt: string | null } | null> | null };

export type GetCommandQueryVariables = Exact<{
  organizationId: string;
  dispatchId: string | number;
}>;


export type GetCommandQuery = { command: { dispatchId: string, commandId: string, deviceId: string, command: string, args: unknown, status: CommandStatus, createdAt: string | null, deliveredAt: string | null } | null };

export type DeviceListFragment = { id: string, imei: string, name: string | null, deviceName: string | null, model: string | null, manufacturer: string | null, status: string | null, lastSeen: string | null, online: boolean };

export type DeviceFragment = { id: string, imei: string, name: string | null, deviceName: string | null, model: string | null, manufacturer: string | null, appVersion: string | null, osVersion: string | null, securityPatch: string | null, buildId: string | null, status: string | null, registeredAt: string | null, lastSeen: string | null, fcmTokenValid: boolean | null, commandSecretSet: boolean | null, connection: { webSocketStatus: string, connectedAt: string | null, protocol: string | null, clientIp: string | null } | null };

export type GetDevicesQueryVariables = Exact<{
  organizationId: string;
  limit?: number | null | undefined;
  offset?: number | null | undefined;
}>;


export type GetDevicesQuery = { devices: Array<{ id: string, imei: string, name: string | null, deviceName: string | null, model: string | null, manufacturer: string | null, status: string | null, lastSeen: string | null, online: boolean } | null> | null };

export type GetDeviceQueryVariables = Exact<{
  organizationId: string;
  id: string | number;
}>;


export type GetDeviceQuery = { device: { id: string, imei: string, name: string | null, deviceName: string | null, model: string | null, manufacturer: string | null, appVersion: string | null, osVersion: string | null, securityPatch: string | null, buildId: string | null, status: string | null, registeredAt: string | null, lastSeen: string | null, fcmTokenValid: boolean | null, commandSecretSet: boolean | null, connection: { webSocketStatus: string, connectedAt: string | null, protocol: string | null, clientIp: string | null } | null } | null };

export type GetDeviceCountQueryVariables = Exact<{
  organizationId: string;
}>;


export type GetDeviceCountQuery = { deviceCount: number | null };

export type GetDeviceInspectionQueryVariables = Exact<{
  imei: string;
  organizationId: string;
}>;


export type GetDeviceInspectionQuery = { deviceInspection: { identity: { imei: string, deviceName: string | null, model: string | null, manufacturer: string | null }, software: { osVersion: string, appVersion: string, securityPatch: string | null, buildId: string | null }, registration: { status: DeviceStatus, registeredAt: string | null, fcmTokenValid: boolean, fcmTokenRefreshedAt: string | null, commandSecretSet: boolean }, connection: { webSocketStatus: string, connectedAt: string | null, fcmStatus: string, lastSeen: string | null, clientIp: string | null, protocol: string | null }, telemetry: { lastTimestamp: string, framesToday: number, avgLatencyMs: number | null, totalBytesToday: number, sessionsToday: number } } | null };

export type GetDeviceTimelineQueryVariables = Exact<{
  imei: string;
  organizationId: string;
  eventType?: TimelineEventType | null | undefined;
  startTime?: number | null | undefined;
  endTime?: number | null | undefined;
  limit?: number | null | undefined;
  cursor?: string | null | undefined;
}>;


export type GetDeviceTimelineQuery = { deviceTimeline: { hasMore: boolean, nextCursor: string | null, events: Array<{ id: string, type: TimelineEventType, timestamp: string, data: unknown }> } | null };

export type LogEntryFragment = { id: string, type: string, timestamp: number, data: unknown };

export type GetLogsQueryVariables = Exact<{
  organizationId: string;
  imei: string | number;
  type?: string | null | undefined;
  startTime?: number | null | undefined;
  endTime?: number | null | undefined;
  limit?: number | null | undefined;
  cursor?: string | null | undefined;
}>;


export type GetLogsQuery = { deviceLogs: { events: Array<{ id: string, type: string, timestamp: number, data: unknown } | null> | null, pagination: { limit: number, hasMore: boolean, nextCursor: string | null } } | null };

export type OnDeviceUpdatedSubscriptionVariables = Exact<{
  deviceId?: string | number | null | undefined;
}>;


export type OnDeviceUpdatedSubscription = { deviceUpdated: { id: string, imei: string, deviceName: string | null, status: string | null, lastSeen: string | null } | null };

export type OnTelemetryReceivedSubscriptionVariables = Exact<{
  deviceId?: string | number | null | undefined;
}>;


export type OnTelemetryReceivedSubscription = { telemetryReceived: { id: string, deviceId: string, receivedAt: string | null, riskScore: number | null, bufferLevel: number | null, thermalTemp: number | null, payload: string | null } | null };

export type OnCommandStatusChangedSubscriptionVariables = Exact<{
  dispatchId?: string | number | null | undefined;
}>;


export type OnCommandStatusChangedSubscription = { commandStatusChanged: { dispatchId: string, commandId: string, deviceId: string, command: string, status: CommandStatus, createdAt: string | null } | null };

export type OnOrganizationEventSubscriptionVariables = Exact<{
  orgId: string | number;
}>;


export type OnOrganizationEventSubscription = { organizationEvent: { type: OrganizationEventType, timestamp: string, data: unknown } | null };

export type OnMemberEventSubscriptionVariables = Exact<{
  orgId: string | number;
}>;


export type OnMemberEventSubscription = { memberEvent: { type: MemberEventType, timestamp: string, memberId: string, data: unknown } | null };

export type AckInboxMutationVariables = Exact<{
  imei: string;
  action: AckAction;
  notes?: string | null | undefined;
}>;


export type AckInboxMutation = { ackInbox: { id: string, imei: string, status: InboxStatus, approvedAt: number | null, rejectedAt: number | null, commandSecret: string | null, fcmPushSent: boolean | null, notes: string | null } | null };

export type DeregisterDeviceMutationVariables = Exact<{
  imei: string;
  hard?: boolean | null | undefined;
}>;


export type DeregisterDeviceMutation = { deregisterDevice: { imei: string, status: string, deregisteredAt: number, retentionUntil: number } | null };

export type InboxEntryFragment = { id: string, imei: string, model: string | null, manufacturer: string | null, osVersion: string | null, appVersion: string | null, firebaseInstallId: string | null, status: string, notes: string | null, operatorId: string | null, createdAt: number, approvedAt: number | null, rejectedAt: number | null };

export type GetInboxEntriesQueryVariables = Exact<{
  organizationId: string;
  status?: string | null | undefined;
  page?: number | null | undefined;
  limit?: number | null | undefined;
}>;


export type GetInboxEntriesQuery = { inbox: { requests: Array<{ id: string, imei: string, model: string | null, manufacturer: string | null, osVersion: string | null, appVersion: string | null, firebaseInstallId: string | null, status: string, notes: string | null, operatorId: string | null, createdAt: number, approvedAt: number | null, rejectedAt: number | null }>, pagination: { total: number, limit: number, offset: number, hasMore: boolean } } | null };

export type GetInboxEntryQueryVariables = Exact<{
  organizationId: string;
  imei: string;
}>;


export type GetInboxEntryQuery = { inboxEntry: { id: string, imei: string, model: string | null, manufacturer: string | null, osVersion: string | null, appVersion: string | null, firebaseInstallId: string | null, status: string, notes: string | null, operatorId: string | null, createdAt: number, approvedAt: number | null, rejectedAt: number | null } | null };

export type OperatorSettingsFragment = { client: { serverUrl: string | null, deviceId: string | null, requestTimeoutMs: number, logBufferLimit: number, signalHistoryLimit: number, autoReconnect: boolean, strictHmac: boolean }, notifications: { enabled: boolean, channels: Array<string>, email: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean }, push: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean }, webhook: { enabled: boolean, url: string | null } } };

export type DeviceSettingsFragment = { id: string, deviceImei: string, customName: string | null, location: string | null, createdAt: string | null, updatedAt: string | null, thresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } | null, effectiveThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } };

export type OrganizationSettingsFragment = { id: string, organizationId: string, dateFormat: string | null, alertCooldownMinutes: number, createdAt: string | null, updatedAt: string | null, defaultThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } };

export type UpdateNotificationsMutationVariables = Exact<{
  input: UpdateNotificationsInput;
}>;


export type UpdateNotificationsMutation = { updateMyNotifications: { enabled: boolean, channels: Array<string>, email: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean, registrationRequest: boolean }, push: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean, registrationRequest: boolean }, webhook: { enabled: boolean, url: string | null, types: Array<string> } } | null };

export type UpdateDeviceSettingsMutationVariables = Exact<{
  organizationId: string;
  deviceImei: string;
  input: UpdateDeviceSettingsInput;
}>;


export type UpdateDeviceSettingsMutation = { updateDeviceSettings: { id: string, deviceImei: string, customName: string | null, location: string | null, createdAt: string | null, updatedAt: string | null, thresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } | null, effectiveThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } } | null };

export type UpdateOrganizationSettingsMutationVariables = Exact<{
  organizationId: string;
  input: UpdateOrganizationSettingsInput;
}>;


export type UpdateOrganizationSettingsMutation = { updateOrganizationSettings: { id: string, organizationId: string, timezone: string | null, dateFormat: string | null, alertCooldownMinutes: number, createdAt: string | null, updatedAt: string | null, defaultThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } } | null };

export type GetSettingsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetSettingsQuery = { mySettings: { client: { serverUrl: string | null, deviceId: string | null, requestTimeoutMs: number, logBufferLimit: number, signalHistoryLimit: number, autoReconnect: boolean, strictHmac: boolean }, notifications: { enabled: boolean, channels: Array<string>, email: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean }, push: { thresholdBreach: boolean, deviceOffline: boolean, deviceOnline: boolean, updateAvailable: boolean, commandFailed: boolean }, webhook: { enabled: boolean, url: string | null } } } | null };

export type GetDeviceSettingsQueryVariables = Exact<{
  organizationId: string;
  deviceImei: string;
}>;


export type GetDeviceSettingsQuery = { deviceSettings: { id: string, deviceImei: string, customName: string | null, location: string | null, createdAt: string | null, updatedAt: string | null, thresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } | null, effectiveThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } } | null };

export type GetOrganizationSettingsQueryVariables = Exact<{
  organizationId: string;
}>;


export type GetOrganizationSettingsQuery = { organizationSettings: { id: string, organizationId: string, dateFormat: string | null, alertCooldownMinutes: number, createdAt: string | null, updatedAt: string | null, defaultThresholds: { riskWarn: number, riskCrit: number, thermalWarn: number, thermalCrit: number, bufferWarn: number, bufferCrit: number } } | null };

export type UpdateVersionFragment = { id: string, version: string, releaseType: ReleaseType, releaseNotes: string | null, apkFilename: string, apkSize: number, sha256: string, releasedAt: string | null, createdAt: string | null, isLatest: boolean | null };

export type PushDeviceFragment = { id: string, deviceId: string, deviceName: string | null, status: DevicePushStatus, sentAt: string | null, acknowledgedAt: string | null, error: string | null };

export type UpdatePushFragment = { id: string, version: string, installType: InstallType, status: UpdatePushStatus, initiatedBy: string, initiatedAt: string | null, completedAt: string | null, cancelledAt: string | null, deviceCount: number, devices: Array<{ id: string, deviceId: string, deviceName: string | null, status: DevicePushStatus, sentAt: string | null, acknowledgedAt: string | null, error: string | null }> | null };

export type SyncStatusFragment = { status: string, lastSyncAt: string | null, nextSyncAt: string | null, versionsFound: number | null, error: string | null };

export type ChangelogEntryFragment = { version: string, date: string, type: string, notes: string };

export type PushHistoryEntryFragment = { id: string, version: string, installType: string, status: string, initiatedBy: string, initiatedAt: number, completedAt: number | null, deviceCount: number, pending: number, acknowledged: number, failed: number };

export type PushUpdateMutationVariables = Exact<{
  organizationId: string;
  version: string;
  deviceIds: Array<string | number> | string | number;
  installType: string;
  scheduledAt?: number | null | undefined;
}>;


export type PushUpdateMutation = { pushUpdate: { pushId: string, version: string, installType: string, scheduledAt: number | null, status: string, initiatedBy: string, initiatedAt: number, deviceCount: number } | null };

export type CancelUpdateMutationVariables = Exact<{
  organizationId: string;
  id: string | number;
}>;


export type CancelUpdateMutation = { cancelUpdate: { id: string, status: string, cancelledAt: number, cancelledBy: string } | null };

export type SyncUpdatesMutationVariables = Exact<{ [key: string]: never; }>;


export type SyncUpdatesMutation = { syncUpdates: { status: string, startedAt: number, message: string | null, versionsFound: number | null } | null };

export type GetUpdatesQueryVariables = Exact<{
  organizationId: string;
  status?: string | null | undefined;
  limit?: number | null | undefined;
  offset?: number | null | undefined;
}>;


export type GetUpdatesQuery = { updatesVersions: { versions: Array<{ id: string, version: string, releaseType: ReleaseType, releaseNotes: string | null, apkFilename: string, apkSize: number, sha256: string, releasedAt: string | null, createdAt: string | null, isLatest: boolean | null }>, pagination: { total: number, limit: number, offset: number, hasMore: boolean } | null } | null };

export type GetUpdatesStatusQueryVariables = Exact<{
  organizationId: string;
  deviceId?: string | number | null | undefined;
}>;


export type GetUpdatesStatusQuery = { updatesStatus: { version: string, apkFilename: string | null, sha256: string | null, sync: { status: string, lastSyncAt: string | null, nextSyncAt: string | null, versionsFound: number | null, error: string | null }, latest: { id: string, version: string, releaseType: ReleaseType, releaseNotes: string | null, apkFilename: string, apkSize: number, sha256: string, releasedAt: string | null, createdAt: string | null, isLatest: boolean | null } | null, device: { currentVersion: string, needsUpdate: boolean } } | null };

export type GetUpdatesChangelogQueryVariables = Exact<{
  organizationId: string;
  version?: string | null | undefined;
  limit?: number | null | undefined;
}>;


export type GetUpdatesChangelogQuery = { updatesChangelog: Array<{ version: string, date: string, type: string, notes: string }> | null };

export type GetUpdatesHistoryQueryVariables = Exact<{
  organizationId: string;
  status?: string | null | undefined;
  page?: number | null | undefined;
  limit?: number | null | undefined;
}>;


export type GetUpdatesHistoryQuery = { updatesHistory: { pushes: Array<{ id: string, version: string, installType: string, status: string, initiatedBy: string, initiatedAt: number, completedAt: number | null, deviceCount: number, pending: number, acknowledged: number, failed: number }> | null, pagination: { total: number, limit: number, offset: number, hasMore: boolean } } | null };

export const DeviceListFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"DeviceList"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"Device"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}},{"kind":"Field","name":{"kind":"Name","value":"online"}}]}}]} as unknown as DocumentNode<DeviceListFragment, unknown>;
export const DeviceFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"Device"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"Device"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"securityPatch"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"registeredAt"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}},{"kind":"Field","name":{"kind":"Name","value":"fcmTokenValid"}},{"kind":"Field","name":{"kind":"Name","value":"commandSecretSet"}},{"kind":"Field","name":{"kind":"Name","value":"connection"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"webSocketStatus"}},{"kind":"Field","name":{"kind":"Name","value":"connectedAt"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"clientIp"}}]}}]}}]} as unknown as DocumentNode<DeviceFragment, unknown>;
export const LogEntryFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"LogEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"LogEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"data"}}]}}]} as unknown as DocumentNode<LogEntryFragment, unknown>;
export const InboxEntryFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"InboxEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"InboxEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"firebaseInstallId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}},{"kind":"Field","name":{"kind":"Name","value":"operatorId"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"approvedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rejectedAt"}}]}}]} as unknown as DocumentNode<InboxEntryFragment, unknown>;
export const OperatorSettingsFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"OperatorSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"OperatorSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"client"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serverUrl"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"requestTimeoutMs"}},{"kind":"Field","name":{"kind":"Name","value":"logBufferLimit"}},{"kind":"Field","name":{"kind":"Name","value":"signalHistoryLimit"}},{"kind":"Field","name":{"kind":"Name","value":"autoReconnect"}},{"kind":"Field","name":{"kind":"Name","value":"strictHmac"}}]}},{"kind":"Field","name":{"kind":"Name","value":"notifications"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"channels"}},{"kind":"Field","name":{"kind":"Name","value":"email"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}}]}},{"kind":"Field","name":{"kind":"Name","value":"push"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}}]}},{"kind":"Field","name":{"kind":"Name","value":"webhook"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]}}]} as unknown as DocumentNode<OperatorSettingsFragment, unknown>;
export const DeviceSettingsFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"DeviceSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"DeviceSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceImei"}},{"kind":"Field","name":{"kind":"Name","value":"customName"}},{"kind":"Field","name":{"kind":"Name","value":"location"}},{"kind":"Field","name":{"kind":"Name","value":"thresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"effectiveThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]} as unknown as DocumentNode<DeviceSettingsFragment, unknown>;
export const OrganizationSettingsFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"OrganizationSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"OrganizationSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"organizationId"}},{"kind":"Field","name":{"kind":"Name","value":"dateFormat"}},{"kind":"Field","name":{"kind":"Name","value":"alertCooldownMinutes"}},{"kind":"Field","name":{"kind":"Name","value":"defaultThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]} as unknown as DocumentNode<OrganizationSettingsFragment, unknown>;
export const UpdateVersionFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"UpdateVersion"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateVersion"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"releaseType"}},{"kind":"Field","name":{"kind":"Name","value":"releaseNotes"}},{"kind":"Field","name":{"kind":"Name","value":"apkFilename"}},{"kind":"Field","name":{"kind":"Name","value":"apkSize"}},{"kind":"Field","name":{"kind":"Name","value":"sha256"}},{"kind":"Field","name":{"kind":"Name","value":"releasedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"isLatest"}}]}}]} as unknown as DocumentNode<UpdateVersionFragment, unknown>;
export const PushDeviceFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"PushDevice"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"PushDevice"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"sentAt"}},{"kind":"Field","name":{"kind":"Name","value":"acknowledgedAt"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]} as unknown as DocumentNode<PushDeviceFragment, unknown>;
export const UpdatePushFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"UpdatePush"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"UpdatePush"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"installType"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedBy"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"completedAt"}},{"kind":"Field","name":{"kind":"Name","value":"cancelledAt"}},{"kind":"Field","name":{"kind":"Name","value":"deviceCount"}},{"kind":"Field","name":{"kind":"Name","value":"devices"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"PushDevice"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"PushDevice"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"PushDevice"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"sentAt"}},{"kind":"Field","name":{"kind":"Name","value":"acknowledgedAt"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]} as unknown as DocumentNode<UpdatePushFragment, unknown>;
export const SyncStatusFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"SyncStatus"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"SyncStatus"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"lastSyncAt"}},{"kind":"Field","name":{"kind":"Name","value":"nextSyncAt"}},{"kind":"Field","name":{"kind":"Name","value":"versionsFound"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]} as unknown as DocumentNode<SyncStatusFragment, unknown>;
export const ChangelogEntryFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"ChangelogEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"ChangelogEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"date"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}}]}}]} as unknown as DocumentNode<ChangelogEntryFragment, unknown>;
export const PushHistoryEntryFragmentDoc = {"kind":"Document","definitions":[{"kind":"FragmentDefinition","name":{"kind":"Name","value":"PushHistoryEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"PushHistoryEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"installType"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedBy"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"completedAt"}},{"kind":"Field","name":{"kind":"Name","value":"deviceCount"}},{"kind":"Field","name":{"kind":"Name","value":"pending"}},{"kind":"Field","name":{"kind":"Name","value":"acknowledged"}},{"kind":"Field","name":{"kind":"Name","value":"failed"}}]}}]} as unknown as DocumentNode<PushHistoryEntryFragment, unknown>;
export const SendCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SendCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"command"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"args"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"JSON"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sendCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"deviceId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}}},{"kind":"Argument","name":{"kind":"Name","value":"command"},"value":{"kind":"Variable","name":{"kind":"Name","value":"command"}}},{"kind":"Argument","name":{"kind":"Name","value":"args"},"value":{"kind":"Variable","name":{"kind":"Name","value":"args"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"commandId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}}]}}]}}]} as unknown as DocumentNode<SendCommandMutation, SendCommandMutationVariables>;
export const RetryCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RetryCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"retryCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"dispatchId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"commandId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"deliveredAt"}}]}}]}}]} as unknown as DocumentNode<RetryCommandMutation, RetryCommandMutationVariables>;
export const CancelCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CancelCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cancelCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"dispatchId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"cancelledAt"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<CancelCommandMutation, CancelCommandMutationVariables>;
export const GetPendingCommandsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetPendingCommands"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"pendingCommands"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"deviceId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"commandId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"args"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"deliveredAt"}}]}}]}}]} as unknown as DocumentNode<GetPendingCommandsQuery, GetPendingCommandsQueryVariables>;
export const GetCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"command"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"dispatchId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"commandId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"args"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"deliveredAt"}}]}}]}}]} as unknown as DocumentNode<GetCommandQuery, GetCommandQueryVariables>;
export const GetDevicesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDevices"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"devices"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}},{"kind":"Field","name":{"kind":"Name","value":"online"}}]}}]}}]} as unknown as DocumentNode<GetDevicesQuery, GetDevicesQueryVariables>;
export const GetDeviceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDevice"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"device"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"securityPatch"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"registeredAt"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}},{"kind":"Field","name":{"kind":"Name","value":"fcmTokenValid"}},{"kind":"Field","name":{"kind":"Name","value":"commandSecretSet"}},{"kind":"Field","name":{"kind":"Name","value":"connection"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"webSocketStatus"}},{"kind":"Field","name":{"kind":"Name","value":"connectedAt"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"clientIp"}}]}}]}}]}}]} as unknown as DocumentNode<GetDeviceQuery, GetDeviceQueryVariables>;
export const GetDeviceCountDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDeviceCount"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceCount"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}}]}]}}]} as unknown as DocumentNode<GetDeviceCountQuery, GetDeviceCountQueryVariables>;
export const GetDeviceInspectionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDeviceInspection"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceInspection"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}},{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"identity"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}}]}},{"kind":"Field","name":{"kind":"Name","value":"software"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"securityPatch"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}}]}},{"kind":"Field","name":{"kind":"Name","value":"registration"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"registeredAt"}},{"kind":"Field","name":{"kind":"Name","value":"fcmTokenValid"}},{"kind":"Field","name":{"kind":"Name","value":"fcmTokenRefreshedAt"}},{"kind":"Field","name":{"kind":"Name","value":"commandSecretSet"}}]}},{"kind":"Field","name":{"kind":"Name","value":"connection"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"webSocketStatus"}},{"kind":"Field","name":{"kind":"Name","value":"connectedAt"}},{"kind":"Field","name":{"kind":"Name","value":"fcmStatus"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}},{"kind":"Field","name":{"kind":"Name","value":"clientIp"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}}]}},{"kind":"Field","name":{"kind":"Name","value":"telemetry"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"lastTimestamp"}},{"kind":"Field","name":{"kind":"Name","value":"framesToday"}},{"kind":"Field","name":{"kind":"Name","value":"avgLatencyMs"}},{"kind":"Field","name":{"kind":"Name","value":"totalBytesToday"}},{"kind":"Field","name":{"kind":"Name","value":"sessionsToday"}}]}}]}}]}}]} as unknown as DocumentNode<GetDeviceInspectionQuery, GetDeviceInspectionQueryVariables>;
export const GetDeviceTimelineDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDeviceTimeline"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"eventType"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"TimelineEventType"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"startTime"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"endTime"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"cursor"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceTimeline"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}},{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"eventType"},"value":{"kind":"Variable","name":{"kind":"Name","value":"eventType"}}},{"kind":"Argument","name":{"kind":"Name","value":"startTime"},"value":{"kind":"Variable","name":{"kind":"Name","value":"startTime"}}},{"kind":"Argument","name":{"kind":"Name","value":"endTime"},"value":{"kind":"Variable","name":{"kind":"Name","value":"endTime"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"cursor"},"value":{"kind":"Variable","name":{"kind":"Name","value":"cursor"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"events"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"data"}}]}},{"kind":"Field","name":{"kind":"Name","value":"hasMore"}},{"kind":"Field","name":{"kind":"Name","value":"nextCursor"}}]}}]}}]} as unknown as DocumentNode<GetDeviceTimelineQuery, GetDeviceTimelineQueryVariables>;
export const GetLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"type"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"startTime"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"endTime"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"cursor"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}},{"kind":"Argument","name":{"kind":"Name","value":"type"},"value":{"kind":"Variable","name":{"kind":"Name","value":"type"}}},{"kind":"Argument","name":{"kind":"Name","value":"startTime"},"value":{"kind":"Variable","name":{"kind":"Name","value":"startTime"}}},{"kind":"Argument","name":{"kind":"Name","value":"endTime"},"value":{"kind":"Variable","name":{"kind":"Name","value":"endTime"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"cursor"},"value":{"kind":"Variable","name":{"kind":"Name","value":"cursor"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"events"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"LogEntry"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pagination"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"hasMore"}},{"kind":"Field","name":{"kind":"Name","value":"nextCursor"}}]}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"LogEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"LogEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"data"}}]}}]} as unknown as DocumentNode<GetLogsQuery, GetLogsQueryVariables>;
export const OnDeviceUpdatedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"OnDeviceUpdated"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceUpdated"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"deviceId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"deviceName"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"lastSeen"}}]}}]}}]} as unknown as DocumentNode<OnDeviceUpdatedSubscription, OnDeviceUpdatedSubscriptionVariables>;
export const OnTelemetryReceivedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"OnTelemetryReceived"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"telemetryReceived"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"deviceId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"receivedAt"}},{"kind":"Field","name":{"kind":"Name","value":"riskScore"}},{"kind":"Field","name":{"kind":"Name","value":"bufferLevel"}},{"kind":"Field","name":{"kind":"Name","value":"thermalTemp"}},{"kind":"Field","name":{"kind":"Name","value":"payload"}}]}}]}}]} as unknown as DocumentNode<OnTelemetryReceivedSubscription, OnTelemetryReceivedSubscriptionVariables>;
export const OnCommandStatusChangedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"OnCommandStatusChanged"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"commandStatusChanged"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"dispatchId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"dispatchId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"dispatchId"}},{"kind":"Field","name":{"kind":"Name","value":"commandId"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<OnCommandStatusChangedSubscription, OnCommandStatusChangedSubscriptionVariables>;
export const OnOrganizationEventDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"OnOrganizationEvent"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"orgId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"organizationEvent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"orgId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"orgId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"data"}}]}}]}}]} as unknown as DocumentNode<OnOrganizationEventSubscription, OnOrganizationEventSubscriptionVariables>;
export const OnMemberEventDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"OnMemberEvent"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"orgId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"memberEvent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"orgId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"orgId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"memberId"}},{"kind":"Field","name":{"kind":"Name","value":"data"}}]}}]}}]} as unknown as DocumentNode<OnMemberEventSubscription, OnMemberEventSubscriptionVariables>;
export const AckInboxDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AckInbox"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"action"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AckAction"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"notes"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"ackInbox"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}},{"kind":"Argument","name":{"kind":"Name","value":"action"},"value":{"kind":"Variable","name":{"kind":"Name","value":"action"}}},{"kind":"Argument","name":{"kind":"Name","value":"notes"},"value":{"kind":"Variable","name":{"kind":"Name","value":"notes"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"approvedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rejectedAt"}},{"kind":"Field","name":{"kind":"Name","value":"commandSecret"}},{"kind":"Field","name":{"kind":"Name","value":"fcmPushSent"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}}]}}]}}]} as unknown as DocumentNode<AckInboxMutation, AckInboxMutationVariables>;
export const DeregisterDeviceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeregisterDevice"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hard"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deregisterDevice"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}},{"kind":"Argument","name":{"kind":"Name","value":"hard"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hard"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"deregisteredAt"}},{"kind":"Field","name":{"kind":"Name","value":"retentionUntil"}}]}}]}}]} as unknown as DocumentNode<DeregisterDeviceMutation, DeregisterDeviceMutationVariables>;
export const GetInboxEntriesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetInboxEntries"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"status"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"page"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inbox"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"status"},"value":{"kind":"Variable","name":{"kind":"Name","value":"status"}}},{"kind":"Argument","name":{"kind":"Name","value":"page"},"value":{"kind":"Variable","name":{"kind":"Name","value":"page"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"requests"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"InboxEntry"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pagination"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"total"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"offset"}},{"kind":"Field","name":{"kind":"Name","value":"hasMore"}}]}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"InboxEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"InboxEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"firebaseInstallId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}},{"kind":"Field","name":{"kind":"Name","value":"operatorId"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"approvedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rejectedAt"}}]}}]} as unknown as DocumentNode<GetInboxEntriesQuery, GetInboxEntriesQueryVariables>;
export const GetInboxEntryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetInboxEntry"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"imei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inboxEntry"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"imei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"imei"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"InboxEntry"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"InboxEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"InboxEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imei"}},{"kind":"Field","name":{"kind":"Name","value":"model"}},{"kind":"Field","name":{"kind":"Name","value":"manufacturer"}},{"kind":"Field","name":{"kind":"Name","value":"osVersion"}},{"kind":"Field","name":{"kind":"Name","value":"appVersion"}},{"kind":"Field","name":{"kind":"Name","value":"firebaseInstallId"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}},{"kind":"Field","name":{"kind":"Name","value":"operatorId"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"approvedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rejectedAt"}}]}}]} as unknown as DocumentNode<GetInboxEntryQuery, GetInboxEntryQueryVariables>;
export const UpdateNotificationsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateNotifications"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateNotificationsInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateMyNotifications"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"channels"}},{"kind":"Field","name":{"kind":"Name","value":"email"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}},{"kind":"Field","name":{"kind":"Name","value":"registrationRequest"}}]}},{"kind":"Field","name":{"kind":"Name","value":"push"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}},{"kind":"Field","name":{"kind":"Name","value":"registrationRequest"}}]}},{"kind":"Field","name":{"kind":"Name","value":"webhook"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"types"}}]}}]}}]}}]} as unknown as DocumentNode<UpdateNotificationsMutation, UpdateNotificationsMutationVariables>;
export const UpdateDeviceSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateDeviceSettings"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceImei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateDeviceSettingsInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateDeviceSettings"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"deviceImei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceImei"}}},{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceImei"}},{"kind":"Field","name":{"kind":"Name","value":"customName"}},{"kind":"Field","name":{"kind":"Name","value":"location"}},{"kind":"Field","name":{"kind":"Name","value":"thresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"effectiveThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<UpdateDeviceSettingsMutation, UpdateDeviceSettingsMutationVariables>;
export const UpdateOrganizationSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateOrganizationSettings"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateOrganizationSettingsInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateOrganizationSettings"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"organizationId"}},{"kind":"Field","name":{"kind":"Name","value":"timezone"}},{"kind":"Field","name":{"kind":"Name","value":"dateFormat"}},{"kind":"Field","name":{"kind":"Name","value":"alertCooldownMinutes"}},{"kind":"Field","name":{"kind":"Name","value":"defaultThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]}}]} as unknown as DocumentNode<UpdateOrganizationSettingsMutation, UpdateOrganizationSettingsMutationVariables>;
export const GetSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetSettings"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"mySettings"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"OperatorSettings"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"OperatorSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"OperatorSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"client"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serverUrl"}},{"kind":"Field","name":{"kind":"Name","value":"deviceId"}},{"kind":"Field","name":{"kind":"Name","value":"requestTimeoutMs"}},{"kind":"Field","name":{"kind":"Name","value":"logBufferLimit"}},{"kind":"Field","name":{"kind":"Name","value":"signalHistoryLimit"}},{"kind":"Field","name":{"kind":"Name","value":"autoReconnect"}},{"kind":"Field","name":{"kind":"Name","value":"strictHmac"}}]}},{"kind":"Field","name":{"kind":"Name","value":"notifications"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"channels"}},{"kind":"Field","name":{"kind":"Name","value":"email"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}}]}},{"kind":"Field","name":{"kind":"Name","value":"push"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"thresholdBreach"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOffline"}},{"kind":"Field","name":{"kind":"Name","value":"deviceOnline"}},{"kind":"Field","name":{"kind":"Name","value":"updateAvailable"}},{"kind":"Field","name":{"kind":"Name","value":"commandFailed"}}]}},{"kind":"Field","name":{"kind":"Name","value":"webhook"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]}}]} as unknown as DocumentNode<GetSettingsQuery, GetSettingsQueryVariables>;
export const GetDeviceSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetDeviceSettings"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceImei"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deviceSettings"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"deviceImei"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceImei"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"DeviceSettings"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"DeviceSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"DeviceSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"deviceImei"}},{"kind":"Field","name":{"kind":"Name","value":"customName"}},{"kind":"Field","name":{"kind":"Name","value":"location"}},{"kind":"Field","name":{"kind":"Name","value":"thresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"effectiveThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]} as unknown as DocumentNode<GetDeviceSettingsQuery, GetDeviceSettingsQueryVariables>;
export const GetOrganizationSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetOrganizationSettings"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"organizationSettings"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"OrganizationSettings"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"OrganizationSettings"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"OrganizationSettings"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"organizationId"}},{"kind":"Field","name":{"kind":"Name","value":"dateFormat"}},{"kind":"Field","name":{"kind":"Name","value":"alertCooldownMinutes"}},{"kind":"Field","name":{"kind":"Name","value":"defaultThresholds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"riskWarn"}},{"kind":"Field","name":{"kind":"Name","value":"riskCrit"}},{"kind":"Field","name":{"kind":"Name","value":"thermalWarn"}},{"kind":"Field","name":{"kind":"Name","value":"thermalCrit"}},{"kind":"Field","name":{"kind":"Name","value":"bufferWarn"}},{"kind":"Field","name":{"kind":"Name","value":"bufferCrit"}}]}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"updatedAt"}}]}}]} as unknown as DocumentNode<GetOrganizationSettingsQuery, GetOrganizationSettingsQueryVariables>;
export const PushUpdateDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"PushUpdate"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"version"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceIds"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"installType"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"scheduledAt"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"pushUpdate"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"version"},"value":{"kind":"Variable","name":{"kind":"Name","value":"version"}}},{"kind":"Argument","name":{"kind":"Name","value":"deviceIds"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceIds"}}},{"kind":"Argument","name":{"kind":"Name","value":"installType"},"value":{"kind":"Variable","name":{"kind":"Name","value":"installType"}}},{"kind":"Argument","name":{"kind":"Name","value":"scheduledAt"},"value":{"kind":"Variable","name":{"kind":"Name","value":"scheduledAt"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"pushId"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"installType"}},{"kind":"Field","name":{"kind":"Name","value":"scheduledAt"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedBy"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"deviceCount"}}]}}]}}]} as unknown as DocumentNode<PushUpdateMutation, PushUpdateMutationVariables>;
export const CancelUpdateDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CancelUpdate"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cancelUpdate"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"cancelledAt"}},{"kind":"Field","name":{"kind":"Name","value":"cancelledBy"}}]}}]}}]} as unknown as DocumentNode<CancelUpdateMutation, CancelUpdateMutationVariables>;
export const SyncUpdatesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SyncUpdates"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"syncUpdates"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"versionsFound"}}]}}]}}]} as unknown as DocumentNode<SyncUpdatesMutation, SyncUpdatesMutationVariables>;
export const GetUpdatesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetUpdates"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"status"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updatesVersions"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"status"},"value":{"kind":"Variable","name":{"kind":"Name","value":"status"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"versions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"UpdateVersion"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pagination"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"total"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"offset"}},{"kind":"Field","name":{"kind":"Name","value":"hasMore"}}]}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"UpdateVersion"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateVersion"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"releaseType"}},{"kind":"Field","name":{"kind":"Name","value":"releaseNotes"}},{"kind":"Field","name":{"kind":"Name","value":"apkFilename"}},{"kind":"Field","name":{"kind":"Name","value":"apkSize"}},{"kind":"Field","name":{"kind":"Name","value":"sha256"}},{"kind":"Field","name":{"kind":"Name","value":"releasedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"isLatest"}}]}}]} as unknown as DocumentNode<GetUpdatesQuery, GetUpdatesQueryVariables>;
export const GetUpdatesStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetUpdatesStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updatesStatus"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"deviceId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"deviceId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"sync"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"lastSyncAt"}},{"kind":"Field","name":{"kind":"Name","value":"nextSyncAt"}},{"kind":"Field","name":{"kind":"Name","value":"versionsFound"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}},{"kind":"Field","name":{"kind":"Name","value":"latest"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"UpdateVersion"}}]}},{"kind":"Field","name":{"kind":"Name","value":"device"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"currentVersion"}},{"kind":"Field","name":{"kind":"Name","value":"needsUpdate"}}]}},{"kind":"Field","name":{"kind":"Name","value":"apkFilename"}},{"kind":"Field","name":{"kind":"Name","value":"sha256"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"UpdateVersion"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateVersion"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"releaseType"}},{"kind":"Field","name":{"kind":"Name","value":"releaseNotes"}},{"kind":"Field","name":{"kind":"Name","value":"apkFilename"}},{"kind":"Field","name":{"kind":"Name","value":"apkSize"}},{"kind":"Field","name":{"kind":"Name","value":"sha256"}},{"kind":"Field","name":{"kind":"Name","value":"releasedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"isLatest"}}]}}]} as unknown as DocumentNode<GetUpdatesStatusQuery, GetUpdatesStatusQueryVariables>;
export const GetUpdatesChangelogDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetUpdatesChangelog"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"version"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updatesChangelog"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"version"},"value":{"kind":"Variable","name":{"kind":"Name","value":"version"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"ChangelogEntry"}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"ChangelogEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"ChangelogEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"date"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"notes"}}]}}]} as unknown as DocumentNode<GetUpdatesChangelogQuery, GetUpdatesChangelogQueryVariables>;
export const GetUpdatesHistoryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetUpdatesHistory"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"status"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"page"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updatesHistory"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"organizationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"organizationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"status"},"value":{"kind":"Variable","name":{"kind":"Name","value":"status"}}},{"kind":"Argument","name":{"kind":"Name","value":"page"},"value":{"kind":"Variable","name":{"kind":"Name","value":"page"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"pushes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"FragmentSpread","name":{"kind":"Name","value":"PushHistoryEntry"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pagination"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"total"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"offset"}},{"kind":"Field","name":{"kind":"Name","value":"hasMore"}}]}}]}}]}},{"kind":"FragmentDefinition","name":{"kind":"Name","value":"PushHistoryEntry"},"typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"PushHistoryEntry"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"installType"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedBy"}},{"kind":"Field","name":{"kind":"Name","value":"initiatedAt"}},{"kind":"Field","name":{"kind":"Name","value":"completedAt"}},{"kind":"Field","name":{"kind":"Name","value":"deviceCount"}},{"kind":"Field","name":{"kind":"Name","value":"pending"}},{"kind":"Field","name":{"kind":"Name","value":"acknowledged"}},{"kind":"Field","name":{"kind":"Name","value":"failed"}}]}}]} as unknown as DocumentNode<GetUpdatesHistoryQuery, GetUpdatesHistoryQueryVariables>;