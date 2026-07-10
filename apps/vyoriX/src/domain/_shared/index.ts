/**
 * Domain Shared Types
 * 
 * Re-exports all shared domain types for convenient imports.
 */

// Pagination
export {
  // Types
  type RawOffsetPagination,
  type RawCursorPagination,
  type OffsetPaginationParams,
  type CursorPaginationParams,
  type OffsetPagination,
  type CursorPagination,
  type PaginatedResult,
  // Functions
  isOffsetPagination,
  isCursorPagination,
  offsetPaginationFromRaw,
  cursorPaginationFromRaw,
  clampLimit,
  calculateOffset,
  // Constants
  DEFAULT_PAGINATION,
  PAGINATION_LIMITS,
} from "./domain-pagination";

// Errors
export {
  // Constants
  ERROR_CODES,
  type ErrorCode,
  // Classes
  DomainError,
  ValidationError,
  NotFoundError,
  UnauthorizedError,
  ForbiddenError,
  NetworkError,
  // Factory
  errorFromHTTP,
  type RawAPIError,
  // Type guards
  isDomainError,
  isValidationError,
  isNetworkError,
} from "./domain-errors";