import type { ApiKeyScope } from "./apikey-entity";
import { MAX_NAME_LENGTH, MIN_NAME_LENGTH, MAX_EXPIRY_DAYS } from "./apikey-constants";

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

export function validateApiKeyName(name: string): ValidationResult {
  const errors: string[] = [];

  if (!name || name.trim().length === 0) {
    errors.push("Name is required");
  }

  if (name.length > MAX_NAME_LENGTH) {
    errors.push(`Name must be ${MAX_NAME_LENGTH} characters or less`);
  }

  if (name.length < MIN_NAME_LENGTH) {
    errors.push(`Name must be at least ${MIN_NAME_LENGTH} character`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

export function validateExpiryDays(days: number | null): ValidationResult {
  const errors: string[] = [];

  if (days === null || days === undefined) {
    return { isValid: true, errors: [] };
  }

  if (typeof days !== "number" || days < 1) {
    errors.push("Expiry days must be a positive number");
  }

  if (days > MAX_EXPIRY_DAYS) {
    errors.push(`Expiry days cannot exceed ${MAX_EXPIRY_DAYS}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

export function validateScope(scope: string): ValidationResult {
  const validScopes: ApiKeyScope[] = ["read", "write", "admin"];
  const errors: string[] = [];

  if (!validScopes.includes(scope as ApiKeyScope)) {
    errors.push(`Scope must be one of: ${validScopes.join(", ")}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

export function validateCreateApiKeyRequest(
  name: string,
  scope: string,
  expiresInDays?: number | null
): ValidationResult {
  const errors: string[] = [];

  const nameResult = validateApiKeyName(name);
  errors.push(...nameResult.errors);

  const scopeResult = validateScope(scope);
  errors.push(...scopeResult.errors);

  if (expiresInDays !== undefined && expiresInDays !== null) {
    const expiryResult = validateExpiryDays(expiresInDays);
    errors.push(...expiryResult.errors);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}
