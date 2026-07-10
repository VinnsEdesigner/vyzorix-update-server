/**
 * Settings Mappers
 * 
 * Transformations from raw API response to domain types.
 * Raw API uses snake_case, domain uses camelCase.
 */

import type {
  ConnectionSettings,
  ThresholdSettings,
  NotificationSettings,
  OperatorInfo,
  AdvancedSettings,
  Settings,
  NotificationChannel,
} from "./settings-entity";

// ============================================================================
// Raw API Types (snake_case)
// ============================================================================

/**
 * Raw connection settings from API
 */
export interface RawConnectionSettings {
  server_url?: string;
  device_id?: string;
  dashboard_token?: string;
  request_timeout_ms?: number;
  auto_reconnect?: boolean;
  strict_hmac?: boolean;
}

/**
 * Raw threshold settings from API
 */
export interface RawThresholdSettings {
  risk_warn?: number;
  risk_crit?: number;
  thermal_warn?: number;
  thermal_crit?: number;
  buffer_warn?: number;
  buffer_crit?: number;
}

/**
 * Raw notification channel from API
 */
export interface RawNotificationChannel {
  enabled?: boolean;
  email?: string;
  webhook_url?: string;
  webhook_secret?: string;
}

/**
 * Raw notification settings from API
 */
export interface RawNotificationSettings {
  enabled?: boolean;
  email?: RawNotificationChannel;
  push?: RawNotificationChannel;
  webhook?: RawNotificationChannel;
  events?: Record<string, boolean>;
}

/**
 * Raw operator info from API
 */
export interface RawOperatorInfo {
  id?: string;
  email?: string;
  name?: string;
  role?: string;
  email_verified?: boolean;
  created_at?: string | number;
}

/**
 * Raw advanced settings from API
 */
export interface RawAdvancedSettings {
  log_buffer_limit?: number;
  signal_history_limit?: number;
}

/**
 * Raw complete settings from API
 */
