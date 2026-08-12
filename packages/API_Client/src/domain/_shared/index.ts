export {
  type RawOffsetPagination,
  type RawCursorPagination,
  type OffsetPaginationParams,
  type CursorPaginationParams,
  type OffsetPagination,
  type CursorPagination,
  type PaginatedResult,
  type Pagination,
  type RawPagination,
  isOffsetPagination,
  isCursorPagination,
  offsetPaginationFromRaw,
  cursorPaginationFromRaw,
  paginationFromRaw,
  clampLimit,
  calculateOffset,
  DEFAULT_PAGINATION,
  PAGINATION_LIMITS,
} from "./domain-pagination";

export {
  type DeviceStatus,
  type MetricThreshold,
} from "./domain-shared";

export {
  type ValidationFieldError,
  type ValidationResult,
  validResult,
  invalidResult,
  fieldErrorList,
  validateIMEI,
} from "./domain-validation";

export {
  ERROR_CODES,
  type ErrorCode,
  DomainError,
  ValidationError,
  NotFoundError,
  UnauthorizedError,
  ForbiddenError,
  NetworkError,
  errorFromHTTP,
  type RawAPIError,
  isDomainError,
  isValidationError,
  isNetworkError,
} from "./domain-errors";
