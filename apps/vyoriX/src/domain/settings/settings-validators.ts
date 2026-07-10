/**
 * Settings Validators
 * 
 * Validation functions for settings input.
 * Returns validation result with errors array.
 */

import { ValidationError } from "@/domain/_shared";
import type { ThresholdSettings, ConnectionSettings } from "./settings-entity";

// ============================================================================
// Validation Result
// ============================================================================

/**
 * Validation result
 */
export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string[]>;
}

// ============================================================================
// Threshold Validators
// ============================================================================

/**
 * Threshold value limits
 */
export const THRESHOLD_LIMITS = {
  risk: { min: 0, max: 100 },
  thermal: { min: 0, max: 100 },
  buffer: { min: 0, max: 100 },
} as const;

/**
 * Validate threshold value is within range
 */
export function validateThresholdValue(
  value: number,
  type: keyof typeof THRESHOLD_LIMITS,
  field: string
): string | null {
  const limits = THRESHOLD_LIMITS[type];
  
  if (typeof value !== "number" || isNaN(value)) {
    return `${field} must be a number`;
  }
  
  if (value < limits.min || value > limits.max) {
    return `${field} must be between ${limits.min} and ${limits.max}`;
  }
  
  return null;
}

/**
 * Validate threshold settings (warning < critical for risk/thermal, warning > critical for buffer)
 */
export function validateThresholdSettings(thresholds: ThresholdSettings): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  // Risk thresholds: warning should be < critical
  if (thresholds.riskWarn >= thresholds.riskCrit) {
    errors.riskWarn = ["Warning threshold must be less than critical threshold"];
  }
  
  // Thermal thresholds: warning should be < critical
  if (thresholds.thermalWarn >= thresholds.thermalCrit) {
    errors.thermalWarn = ["Warning threshold must be less than critical threshold"];
  }
  
  // Buffer thresholds (inverted): warning should be > critical
  if (thresholds.bufferWarn <= thresholds.bufferCrit) {
    errors.bufferWarn = ["Warning threshold must be greater than critical threshold"];
  }
  
  // Validate individual values
  const riskWarnError = validateThresholdValue(thresholds.riskWarn, "risk", "Risk warning");
  if (riskWarnError) errors.riskWarn = [riskWarnError];
  
  const riskCritError = validateThresholdValue(thresholds.riskCrit, "risk", "Risk critical");
  if (riskCritError) errors.riskCrit = [riskCritError];
  
  const thermalWarnError = validateThresholdValue(thresholds.thermalWarn, "thermal", "Thermal warning");
  if (thermalWarnError) errors.thermalWarn = [thermalWarnError];
  
  const thermalCritError = validateThresholdValue(thresholds.thermalCrit, "thermal", "Thermal critical");
  if (thermalCritError) errors.thermalCrit = [thermalCritError];
  
  const bufferWarnError = validateThresholdValue(thresholds.bufferWarn, "buffer", "Buffer warning");
  if (bufferWarnError) errors.bufferWarn = [bufferWarnError];
  
  const bufferCritError = validateThresholdValue(thresholds.bufferCrit, "buffer", "Buffer critical");
  if (bufferCritError) errors.bufferCrit = [bufferCritError];
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Connection Validators
// ============================================================================

/**
 * Validate URL format
 */
export function isValidUrl(url: string): boolean {
  if (!url) return true; // Empty is valid (optional)
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Validate server URL
 */
export function validateServerUrl(url: string): string | null {
  if (!url) return null; // Optional
  
  if (!isValidUrl(url)) {
    return "Invalid server URL format";
  }
  
  return null;
}

/**
 * Validate device ID (IMEI format)
 */
export function validateDeviceId(id: string): string | null {
  if (!id) return null; // Optional
  
  // IMEI is 15 digits
  if (!/^\d{15}$/.test(id)) {
    return "Device ID must be a valid IMEI (15 digits)";
  }
  
  return null;
}

/**
 * Validate request timeout (1-60 seconds)
 */
export function validateRequestTimeout(ms: number): string | null {
  if (typeof ms !== "number" || isNaN(ms)) {
    return "Request timeout must be a number";
  }
  
  if (ms < 1000 || ms > 60000) {
    return "Request timeout must be between 1 and 60 seconds";
  }
  
  return null;
}

/**
 * Validate connection settings
 */
export function validateConnectionSettings(settings: ConnectionSettings): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  const serverUrlError = validateServerUrl(settings.serverUrl);
  if (serverUrlError) errors.serverUrl = [serverUrlError];
  
  const deviceIdError = validateDeviceId(settings.deviceId);
  if (deviceIdError) errors.deviceId = [deviceIdError];
  
  const timeoutError = validateRequestTimeout(settings.requestTimeoutMs);
  if (timeoutError) errors.requestTimeoutMs = [timeoutError];
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Webhook Validators
// ============================================================================

/**
 * Validate webhook URL
 */
export function validateWebhookUrl(url: string): string | null {
  if (!url) return null; // Optional
  
  if (!isValidUrl(url)) {
    return "Invalid webhook URL format";
  }
  
  // Should be HTTPS in production
  if (url.startsWith("http://")) {
    return "Webhook URL should use HTTPS for security";
  }
  
  return null;
}

// ============================================================================
// Advanced Settings Validators
// ============================================================================

/**
 * Validate buffer limit (1-1000)
 */
export function validateBufferLimit(limit: number): string | null {
  if (typeof limit !== "number" || isNaN(limit)) {
    return "Buffer limit must be a number";
  }
  
  if (limit < 1 || limit > 1000) {
    return "Buffer limit must be between 1 and 1000";
  }
  
  return null;
}

/**
 * Validate signal history limit (10-1000)
 */
export function validateSignalHistoryLimit(limit: number): string | null {
  if (typeof limit !== "number" || isNaN(limit)) {
    return "Signal history limit must be a number";
  }
  
  if (limit < 10 || limit > 1000) {
    return "Signal history limit must be between 10 and 1000";
  }
  
  return null;
}
