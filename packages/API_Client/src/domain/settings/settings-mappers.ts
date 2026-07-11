import type {
  Thresholds,
  ClientSettings,
  EmailNotifications,
  PushNotifications,
  WebhookNotifications,
  NotificationSettings,
  SecuritySettings,
  NotificationEvent,
} from "./settings-entity";

export interface RawThresholds {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

export interface RawClientSettings {
  serverUrl: string;
  deviceId: string;
  requestTimeoutMs: number;
  logBufferLimit: number;
  signalHistoryLimit: number;
  autoReconnect: boolean;
  strictHmac: boolean;
  notificationsEnabled: boolean;
}

export interface RawEmailNotifications {
  thresholdBreach: boolean;
  deviceOffline: boolean;
  deviceOnline: boolean;
  updateAvailable: boolean;
  commandFailed: boolean;
  registrationRequest: boolean;
}

export interface RawPushNotifications {
  thresholdBreach: boolean;
  deviceOffline: boolean;
  deviceOnline: boolean;
  updateAvailable: boolean;
  commandFailed: boolean;
  registrationRequest: boolean;
}

export interface RawWebhookNotifications {
  enabled: boolean;
  url: string;
  secret: string;
  types: string[];
}

export interface RawNotificationSettings {
  enabled: boolean;
  channels: string[];
  email: RawEmailNotifications;
  push: RawPushNotifications;
  webhook: RawWebhookNotifications;
}

export type RawNotificationSettingsResponse = RawNotificationSettings;

export interface RawSecuritySettings {
  maxConcurrentSessions: number;
  passwordMinAgeDays: number;
  passwordMaxAgeDays: number;
  passwordHistoryCount: number;
  sessionPinRequired: boolean;
}

export function thresholdsFromRaw(raw: RawThresholds): Thresholds {
  return {
    riskWarn: raw.riskWarn,
    riskCrit: raw.riskCrit,
    thermalWarn: raw.thermalWarn,
    thermalCrit: raw.thermalCrit,
    bufferWarn: raw.bufferWarn,
    bufferCrit: raw.bufferCrit,
  };
}

export function clientSettingsFromRaw(raw: RawClientSettings): ClientSettings {
  return {
    serverUrl: raw.serverUrl,
    deviceId: raw.deviceId,
    requestTimeoutMs: raw.requestTimeoutMs,
    logBufferLimit: raw.logBufferLimit,
    signalHistoryLimit: raw.signalHistoryLimit,
    autoReconnect: raw.autoReconnect,
    strictHmac: raw.strictHmac,
    notificationsEnabled: raw.notificationsEnabled,
  };
}

export function emailNotificationsFromRaw(raw: RawEmailNotifications): EmailNotifications {
  return {
    thresholdBreach: raw.thresholdBreach,
    deviceOffline: raw.deviceOffline,
    deviceOnline: raw.deviceOnline,
    updateAvailable: raw.updateAvailable,
    commandFailed: raw.commandFailed,
    registrationRequest: raw.registrationRequest,
  };
}

export function pushNotificationsFromRaw(raw: RawPushNotifications): PushNotifications {
  return {
    thresholdBreach: raw.thresholdBreach,
    deviceOffline: raw.deviceOffline,
    deviceOnline: raw.deviceOnline,
    updateAvailable: raw.updateAvailable,
    commandFailed: raw.commandFailed,
    registrationRequest: raw.registrationRequest,
  };
}

export function webhookNotificationsFromRaw(raw: RawWebhookNotifications): WebhookNotifications {
  return {
    enabled: raw.enabled,
    url: raw.url,
    secret: raw.secret,
    types: raw.types as NotificationEvent[],
  };
}

export function notificationSettingsFromRaw(raw: RawNotificationSettings): NotificationSettings {
  return {
    enabled: raw.enabled,
    channels: raw.channels,
    email: emailNotificationsFromRaw(raw.email),
    push: pushNotificationsFromRaw(raw.push),
    webhook: webhookNotificationsFromRaw(raw.webhook),
  };
}

export function securitySettingsFromRaw(raw: RawSecuritySettings): SecuritySettings {
  return {
    maxConcurrentSessions: raw.maxConcurrentSessions,
    passwordMinAgeDays: raw.passwordMinAgeDays,
    passwordMaxAgeDays: raw.passwordMaxAgeDays,
    passwordHistoryCount: raw.passwordHistoryCount,
    sessionPinRequired: raw.sessionPinRequired,
  };
}

export function thresholdsToRaw(thresholds: Partial<Thresholds>): Partial<RawThresholds> {
  return {
    riskWarn: thresholds.riskWarn,
    riskCrit: thresholds.riskCrit,
    thermalWarn: thresholds.thermalWarn,
    thermalCrit: thresholds.thermalCrit,
    bufferWarn: thresholds.bufferWarn,
    bufferCrit: thresholds.bufferCrit,
  };
}

export function clientSettingsToRaw(client: Partial<ClientSettings>): Partial<RawClientSettings> {
  return {
    serverUrl: client.serverUrl,
    deviceId: client.deviceId,
    requestTimeoutMs: client.requestTimeoutMs,
    logBufferLimit: client.logBufferLimit,
    signalHistoryLimit: client.signalHistoryLimit,
    autoReconnect: client.autoReconnect,
    strictHmac: client.strictHmac,
    notificationsEnabled: client.notificationsEnabled,
  };
}

export function emailNotificationsToRaw(email: Partial<EmailNotifications>): Partial<RawEmailNotifications> {
  return {
    thresholdBreach: email.thresholdBreach,
    deviceOffline: email.deviceOffline,
    deviceOnline: email.deviceOnline,
    updateAvailable: email.updateAvailable,
    commandFailed: email.commandFailed,
    registrationRequest: email.registrationRequest,
  };
}

export function pushNotificationsToRaw(push: Partial<PushNotifications>): Partial<RawPushNotifications> {
  return {
    thresholdBreach: push.thresholdBreach,
    deviceOffline: push.deviceOffline,
    deviceOnline: push.deviceOnline,
    updateAvailable: push.updateAvailable,
    commandFailed: push.commandFailed,
    registrationRequest: push.registrationRequest,
  };
}

export function webhookNotificationsToRaw(webhook: Partial<WebhookNotifications>): Partial<RawWebhookNotifications> {
  return {
    enabled: webhook.enabled,
    url: webhook.url,
    types: webhook.types,
  };
}

export function notificationSettingsToRaw(settings: Partial<NotificationSettings>): Partial<RawNotificationSettings> {
  return {
    enabled: settings.enabled,
    channels: settings.channels,
    email: settings.email ? emailNotificationsToRaw(settings.email) : undefined,
    push: settings.push ? pushNotificationsToRaw(settings.push) : undefined,
    webhook: settings.webhook ? webhookNotificationsToRaw(settings.webhook) : undefined,
  };
}
