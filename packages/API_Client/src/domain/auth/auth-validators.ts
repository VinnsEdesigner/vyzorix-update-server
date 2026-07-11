export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
}

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function validateEmail(email: string): ValidationResult {
  if (!email) {
    return { isValid: false, errors: [{ field: "email", message: "Email is required" }] };
  }
  if (!EMAIL_REGEX.test(email)) {
    return { isValid: false, errors: [{ field: "email", message: "Invalid email format" }] };
  }
  return { isValid: true, errors: [] };
}

export function validatePassword(password: string): ValidationResult {
  const errors: ValidationError[] = [];
  if (!password) {
    return { isValid: false, errors: [{ field: "password", message: "Password is required" }] };
  }
  if (password.length < 8) {
    errors.push({ field: "password", message: "Password must be at least 8 characters" });
  }
  return { isValid: errors.length === 0, errors };
}

export function validateName(name: string): ValidationResult {
  if (!name || name.trim().length < 2) {
    return { isValid: false, errors: [{ field: "name", message: "Name must be at least 2 characters" }] };
  }
  return { isValid: true, errors: [] };
}

export function validateMFACode(code: string): ValidationResult {
  if (!code || !/^\d{6}$/.test(code)) {
    return { isValid: false, errors: [{ field: "code", message: "MFA code must be 6 digits" }] };
  }
  return { isValid: true, errors: [] };
}

export function validateLogin(credentials: { email: string; password: string }): ValidationResult {
  const emailResult = validateEmail(credentials.email);
  const passwordResult = validatePassword(credentials.password);
  return {
    isValid: emailResult.isValid && passwordResult.isValid,
    errors: [...emailResult.errors, ...passwordResult.errors],
  };
}

export function validateRegister(credentials: { email: string; password: string; name: string }): ValidationResult {
  const emailResult = validateEmail(credentials.email);
  const passwordResult = validatePassword(credentials.password);
  const nameResult = validateName(credentials.name);
  return {
    isValid: emailResult.isValid && passwordResult.isValid && nameResult.isValid,
    errors: [...emailResult.errors, ...passwordResult.errors, ...nameResult.errors],
  };
}
