

import type { PresetCommandType, CommandParams } from "./commands-entity";






export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string[]>;
}






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






const VALID_LOG_LEVELS = ["debug", "info", "warn", "error", "verbose", "assert"];


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
      
      return { isValid: true, errors: {} };
  }
}






export function validateSendCommand(imei: string, commandType: string, params?: CommandParams): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  
  const imeiResult = validateIMEI(imei);
  Object.assign(errors, imeiResult.errors);
  
  
  const typeResult = validatePresetCommandType(commandType);
  Object.assign(errors, typeResult.errors);
  
  
  if (typeResult.isValid) {
    const paramsResult = validatePresetCommandParams(commandType as PresetCommandType, params ?? {});
    Object.assign(errors, paramsResult.errors);
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}
