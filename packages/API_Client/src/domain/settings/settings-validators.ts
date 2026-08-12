




import type { Thresholds, ClientSettings } from "./settings-entity";
import type { ValidationResult } from "../_shared";

export { type ValidationResult } from "../_shared";






export const THRESHOLD_LIMITS = {
  risk: { min: 0, max: 100 },
  thermal: { min: 0, max: 100 },
  buffer: { min: 0, max: 100 },
} as const;


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


export function validateThresholds(thresholds: Thresholds): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  
  if (thresholds.riskWarn >= thresholds.riskCrit) {
    errors.riskWarn = ["Warning threshold must be less than critical threshold"];
  }
  
  
  if (thresholds.thermalWarn >= thresholds.thermalCrit) {
    errors.thermalWarn = ["Warning threshold must be less than critical threshold"];
  }
  
  
  if (thresholds.bufferWarn <= thresholds.bufferCrit) {
    errors.bufferWarn = ["Warning threshold must be greater than critical threshold"];
  }
  
  
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






export function isValidUrl(url: string): boolean {
  if (!url) return true; 
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}


export function validateServerUrl(url: string): string | null {
  if (!url) return null; 
  
  if (!isValidUrl(url)) {
    return "Invalid server URL format";
  }
  
  return null;
}


export function validateDeviceId(id: string): string | null {
  if (!id) return null; 
  
  
  if (!/^\d{15}$/.test(id)) {
    return "Device ID must be a valid IMEI (15 digits)";
  }
  
  return null;
}


export function validateRequestTimeout(ms: number): string | null {
  if (typeof ms !== "number" || isNaN(ms)) {
    return "Request timeout must be a number";
  }
  
  if (ms < 1000 || ms > 60000) {
    return "Request timeout must be between 1 and 60 seconds";
  }
  
  return null;
}


export function validateClientSettings(settings: ClientSettings): ValidationResult {
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






export function validateWebhookUrl(url: string): string | null {
  if (!url) return null; 
  
  if (!isValidUrl(url)) {
    return "Invalid webhook URL format";
  }
  
  
  if (url.startsWith("http://")) {
    return "Webhook URL should use HTTPS for security";
  }
  
  return null;
}






export function validateBufferLimit(limit: number): string | null {
  if (typeof limit !== "number" || isNaN(limit)) {
    return "Buffer limit must be a number";
  }
  
  if (limit < 1 || limit > 1000) {
    return "Buffer limit must be between 1 and 1000";
  }
  
  return null;
}


export function validateSignalHistoryLimit(limit: number): string | null {
  if (typeof limit !== "number" || isNaN(limit)) {
    return "Signal history limit must be a number";
  }
  
  if (limit < 10 || limit > 1000) {
    return "Signal history limit must be between 10 and 1000";
  }
  
  return null;
}