export interface RawSettings {
  connection?: RawConnectionSettings;
  thresholds?: RawThresholdSettings;
  notifications?: RawNotificationSettings;
  operator?: RawOperatorInfo;
  advanced?: RawAdvancedSettings;
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform raw connection settings to domain
 */
export function connectionFromRaw(raw?: RawConnectionSettings | null): ConnectionSettings | null {
  if (!raw) return null;
  
  return {
    serverUrl: raw.server_url ?? "",
    deviceId: raw.device_id ?? "",
    dashboardToken: raw.dashboard_token ?? "",
    requestTimeoutMs: raw.request_timeout_ms ?? 8000,
    autoReconnect: raw.auto_reconnect ?? true,
    strictHmac: raw.strict_hmac ?? false,
  };
}

/**
 * Transform raw threshold settings to domain
 */
export function thresholdsFromRaw(raw?: RawThresholdSettings | null): ThresholdSettings | null {
  if (!raw) return null;
  
  return {
    riskWarn: raw.risk_warn ?? 70,
    riskCrit: raw.risk_crit ?? 85,
    thermalWarn: raw.thermal_warn ?? 45,
    thermalCrit: raw.thermal_crit ?? 50,
    bufferWarn: raw.buffer_warn ?? 30,
    bufferCrit: raw.buffer_crit ?? 15,
  };
}

/**
 * Transform raw notification channel to domain
 */
export function notificationChannelFromRaw(raw?: RawNotificationChannel | null): NotificationChannel | null {
  if (!raw) return null;
  
  return {
    enabled: raw.enabled ?? false,
    email: raw.email,
    webhookUrl: raw.webhook_url,
    webhookSecret: raw.webhook_secret,
  };
}

/**
 * Transform raw notification settings to domain
 */
export function notificationsFromRaw(raw?: RawNotificationSettings | null): NotificationSettings | null {
  if (!raw) return null;
  
  return {
    enabled: raw.enabled ?? true,
    email: notificationChannelFromRaw(raw.email) ?? undefined,
    push: notificationChannelFromRaw(raw.push) ?? undefined,
    webhook: notificationChannelFromRaw(raw.webhook) ?? undefined,
    events: raw.events ?? {},
  };
}

/**
 * Transform raw operator info to domain
 */
export function operatorFromRaw(raw?: RawOperatorInfo | null): OperatorInfo | null {
  if (!raw) return null;
  
  return {
    id: raw.id ?? "",
    email: raw.email ?? "",
    name: raw.name ?? "",
    role: (raw.role as "operator" | "admin") ?? "operator",
    emailVerified: raw.email_verified ?? false,
    createdAt: raw.created_at 
      ? new Date(typeof raw.created_at === "number"" ? raw.created_at * 1000 : raw.created_at)
      : new Date(),
  };
}

/**
 * Transform raw advanced settings to domain
 */
export function advancedFromRaw(raw?: RawAdvancedSettings | null): AdvancedSettings | null {
  if (!raw) return null;
  
  return {
    logBufferLimit: raw.log_buffer_limit ?? 500,
    signalHistoryLimit: raw.signal_history_limit ?? 240,
  };
}

/**
 * Transform raw settings to domain
 */
export function settingsFromRaw(raw?: RawSettings | null): Settings | null {
  if (!raw) return null;
  
  return {
    operator: operatorFromRaw(raw.operator) ?? {
      id: "",
      email: "",
      name: "",
      role: "operator",
      emailVerified: false,
      createdAt: new Date(),
    },
    connection: connectionFromRaw(raw.connection) ?? {
      serverUrl: "",
      deviceId: "",
      dashboardToken: "",
      requestTimeoutMs: 8000,
      autoReconnect: true,
      strictHmac: false,
    },
    thresholds: thresholdsFromRaw(raw.thresholds) ?? {
      riskWarn: 70,
      riskCrit: 85,
      thermalWarn: 45,
      thermalCrit: 50,
      bufferWarn: 30,
      bufferCrit: 15,
    },
    notifications: notificationsFromRaw(raw.notifications) ?? {
      enabled: true,
      events: {},
    },
    advanced: advancedFromRaw(raw.advanced) ?? {
      logBufferLimit: 500,
      signalHistoryLimit: 240,
    },
  };
}

// ============================================================================
// Reverse Transform (Domain â API)
// ============================================================================

/**
 * Transform connection settings to API format
 */
export function connectionToRaw(settings: ConnectionSettings): RawConnectionSettings {
  return {
    server_url: settings.serverUrl,
    device_id: settings.deviceId,
    dashboard_token: settings.dashboardToken,
    request_timeout_ms: settings.requestTimeoutMs,
    auto_reconnect: settings.autoReconnect,
    strict_hmac: settings.strictHmac,
  };
}

/**
 * Transform threshold settings to API format
 */
export function thresholdsToRaw(settings: ThresholdSettings): RawThresholdSettings {
  return {
    risk_warn: settings.riskWarn,
    risk_crit: settings.riskCrit,
    thermal_warn: settings.thermalWarn,
    thermal_crit: settings.thermalCrit,
    buffer_warn: settings.bufferWarn,
    buffer_crit: settings.bufferCrit,
  };
}

/**
 * Transform notification channel to API format
 */
export function notificationChannelToRaw(channel: NotificationChannel): RawNotificationChannel {
  return {
    enabled: channel.enabled,
    email: channel.email,
    webhook_url: channel.webhookUrl,
    webhook_secret: channel.webhookSecret,
  };
}

/**
 * Transform notification settings to API format
 */
export function notificationsToRaw(settings: NotificationSettings): RawNotificationSettings {
  return {
    enabled: settings.enabled,
    email: settings.email ? notificationChannelToRaw(settings.email) : undefined,
    push: settings.push ? notificationChannelToRaw(settings.push) : undefined,
    webhook: settings.webhook ? notificationChannelToRaw(settings.webhook) : undefined,
    events: settings.events,
  };
}

/**
 * Transform advanced settings to API format
 */
export function advancedToRaw(settings: AdvancedSettings): RawAdvancedSettings {
  return {
    log_buffer_limit: settings.logBufferLimit,
    signal_history_limit: settings.signalHistoryLimit,
  };
}
