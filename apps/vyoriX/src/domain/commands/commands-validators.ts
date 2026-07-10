/**
 * Commands Validators
 * 
 * Validation functions for command-related input.
 * Commands are HMAC-SHA256 signed for security.
 */

import type { PresetCommandType, CommandParams } from "./commands-entity";

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
// Preset Command Type Validation
// ============================================================================

/**
 * Valid preset command types
 */
const VALID_PRESET_TYPES: PresetCommandType[] = [
  "FORCE_SPEAKER",
  "RESET_AUDIO_HAL",
  "TOGGLE_CAPTURE",
  "REINIT_PROJECTION",
  "DUMP_FLIGHT_DATA",
  "UPLOAD_CRASH_ZIP",
  "SET_LOG_LEVEL",
  "WAKE_UP_UPDATER",
];

/**
 * Validate preset command type
 */
export function validatePresetCommandType(type: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!type) {
    errors.type = ["Command type is required"];
  } else if (!VALID_PRESET_TYPES.includes(type as PresetCommandType)) {
    errors.type = [`Invalid preset command type: ${type}`];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// IMEI Validation
// ============================================================================

/**
 * Validate IMEI format (15 digits)
 */
export function validateIMEI(imei: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!imei) {
    errors.imei = ["Device IMEI is required"];
  } else if (!/^\d{15}$/.test(imei)) {
    errors.imei = ["Device IMEI must be 15 digits"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Log Level Validation
// ============================================================================

/**
 * Valid log levels
 */
const VALID_LOG_LEVELS = ["debug", "info", "warn", "error", "verbose", "assert"];

/**
 * Validate log level parameter for SET_LOG_LEVEL command
 */
export function validateLogLevel(level: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!level) {
    errors.level = ["Log level is required for SET_LOG_LEVEL"];
  } else if (!VALID_LOG_LEVELS.includes(level.toLowerCase())) {
    errors.level = [`Invalid log level: ${level}. Valid levels: ${VALID_LOG_LEVELS.join(", ")}`];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Command Params Validation
// ============================================================================

/**
 * Validate command params for TOGGLE_CAPTURE
 */
export function validateToggleCaptureParams(params: CommandParams): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (typeof params.active !== "boolean") {
    errors.active = ["'active' parameter must be a boolean for TOGGLE_CAPTURE"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

/**
 * Validate command params based on preset type
 */
export function validatePresetCommandParams(type: PresetCommandType, params: CommandParams): ValidationResult {
  switch (type) {
    case "TOGGLE_CAPTURE":
      return validateToggleCaptureParams(params);
    case "SET_LOG_LEVEL":
      if (params.level) {
        return validateLogLevel(params.level);
      }
      return { isValid: false, errors: { level: ["Log level is required for SET_LOG_LEVEL"] } };
    default:
      // Other preset commands don't require specific params
      return { isValid: true, errors: {} };
  }
}

// ============================================================================
// Full Command Validation
// ============================================================================

/**
 * Validate complete command request
 */
export function validateSendCommand(imei: string, commandType: string, params?: CommandParams): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  // Validate IMEI
  const imeiResult = validateIMEI(imei);
  Object.assign(errors, imeiResult.errors);
  
  // Validate preset command type
  const typeResult = validatePresetCommandType(commandType);
  Object.assign(errors, typeResult.errors);
  
  // Validate params if type is valid
  if (typeResult.isValid) {
    const paramsResult = validatePresetCommandParams(commandType as PresetCommandType, params ?? {});
    Object.assign(errors, paramsResult.errors);
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}
