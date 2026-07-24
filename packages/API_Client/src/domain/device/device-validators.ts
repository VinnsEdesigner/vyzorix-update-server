

import { ValidationError } from "@/domain/_shared";






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


export function validateIMEIChecksum(imei: string): boolean {
  if (!/^\d{15}$/.test(imei)) return false;
  
  let sum = 0;
  let isEven = false;
  
  for (let i = imei.length - 1; i >= 0; i--) {
    let digit = parseInt(imei[i], 10);
    
    if (isEven) {
      digit *= 2;
      if (digit > 9) {
        digit -= 9;
      }
    }
    
    sum += digit;
    isEven = !isEven;
  }
  
  return sum % 10 === 0;
}






export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string[]>;
}






export function validateFCMToken(token: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  
  if (!token) {
    errors.fcmToken = ["FCM token is required"];
  } else if (token.length < 100 || token.length > 4000) {
    errors.fcmToken = ["FCM token has invalid length"];
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}
