import type { ValidationResult } from "../_shared";

export { type ValidationResult } from "../_shared";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function fieldError(field: string, message: string): Record<string, string[]> {
  return { [field]: [message] };
}

export function validateEmail(email: string): ValidationResult {
  if (!email) {
    return { isValid: false, errors: fieldError("email", "Email is required") };
  }
  if (!EMAIL_REGEX.test(email)) {
    return { isValid: false, errors: fieldError("email", "Invalid email format") };
  }
  return { isValid: true, errors: {} };
}

export function validatePassword(password: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  if (!password) {
    return { isValid: false, errors: fieldError("password", "Password is required") };
  }
  if (password.length < 8) {
    errors.password = ["Password must be at least 8 characters"];
  }
  return { isValid: Object.keys(errors).length === 0, errors };
}

export function validateName(name: string): ValidationResult {
  if (!name || name.trim().length < 2) {
    return { isValid: false, errors: fieldError("name", "Name must be at least 2 characters") };
  }
  return { isValid: true, errors: {} };
}

export function validateMFACode(code: string): ValidationResult {
  if (!code || !/^\d{6}$/.test(code)) {
    return { isValid: false, errors: fieldError("code", "MFA code must be 6 digits") };
  }
  return { isValid: true, errors: {} };
}

export function validateLogin(credentials: { email: string; password: string }): ValidationResult {
  const emailResult = validateEmail(credentials.email);
  const passwordResult = validatePassword(credentials.password);
  return {
    isValid: emailResult.isValid && passwordResult.isValid,
    errors: { ...emailResult.errors, ...passwordResult.errors },
  };
}

export function validateRegister(credentials: { email: string; password: string; name: string }): ValidationResult {
  const emailResult = validateEmail(credentials.email);
  const passwordResult = validatePassword(credentials.password);
  const nameResult = validateName(credentials.name);
  return {
    isValid: emailResult.isValid && passwordResult.isValid && nameResult.isValid,
    errors: { ...emailResult.errors, ...passwordResult.errors, ...nameResult.errors },
  };
}
