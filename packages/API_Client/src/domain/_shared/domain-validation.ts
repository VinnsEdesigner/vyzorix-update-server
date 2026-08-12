// Shared validation primitives. ValidationResult is the canonical shape
// across all bounded contexts; per-context validators return this type so
// the domain barrel can re-export them without name collisions.

export interface ValidationFieldError {
  field: string;
  message: string;
}

/**
 * Unified validation result. `errors` is keyed by field name, where each
 * field may carry one or more messages. Callers needing a flat list can
 * use {@link fieldErrorList}.
 */
export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string[]>;
}

export function validResult(): ValidationResult {
  return { isValid: true, errors: {} };
}

export function invalidResult(errors: Record<string, string[]>): ValidationResult {
  return { isValid: Object.keys(errors).length === 0, errors };
}

export function fieldErrorList(result: ValidationResult): ValidationFieldError[] {
  return Object.entries(result.errors).flatMap(([field, messages]) =>
    messages.map((message) => ({ field, message }))
  );
}

const IMEI_REGEX = /^\d{15}$/;

export function validateIMEI(imei: string): ValidationResult {
  const errors: Record<string, string[]> = {};
  if (!imei) {
    errors.imei = ["IMEI is required"];
  } else if (!IMEI_REGEX.test(imei)) {
    errors.imei = ["IMEI must be 15 digits"];
  }
  return { isValid: Object.keys(errors).length === 0, errors };
}
