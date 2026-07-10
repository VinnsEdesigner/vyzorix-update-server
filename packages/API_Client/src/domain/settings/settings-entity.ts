/**
 * Settings Domain Types
 * 
 * Domain types for operator settings configuration.
 * Pure TypeScript - no external imports.
 */

// ============================================================================
// Connection Settings
// ============================================================================

/**
 * Connection configuration for device communication
 */
export interface ConnectionSettings {
  serverUrl: string;
  deviceId: string;
  dashboardToken: string;
  requestTimeoutMs: number;
  autoReconnect: boolean;
  strictHmac: boolean;
}

// ============================================================================
// Threshold Settings
// ============================================================================

/**
 * Threshold levels for alerts
 */
export interface ThresholdSettings {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

// ============================================================================
// Notification Settings
// ============================================================================

/**
 * Notification channel configuration
 */
export interface NotificationChannel {
  enabled: boolean;
  email?: string;
  webhookUrl?: string;
  webhookSecret?: string;
}

/**
 * Notification event types
 */
export type NotificationEvent =
  | "threshold_breach"
  | "device_offline"
  | "device_online"
  | "update_available"
  | "command_failed"
  | "registration_request";

/**
 * Notification settings
 */
export interface NotificationSettings {
  enabled: boolean;
  email?: NotificationChannel;
  push?: NotificationChannel;
  webhook?: NotificationChannel;
  events: Partial<Record<NotificationEvent, boolean>>;
}

// ============================================================================
// Operator Settings
// ============================================================================

/**
 * Operator account information
 */
export interface OperatorInfo {
  id: string;
  email: string;
  name: string;
  role: "operator" | "admin";
  emailVerified: boolean;
  createdAt: Date;
}

// ============================================================================
// Advanced Settings
// ============================================================================

/**
 * Advanced client settings
 */
export interface AdvancedSettings {
  logBufferLimit: number;
  signalHistoryLimit: number;
}

// ============================================================================
// Complete Settings
// ============================================================================

/**
 * Complete operator settings
 */
export interface Settings {
  operator: OperatorInfo;
  connection: ConnectionSettings;
  thresholds: ThresholdSettings;
  notifications: NotificationSettings;
  advanced: AdvancedSettings;
}

// ============================================================================
// Default Values
// ============================================================================

/**
 * Default threshold values
 */
export const DEFAULT_THRESHOLDS: ThresholdSettings = {
  riskWarn: 70,
  riskCrit: 85,
  thermalWarn: 45,
  thermalCrit: 50,
  bufferWarn: 30,
  bufferCrit: 15,
};

/**
 * Default connection settings
 */
export const DEFAULT_CONNECTION: ConnectionSettings = {
  serverUrl: "",
  deviceId: "",
  dashboardToken: "",
  requestTimeoutMs: 8000,
  autoReconnect: true,
  strictHmac: false,
};

/**
 * Default advanced settings
 */
export const DEFAULT_ADVANCED: AdvancedSettings = {
  logBufferLimit: 500,
  signalHistoryLimit: 240,
};

/**
 * Create default settings
 */
export function createDefaultSettings(operator: OperatorInfo): Settings {
  return {
    operator,
    connection: { ...DEFAULT_CONNECTION },
    thresholds: { ...DEFAULT_THRESHOLDS },
    notifications: { enabled: true, events: {} },
    advanced: { ...DEFAULT_ADVANCED },
  };
}

// ============================================================================
// Threshold Helpers
// ============================================================================

/**
 * Check if risk level is in warning zone
 */
export function isRiskWarning(threshold: ThresholdSettings, value: number): boolean {
  return value >= threshold.riskWarn && value < threshold.riskCrit;
}

/**
 * Check if risk level is in critical zone
 */
export function isRiskCritical(threshold: ThresholdSettings, value: number): boolean {
  return value >= threshold.riskCrit;
}

/**
 * Check if thermal is in warning zone
 */
export function isThermalWarning(threshold: ThresholdSettings, value: number): boolean {
  return value >= threshold.thermalWarn && value < threshold.thermalCrit;
}

/**
 * Check if thermal is in critical zone
 */
export function isThermalCritical(threshold: ThresholdSettings, value: number): boolean {
  return value >= threshold.thermalCrit;
}

/**
 * Check if buffer is in warning zone (inverted - low is bad)
 */
export function isBufferWarning(threshold: ThresholdSettings, value: number): boolean {
  return value <= threshold.bufferWarn && value > threshold.bufferCrit;
}

/**
 * Check if buffer is in critical zone (inverted - low is bad)
 */
export function isBufferCritical(threshold: ThresholdSettings, value: number): boolean {
  return value <= threshold.bufferCrit;
}
