/**
 * Settings Domain Index
 * 
 * Re-exports all settings domain types, mappers, and validators.
 */

// Types
export type {
  ConnectionSettings,
  ThresholdSettings,
  NotificationSettings,
  NotificationChannel,
  NotificationEvent,
  OperatorInfo,
  AdvancedSettings,
  Settings,
} from "./settings-entity";

export {
  DEFAULT_THRESHOLDS,
  DEFAULT_CONNECTION,
  DEFAULT_ADVANCED,
  createDefaultSettings,
  isRiskWarning,
  isRiskCritical,
  isThermalWarning,
  isThermalCritical,
  isBufferWarning,
  isBufferCritical,
} from "./settings-entity";

// Mappers
export type {
  RawConnectionSettings,
  RawThresholdSettings,
  RawNotificationChannel,
  RawNotificationSettings,
  RawOperatorInfo,
  RawAdvancedSettings,
  RawSettings,
} from "./settings-mappers";

export {
  connectionFromRaw,
  thresholdsFromRaw,
  notificationsFromRaw,
  operatorFromRaw,
  advancedFromRaw,
  settingsFromRaw,
  connectionToRaw,
  thresholdsToRaw,
  notificationChannelToRaw,
  notificationsToRaw,
  advancedToRaw,
} from "./settings-mappers";

// Validators
export type { ValidationResult } from "./settings-validators";

export {
  THRESHOLD_LIMITS,
  validateThresholdValue,
  validateThresholdSettings,
  isValidUrl,
  validateServerUrl,
  validateDeviceId,
  validateRequestTimeout,
  validateConnectionSettings,
  validateWebhookUrl,
  validateBufferLimit,
  validateSignalHistoryLimit,
} from "./settings-validators";
