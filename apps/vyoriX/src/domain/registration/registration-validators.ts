/**
 * Registration Validators
 * 
 * Validation functions for registration-related input.
 */

import type { RegistrationRequest, AcknowledgeAction } from "./registration-entity";
import { INBOX_STATUSES } from "./registration-entity";

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
// IMEI Validation
// ============================================================================

/**
 * Validate IMEI format (15 digits)
 */
export function validateIMEI(imei: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!imei) {
    errors.imei = ["IMEI is required"];
  } else if (!/^\d{15}$/.test(imei)) {
    errors.imei = ["IMEI must be 15 digits"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Device Name Validation
// ============================================================================

/**
 * Validate device name
 */
export function validateDeviceName(name: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!name) {
    errors.deviceName = ["Device name is required"];
  } else if (name.length < 1 || name.length > 100) {
    errors.deviceName = ["Device name must be 1-100 characters"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Version Validation
// ============================================================================

/**
 * Validate version format (semver)
 */
export function validateVersion(version: string, fieldName: string = "version"): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!version) {
    errors[fieldName] = [`${fieldName} is required`];
  } else if (!/^\d+\.\d+\.\d+$/.test(version)) {
    errors[fieldName] = [`${fieldName} must be in semver format (e.g., 1.2.3)`];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Registration Request Validation
// ============================================================================

/**
 * Validate complete registration request
 */
export function validateRegistrationRequest(request: Partial<RegistrationRequest>): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  // IMEI is required
  const imeiResult = validateIMEI(request.imei ?? "");
  Object.assign(errors, imeiResult.errors);
  
  // Device name is required
  const nameResult = validateDeviceName(request.deviceName ?? "");
  Object.assign(errors, nameResult.errors);
  
  // OS version is required
  if (!request.osVersion) {
    errors.osVersion = ["OS version is required"];
  }
  
  // App version is required
  const versionResult = validateVersion(request.appVersion ?? "", "appVersion");
  Object.assign(errors, versionResult.errors);
  
  // FCM token is required
  if (!request.fcmToken) {
    errors.fcmToken = ["FCM token is required"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Acknowledge Action Validation
// ============================================================================

/**
 * Valid acknowledge actions
 */
const VALID_ACKNOWLEDGE_ACTIONS: AcknowledgeAction[] = ["acknowledge", "approve", "reject"];

/**
 * Validate acknowledge action
 */
export function validateAcknowledgeAction(action: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!action) {
    errors.action = ["Action is required"];
  } else if (!VALID_ACKNOWLEDGE_ACTIONS.includes(action as AcknowledgeAction)) {
    errors.action = [`Invalid action: ${action}. Valid actions: ${VALID_ACKNOWLEDGE_ACTIONS.join(", ")}`];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

// ============================================================================
// Status Transition Validation
// ============================================================================

/**
 * Valid status transitions
 */
const VALID_TRANSITIONS: Record<string, InboxStatus[]> = {
  pending: ["acknowledged", "rejected"],
  acknowledged: ["approving", "rejected"], // approving is intermediate, approved is terminal
  approving: ["approved", "rejected"],
  approved: [], // terminal
  rejected: [], // terminal
  expired: [], // terminal
};

/**
 * Check if status transition is valid
 */
export function isValidStatusTransition(
  currentStatus: string,
  newStatus: string
): boolean {
  const validTargets = VALID_TRANSITIONS[currentStatus] ?? [];
  return validTargets.includes(newStatus as InboxStatus);
}

/**
 * Validate status transition
 */
export function validateStatusTransition(
  currentStatus: InboxStatus,
  newStatus: InboxStatus
): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!isValidStatusTransition(currentStatus, newStatus)) {
    errors.status = [
      `Cannot transition from ${currentStatus} to ${newStatus}`,
    ];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}