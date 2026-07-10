/**
 * Domain Error Types
 * 
 * Shared error types used across all features.
 * Domain errors are pure TypeScript - no external imports.
 */

// ============================================================================
// Error Codes
// ============================================================================

/**
 * Standard error codes
 */
export const ERROR_CODES = {
  // Authentication
  UNAUTHORIZED: "unauthorized",
  FORBIDDEN: "forbidden",
  SESSION_EXPIRED: "session_expired",
  
  // Resource
  NOT_FOUND: "not_found",
  ALREADY_EXISTS: "already_exists",
  
  // Validation
  VALIDATION_ERROR: "validation_error",
  INVALID_INPUT: "invalid_input",
  
  // Request
  BAD_REQUEST: "bad_request",
  TIMEOUT: "timeout",
  
  // Server
  INTERNAL_ERROR: "internal_error",
  SERVICE_UNAVAILABLE: "service_unavailable",
} as const;

export type ErrorCode = typeof ERROR_CODES[keyof typeof ERROR_CODES];

// ============================================================================
// Base Error Classes
// ============================================================================

/**
 * Base domain error
 */
export class DomainError extends Error {
  public readonly code: ErrorCode;
  public readonly details?: Record<string, unknown>;
  public readonly statusCode: number;

  constructor(
    message: string,
    code: ErrorCode,
    statusCode: number = 500,
    details?: Record<string, unknown>
  ) {
    super(message);
    this.name = "DomainError";
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
  }

  toJSON(): Record<string, unknown> {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      statusCode: this.statusCode,
      details: this.details,
    };
  }
}

/**
 * Validation error with field-specific messages
 */
export class ValidationError extends DomainError {
  public readonly fieldErrors: Record<string, string[]>;

  constructor(
    message: string,
    fieldErrors: Record<string, string[]>
  ) {
    super(message, ERROR_CODES.VALIDATION_ERROR, 400, { fieldErrors });
    this.name = "ValidationError";
    this.fieldErrors = fieldErrors;
  }

  /**
   * Create from a single field error
   */
  static fromField(field: string, message: string): ValidationError {
    return new ValidationError(message, { [field]: [message] });
  }

  /**
   * Create from multiple field errors
   */
  static fromFields(errors: Record<string, string>): ValidationError {
    const fieldErrors = Object.entries(errors).reduce(
      (acc, [field, message]) => {
        acc[field] = [message];
        return acc;
      },
      {} as Record<string, string[]>
    );
    return new ValidationError("Validation failed", fieldErrors);
  }

  toJSON(): Record<string, unknown> {
    return {
      ...super.toJSON(),
      fieldErrors: this.fieldErrors,
    };
  }
}

/**
 * Not found error
 */
export class NotFoundError extends DomainError {
  constructor(resource: string, identifier?: string) {
    const message = identifier
      ? `${resource} '${identifier}' not found`
      : `${resource} not found`;
    super(message, ERROR_CODES.NOT_FOUND, 404);
    this.name = "NotFoundError";
  }
}

/**
 * Unauthorized error
 */
export class UnauthorizedError extends DomainError {
  constructor(message: string = "Authentication required") {
    super(message, ERROR_CODES.UNAUTHORIZED, 401);
    this.name = "UnauthorizedError";
  }
}

/**
 * Forbidden error
 */
export class ForbiddenError extends DomainError {
  constructor(message: string = "Access denied") {
    super(message, ERROR_CODES.FORBIDDEN, 403);
    this.name = "ForbiddenError";
  }
}

/**
 * Network/timeout error
 */
export class NetworkError extends DomainError {
  constructor(message: string = "Network request failed") {
    super(message, ERROR_CODES.TIMEOUT, 0);
    this.name = "NetworkError";
  }
}

// ============================================================================
// HTTP Error Factory
// ============================================================================

/**
 * Raw API error response
 */
export interface RawAPIError {
  error?: string;
  message?: string;
  code?: string;
}

/**
 * Create domain error from HTTP response
 */
export function errorFromHTTP(
  status: number,
  data?: RawAPIError
): DomainError {
  const message = data?.message ?? data?.error ?? "An error occurred";
  
  switch (status) {
    case 400:
      return new DomainError(message, ERROR_CODES.BAD_REQUEST, status);
    case 401:
      return new UnauthorizedError(message);
    case 403:
      return new ForbiddenError(message);
    case 404:
      return new NotFoundError("Resource");
    case 422:
      if (data?.error === "validation_error") {
        return new ValidationError(message, {});
      }
      return new DomainError(message, ERROR_CODES.VALIDATION_ERROR, status);
    case 429:
      return new DomainError("Rate limit exceeded", ERROR_CODES.BAD_REQUEST, status);
    case 500:
    default:
      return new DomainError(message, ERROR_CODES.INTERNAL_ERROR, status);
  }
}

// ============================================================================
// Type Guards
// ============================================================================

/**
 * Check if error is a domain error
 */
export function isDomainError(error: unknown): error is DomainError {
  return error instanceof DomainError;
}

/**
 * Check if error is a validation error
 */
export function isValidationError(error: unknown): error is ValidationError {
  return error instanceof ValidationError;
}

/**
 * Check if error is a network error
 */
export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof NetworkError;
}
