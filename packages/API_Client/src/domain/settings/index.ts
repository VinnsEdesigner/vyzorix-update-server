// Settings domain — generated types + zod structural validation + hand-rolled
// business rules (.refine). Entity types and Raw mappers are eliminated; the
// generated schemas ARE the wire types.
import {
  OperatorThresholdsSchema,
  ClientSettingsSchema,
  WebhookNotificationsSchema,
} from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  OperatorThresholds,
  ClientSettings,
  NotificationSettings,
  WebhookNotifications,
  ThresholdUpdateRequest,
  NotificationUpdateRequest,
  UpdateSettingsRequest,
  OperatorSettingsResult,
  SettingsResponseResult,
  ThresholdsResult,
  WebhookTestResult,
  WebhookSecretResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

// Re-export generated types so domain consumers don't import from generated directly.
export type {
  OperatorThresholds,
  OperatorThresholds as Thresholds,
  ClientSettings,
  NotificationSettings,
  WebhookNotifications,
  ThresholdUpdateRequest,
  NotificationUpdateRequest,
  UpdateSettingsRequest,
  OperatorSettingsResult,
  SettingsResponseResult,
  ThresholdsResult,
  WebhookTestResult,
  WebhookSecretResult,
};

// ---- Constants (hand-rolled, not in OpenAPI) ----

export const THRESHOLD_LIMITS = {
  risk: { min: 0, max: 100 },
  thermal: { min: 0, max: 100 },
  buffer: { min: 0, max: 100 },
} as const;

export const CLIENT_SETTINGS_LIMITS = {
  requestTimeoutMs: { min: 1000, max: 60000 },
  logBufferLimit: { min: 1, max: 1000 },
  signalHistoryLimit: { min: 10, max: 1000 },
} as const;

export type NotificationEvent =
  | 'threshold_breach'
  | 'device_offline'
  | 'device_online'
  | 'update_available'
  | 'command_failed'
  | 'registration_request';

// ---- Threshold validators (business rules on top of generated zod) ----

export const thresholdsValidator = OperatorThresholdsSchema
  // Range checks: all threshold values must be 0-100
  .refine(
    (t) => t.riskWarn === undefined || (t.riskWarn >= 0 && t.riskWarn <= 100),
    { message: 'Risk warning must be between 0 and 100', path: ['riskWarn'] },
  )
  .refine(
    (t) => t.riskCrit === undefined || (t.riskCrit >= 0 && t.riskCrit <= 100),
    { message: 'Risk critical must be between 0 and 100', path: ['riskCrit'] },
  )
  .refine(
    (t) => t.thermalWarn === undefined || (t.thermalWarn >= 0 && t.thermalWarn <= 100),
    { message: 'Thermal warning must be between 0 and 100', path: ['thermalWarn'] },
  )
  .refine(
    (t) => t.thermalCrit === undefined || (t.thermalCrit >= 0 && t.thermalCrit <= 100),
    { message: 'Thermal critical must be between 0 and 100', path: ['thermalCrit'] },
  )
  .refine(
    (t) => t.bufferWarn === undefined || (t.bufferWarn >= 0 && t.bufferWarn <= 100),
    { message: 'Buffer warning must be between 0 and 100', path: ['bufferWarn'] },
  )
  .refine(
    (t) => t.bufferCrit === undefined || (t.bufferCrit >= 0 && t.bufferCrit <= 100),
    { message: 'Buffer critical must be between 0 and 100', path: ['bufferCrit'] },
  )
  // Cross-field: warn must be less than crit for each metric
  .refine(
    (t) => t.riskWarn === undefined || t.riskCrit === undefined || t.riskWarn < t.riskCrit,
    { message: 'Warning threshold must be less than critical threshold', path: ['riskWarn'] },
  )
  .refine(
    (t) => t.thermalWarn === undefined || t.thermalCrit === undefined || t.thermalWarn < t.thermalCrit,
    { message: 'Warning threshold must be less than critical threshold', path: ['thermalWarn'] },
  )
  .refine(
    (t) => t.bufferWarn === undefined || t.bufferCrit === undefined || t.bufferWarn < t.bufferCrit,
    { message: 'Warning threshold must be less than critical threshold', path: ['bufferWarn'] },
  );

export function validateThresholds(input: unknown) {
  return thresholdsValidator.safeParse(input);
}

// ---- Client settings validators ----

export const clientSettingsValidator = ClientSettingsSchema
  .refine(
    (s) => !s.serverUrl || isValidUrl(s.serverUrl),
    { message: 'Invalid server URL format', path: ['serverUrl'] },
  )
  .refine(
    (s) => !s.deviceId || /^\d{15}$/.test(s.deviceId),
    { message: 'Device ID must be a valid IMEI (15 digits)', path: ['deviceId'] },
  )
  .refine(
    (s) => s.requestTimeoutMs !== undefined && s.requestTimeoutMs >= CLIENT_SETTINGS_LIMITS.requestTimeoutMs.min && s.requestTimeoutMs <= CLIENT_SETTINGS_LIMITS.requestTimeoutMs.max,
    { message: `Request timeout must be between ${CLIENT_SETTINGS_LIMITS.requestTimeoutMs.min} and ${CLIENT_SETTINGS_LIMITS.requestTimeoutMs.max}ms`, path: ['requestTimeoutMs'] },
  )
  .refine(
    (s) => s.logBufferLimit !== undefined && s.logBufferLimit >= CLIENT_SETTINGS_LIMITS.logBufferLimit.min && s.logBufferLimit <= CLIENT_SETTINGS_LIMITS.logBufferLimit.max,
    { message: `Buffer limit must be between ${CLIENT_SETTINGS_LIMITS.logBufferLimit.min} and ${CLIENT_SETTINGS_LIMITS.logBufferLimit.max}`, path: ['logBufferLimit'] },
  )
  .refine(
    (s) => s.signalHistoryLimit !== undefined && s.signalHistoryLimit >= CLIENT_SETTINGS_LIMITS.signalHistoryLimit.min && s.signalHistoryLimit <= CLIENT_SETTINGS_LIMITS.signalHistoryLimit.max,
    { message: `Signal history limit must be between ${CLIENT_SETTINGS_LIMITS.signalHistoryLimit.min} and ${CLIENT_SETTINGS_LIMITS.signalHistoryLimit.max}`, path: ['signalHistoryLimit'] },
  );

export function validateClientSettings(input: unknown) {
  return clientSettingsValidator.safeParse(input);
}

// ---- Webhook validators ----

export const webhookValidator = WebhookNotificationsSchema
  .refine(
    (w) => !w.url || isValidUrl(w.url),
    { message: 'Invalid webhook URL format', path: ['url'] },
  )
  .refine(
    (w) => !w.url || !w.url.startsWith('http://'),
    { message: 'Webhook URL should use HTTPS for security', path: ['url'] },
  );

export function validateWebhook(input: unknown) {
  return webhookValidator.safeParse(input);
}

// ---- Helpers ----

function isValidUrl(url: string): boolean {
  if (!url) return true;
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

export function validateThresholdValue(
  value: number,
  type: keyof typeof THRESHOLD_LIMITS,
  field: string,
): string | null {
  const limits = THRESHOLD_LIMITS[type];
  if (typeof value !== 'number' || isNaN(value)) return `${field} must be a number`;
  if (value < limits.min || value > limits.max) return `${field} must be between ${limits.min} and ${limits.max}`;
  return null;
}
