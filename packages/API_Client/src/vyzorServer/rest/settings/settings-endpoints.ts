/**
 * Settings REST Endpoints
 * 
 * REST API client for operator settings operations.
 * Based on SERVER_BACKEND_SETTINGS_API.md specification.
 * Uses session-based authentication (cookies).
 */

import { restClient } from "../_shared/rest-client";

// ============================================================================
// API Paths (from SERVER_BACKEND_SETTINGS_API.md)
// ============================================================================

/**
 * Settings API paths
 */
export const SETTINGS_PATHS = {
  // GET/PATCH /v1/auth/me/settings - Complete settings
  settings: "/v1/auth/me/settings",
  // GET/PATCH /v1/auth/me/thresholds - Alert thresholds
  thresholds: "/v1/auth/me/thresholds",
  // GET/PATCH /v1/auth/me/notifications - Notification settings
  notifications: "/v1/auth/me/notifications",
  // POST /v1/auth/me/notifications/webhook/test - Test webhook
  webhookTest: "/v1/auth/me/notifications/webhook/test",
  // POST /v1/auth/me/notifications/webhook/rotate - Rotate webhook secret
  webhookRotate: "/v1/auth/me/notifications/webhook/rotate",
  // GET/PATCH /v1/auth/me - Operator profile
  operator: "/v1/auth/me",
} as const;

// ============================================================================
// Raw Response Types (snake_case from API)
// ============================================================================

/**
 * Raw client settings from API
 */
export interface RawClientSettings {
  server_url?: string;
  device_id?: string;
  request_timeout_ms?: number;
  auto_reconnect?: boolean;
  strict_hmac?: boolean;
  log_buffer_limit?: number;
  signal_history_limit?: number;
}

/**
 * Raw thresholds from API
 */
export interface RawThresholds {
  risk_warn?: number;
  risk_crit?: number;
  thermal_warn?: number;
  thermal_crit?: number;
  buffer_warn?: number;
  buffer_crit?: number;
}

/**
 * Raw notification types (email/push) from API
 */
export interface RawNotificationTypes {
  threshold_breach?: boolean;
  device_offline?: boolean;
  device_online?: boolean;
  update_available?: boolean;
  command_failed?: boolean;
  registration_request?: boolean;
}

/**
 * Raw webhook settings from API
 */
export interface RawWebhookSettings {
  enabled?: boolean;
  url?: string;
  secret?: string;
  types?: string[];
}

/**
 * Raw notification settings from API
 */
export interface RawNotificationSettings {
  enabled?: boolean;
  channels?: string[];
  email?: RawNotificationTypes;
  push?: RawNotificationTypes;
  webhook?: RawWebhookSettings;
}

/**
 * Raw complete settings from API
 */
export interface RawSettings {
  client?: RawClientSettings;
  thresholds?: RawThresholds;
  notifications?: RawNotificationSettings;
}

/**
 * Raw operator from API
 */
export interface RawOperator {
  id?: string;
  email?: string;
  name?: string;
  role?: string;
  permissions?: string[];
  created_at?: string | number;
}

// ============================================================================
// Transformed Types (camelCase for domain)
// ============================================================================

/**
 * Client settings (connection + advanced combined)
 */
export interface ClientSettings {
  serverUrl: string;
  deviceId: string;
  requestTimeoutMs: number;
  autoReconnect: boolean;
  strictHmac: boolean;
  logBufferLimit: number;
  signalHistoryLimit: number;
}

/**
 * Threshold settings
 */
