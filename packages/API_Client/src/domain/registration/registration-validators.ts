




import type { CreateInboxRequest, AcknowledgeAction, InboxStatus } from "./registration-entity";










export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string[]>;
}






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






export function validateRegistrationRequest(request: Partial<CreateInboxRequest>): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  
  const imeiResult = validateIMEI(request.imei ?? "");
  Object.assign(errors, imeiResult.errors);
  
  
  const nameResult = validateDeviceName(request.deviceName ?? "");
  Object.assign(errors, nameResult.errors);
  
  
  if (!request.osVersion) {
    errors.osVersion = ["OS version is required"];
  }
  
  
  const versionResult = validateVersion(request.appVersion ?? "", "appVersion");
  Object.assign(errors, versionResult.errors);
  
  
  if (!request.fcmToken) {
    errors.fcmToken = ["FCM token is required"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}






const VALID_ACKNOWLEDGE_ACTIONS: AcknowledgeAction[] = ["acknowledge", "approve", "reject"];


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






const VALID_TRANSITIONS: Record<string, InboxStatus[]> = {
  pending: ["acknowledged", "rejected"],
  acknowledged: ["approving", "rejected"], 
  approving: ["approved", "rejected"],
  approved: [], 
  rejected: [], 
  expired: [], 
};


export function isValidStatusTransition(
  currentStatus: string,
  newStatus: string
): boolean {
  const validTargets = VALID_TRANSITIONS[currentStatus] ?? [];
  return validTargets.includes(newStatus as InboxStatus);
}


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