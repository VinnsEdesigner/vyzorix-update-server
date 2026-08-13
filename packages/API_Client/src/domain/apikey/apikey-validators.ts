import type { ApiKeyScope } from "./apikey-entity";
import { MAX_NAME_LENGTH, MIN_NAME_LENGTH, MAX_EXPIRY_DAYS } from "./apikey-constants";
import type { ValidationResult } from "../_shared";

export { type ValidationResult } from "../_shared";

export function validateApiKeyName(name: string): ValidationResult {
  const errors: Record<string, string[]> = {};

  if (!name || name.trim().length === 0) {
    errors.name = ["Name is required"];
  }

  if (name.length > MAX_NAME_LENGTH) {
    errors.name = [...(errors.name ?? []), `Name must be ${MAX_NAME_LENGTH} characters or less`];
  }

  if (name.length < MIN_NAME_LENGTH) {
    errors.name = [...(errors.name ?? []), `Name must be at least ${MIN_NAME_LENGTH} character`];
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

export function validateExpiryDays(days: number | null): ValidationResult {
  const errors: Record<string, string[]> = {};

  if (days === null || days === undefined) {
    return { isValid: true, errors: {} };
  }

  if (typeof days !== "number" || days < 1) {
    errors.expiryDays = ["Expiry days must be a positive number"];
  }

  if (days > MAX_EXPIRY_DAYS) {
    errors.expiryDays = [...(errors.expiryDays ?? []), `Expiry days cannot exceed ${MAX_EXPIRY_DAYS}`];
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

export function validateScope(scope: string): ValidationResult {
  const validScopes: ApiKeyScope[] = ["read", "write", "admin"];
  const errors: Record<string, string[]> = {};

  if (!validScopes.includes(scope as ApiKeyScope)) {
    errors.scope = [`Scope must be one of: ${validScopes.join(", ")}`];
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

export function validateCreateApiKeyRequest(
  name: string,
  scope: string,
  expiresInDays?: number | null
): ValidationResult {
  const errors: Record<string, string[]> = {};

  const nameResult = validateApiKeyName(name);
  Object.assign(errors, nameResult.errors);

  const scopeResult = validateScope(scope);
  Object.assign(errors, scopeResult.errors);

  if (expiresInDays !== undefined && expiresInDays !== null) {
    const expiryResult = validateExpiryDays(expiresInDays);
    Object.assign(errors, expiryResult.errors);
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}