export interface ThresholdSettings {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

/**
 * Notification event types
 */
export interface NotificationTypes {
  thresholdBreach: boolean;
  deviceOffline: boolean;
  deviceOnline: boolean;
  updateAvailable: boolean;
  commandFailed: boolean;
  registrationRequest: boolean;
}

/**
 * Webhook settings
 */
export interface WebhookSettings {
  enabled: boolean;
  url: string;
  secret: string;
  types: string[];
}

/**
 * Notification settings
 */
export interface NotificationSettings {
  enabled: boolean;
  channels: string[];
  email: NotificationTypes;
  push: NotificationTypes;
  webhook: WebhookSettings;
}

/**
 * Complete settings
 */
export interface Settings {
  client: ClientSettings;
  thresholds: ThresholdSettings;
  notifications: NotificationSettings;
}

/**
 * Operator profile
 */
export interface Operator {
  id: string;
  email: string;
  name: string;
  role: string;
  permissions: string[];
  createdAt: Date;
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform raw client settings to domain
 */
function clientSettingsFromRaw(raw?: RawClientSettings | null): ClientSettings {
  if (!raw) {
    return {
      serverUrl: "",
      deviceId: "",
      requestTimeoutMs: 8000,
      autoReconnect: true,
      strictHmac: false,
      logBufferLimit: 500,
      signalHistoryLimit: 240,
    };
  }
  
  return {
    serverUrl: raw.server_url ?? "",
    deviceId: raw.device_id ?? "",
    requestTimeoutMs: raw.request_timeout_ms ?? 8000,
    autoReconnect: raw.auto_reconnect ?? true,
    strictHmac: raw.strict_hmac ?? false,
    logBufferLimit: raw.log_buffer_limit ?? 500,
    signalHistoryLimit: raw.signal_history_limit ?? 240,
  };
}

/**
 * Transform raw thresholds to domain
 */
function thresholdsFromRaw(raw?: RawThresholds | null): ThresholdSettings {
  if (!raw) {
    return {
      riskWarn: 70,
      riskCrit: 85,
      thermalWarn: 45,
      thermalCrit: 50,
      bufferWarn: 30,
      bufferCrit: 15,
    };
  }
  
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
 * Transform raw notification types to domain
 */
function notificationTypesFromRaw(raw?: RawNotificationTypes | null): NotificationTypes {
  if (!raw) {
    return {
      thresholdBreach: false,
      deviceOffline: false,
      deviceOnline: false,
      updateAvailable: false,
      commandFailed: false,
      registrationRequest: false,
    };
  }
  
  return {
    thresholdBreach: raw.threshold_breach ?? false,
    deviceOffline: raw.device_offline ?? false,
    deviceOnline: raw.device_online ?? false,
    updateAvailable: raw.update_available ?? false,
    commandFailed: raw.command_failed ?? false,
    registrationRequest: raw.registration_request ?? false,
  };
}

/**
 * Transform raw webhook settings to domain
 */
function webhookSettingsFromRaw(raw?: RawWebhookSettings | null): WebhookSettings {
  if (!raw) {
    return {
      enabled: false,
      url: "",
      secret: "",
      types: [],
    };
  }
  
  return {
    enabled: raw.enabled ?? false,
    url: raw.url ?? "",
    secret: raw.secret ?? "",
    types: raw.types ?? [],
  };
}

/**
 * Transform raw notification settings to domain
 */
function notificationSettingsFromRaw(raw?: RawNotificationSettings | null): NotificationSettings {
  if (!raw) {
    return {
      enabled: true,
      channels: [],
      email: {
        thresholdBreach: false,
        deviceOffline: false,
        deviceOnline: false,
        updateAvailable: false,
        commandFailed: false,
        registrationRequest: false,
      },
      push: {
        thresholdBreach: false,
        deviceOffline: false,
        deviceOnline: false,
        updateAvailable: false,
        commandFailed: false,
        registrationRequest: false,
      },
      webhook: {
        enabled: false,
        url: "",
        secret: "",
        types: [],
      },
    };
  }
  
  return {
    enabled: raw.enabled ?? true,
    channels: raw.channels ?? [],
    email: notificationTypesFromRaw(raw.email),
    push: notificationTypesFromRaw(raw.push),
    webhook: webhookSettingsFromRaw(raw.webhook),
  };
}

/**
 * Transform raw settings to domain
 */
export function settingsFromRaw(raw?: RawSettings | null): Settings {
  if (!raw) {
    return {
      client: clientSettingsFromRaw(),
      thresholds: thresholdsFromRaw(),
      notifications: notificationSettingsFromRaw(),
    };
  }
  
  return {
    client: clientSettingsFromRaw(raw.client),
    thresholds: thresholdsFromRaw(raw.thresholds),
    notifications: notificationSettingsFromRaw(raw.notifications),
  };
}

/**
 * Transform raw operator to domain
 */
export function operatorFromRaw(raw?: RawOperator | null): Operator | null {
  if (!raw) return null;
  
  const timestamp = raw.created_at
    ? new Date(typeof raw.created_at === "number" ? raw.created_at * 1000 : raw.created_at)
    : new Date();
  
  return {
    id: raw.id ?? "",
    email: raw.email ?? "",
    name: raw.name ?? "",
    role: raw.role ?? "operator",
    permissions: raw.permissions ?? [],
    createdAt: timestamp,
  };
}

// ============================================================================
// Transform to API format (domain â raw)
// ============================================================================

/**
 * Transform client settings to API format
 */
export function clientSettingsToRaw(settings: Partial<ClientSettings>): Partial<RawClientSettings> {
  const raw: Partial<RawClientSettings> = {};
  
  if (settings.serverUrl !== undefined) raw.server_url = settings.serverUrl;
  if (settings.deviceId !== undefined) raw.device_id = settings.deviceId;
  if (settings.requestTimeoutMs !== undefined) raw.request_timeout_ms = settings.requestTimeoutMs;
  if (settings.autoReconnect !== undefined) raw.auto_reconnect = settings.autoReconnect;
  if (settings.strictHmac !== undefined) raw.strict_hmac = settings.strictHmac;
  if (settings.logBufferLimit !== undefined) raw.log_buffer_limit = settings.logBufferLimit;
  if (settings.signalHistoryLimit !== undefined) raw.signal_history_limit = settings.signalHistoryLimit;
  
  return raw;
}

/**
 * Transform thresholds to API format
 */
export function thresholdsToRaw(thresholds: Partial<ThresholdSettings>): Partial<RawThresholds> {
  const raw: Partial<RawThresholds> = {};
  
  if (thresholds.riskWarn !== undefined) raw.risk_warn = thresholds.riskWarn;
  if (thresholds.riskCrit !== undefined) raw.risk_crit = thresholds.riskCrit;
  if (thresholds.thermalWarn !== undefined) raw.thermal_warn = thresholds.thermalWarn;
  if (thresholds.thermalCrit !== undefined) raw.thermal_crit = thresholds.thermalCrit;
  if (thresholds.bufferWarn !== undefined) raw.buffer_warn = thresholds.bufferWarn;
  if (thresholds.bufferCrit !== undefined) raw.buffer_crit = thresholds.bufferCrit;
  
  return raw;
}

/**
 * Transform notification types to API format
 */
export function notificationTypesToRaw(types: Partial<NotificationTypes>): Partial<RawNotificationTypes> {
  const raw: Partial<RawNotificationTypes> = {};
  
  if (types.thresholdBreach !== undefined) raw.threshold_breach = types.thresholdBreach;
  if (types.deviceOffline !== undefined) raw.device_offline = types.deviceOffline;
  if (types.deviceOnline !== undefined) raw.device_online = types.deviceOnline;
  if (types.updateAvailable !== undefined) raw.update_available = types.updateAvailable;
  if (types.commandFailed !== undefined) raw.command_failed = types.commandFailed;
  if (types.registrationRequest !== undefined) raw.registration_request = types.registrationRequest;
  
  return raw;
}

/**
 * Transform webhook settings to API format
 */
export function webhookSettingsToRaw(webhook: Partial<WebhookSettings>): Partial<RawWebhookSettings> {
  const raw: Partial<RawWebhookSettings> = {};
  
  if (webhook.enabled !== undefined) raw.enabled = webhook.enabled;
  if (webhook.url !== undefined) raw.url = webhook.url;
  if (webhook.types !== undefined) raw.types = webhook.types;
  
  return raw;
}

/**
 * Transform notification settings to API format
 */
export function notificationSettingsToRaw(settings: Partial<NotificationSettings>): Partial<RawNotificationSettings> {
  const raw: Partial<RawNotificationSettings> = {};
  
  if (settings.enabled !== undefined) raw.enabled = settings.enabled;
  if (settings.email !== undefined) raw.email = notificationTypesToRaw(settings.email);
  if (settings.push !== undefined) raw.push = notificationTypesToRaw(settings.push);
  if (settings.webhook !== undefined) raw.webhook = webhookSettingsToRaw(settings.webhook);
  
  return raw;
}

// ============================================================================
// Fetch Operations
// ============================================================================

/**
 * Fetch complete settings
 * GET /v1/auth/me/settings
 */
export async function fetchSettings(): Promise<Settings> {
  const data = await restClient.get<RawSettings>(SETTINGS_PATHS.settings);
  return settingsFromRaw(data);
}

/**
 * Fetch thresholds only
 * GET /v1/auth/me/thresholds
 */
export async function fetchThresholds(): Promise<ThresholdSettings> {
  const data = await restClient.get<RawThresholds>(SETTINGS_PATHS.thresholds);
  return thresholdsFromRaw(data);
}

/**
 * Fetch notification settings only
 * GET /v1/auth/me/notifications
 */
export async function fetchNotifications(): Promise<NotificationSettings> {
  const data = await restClient.get<RawNotificationSettings>(SETTINGS_PATHS.notifications);
  return notificationSettingsFromRaw(data);
}

/**
 * Fetch operator profile
 * GET /v1/auth/me
 */
export async function fetchOperator(): Promise<Operator | null> {
  const data = await restClient.get<RawOperator>(SETTINGS_PATHS.operator);
  return operatorFromRaw(data);
}

// ============================================================================
// Update Operations
// ============================================================================

/**
 * Update settings (client + thresholds)
 * PATCH /v1/auth/me/settings
 */
export async function updateSettings(settings: {
  client?: Partial<ClientSettings>;
}): Promise<Settings> {
  const payload: Record<string, unknown> = {};
  
  if (settings.client) {
    payload.client = clientSettingsToRaw(settings.client);
  }
  
  const data = await restClient.patch<RawSettings>(SETTINGS_PATHS.settings, payload);
  return settingsFromRaw(data);
}

/**
 * Update thresholds
 * PATCH /v1/auth/me/thresholds
 */
export async function updateThresholds(thresholds: Partial<ThresholdSettings>): Promise<ThresholdSettings> {
  const payload = thresholdsToRaw(thresholds);
  const data = await restClient.patch<RawThresholds>(SETTINGS_PATHS.thresholds, payload);
  return thresholdsFromRaw(data);
}

/**
 * Update notification settings
 * PATCH /v1/auth/me/notifications
 */
export async function updateNotifications(settings: Partial<NotificationSettings>): Promise<NotificationSettings> {
  const payload = notificationSettingsToRaw(settings);
  const data = await restClient.patch<RawNotificationSettings>(SETTINGS_PATHS.notifications, payload);
  return notificationSettingsFromRaw(data);
}

/**
 * Update operator profile
 * PATCH /v1/auth/me
 */
export async function updateOperator(data: {
  name?: string;
  email?: string;
}): Promise<Operator | null> {
  return restClient.patch<RawOperator>(SETTINGS_PATHS.operator, data);
}

// ============================================================================
// Webhook Operations
// ============================================================================

/**
 * Webhook test result
 */
export interface WebhookTestResult {
  success: boolean;
  statusCode?: number;
  responseTime?: number;
  error?: string;
}

/**
 * Test webhook URL
 * POST /v1/auth/me/notifications/webhook/test
 */
export async function testWebhook(url: string): Promise<WebhookTestResult> {
  return restClient.post<WebhookTestResult>(SETTINGS_PATHS.webhookTest, { url });
}

/**
 * Rotate webhook secret
 * POST /v1/auth/me/notifications/webhook/rotate
 */
export async function rotateWebhookSecret(): Promise<{ secret: string }> {
  return restClient.post<{ secret: string }>(SETTINGS_PATHS.webhookRotate);
}
