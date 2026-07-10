/**
 * Domain Pagination Types
 * 
 * Shared pagination types used across all features.
 * Raw API uses snake_case, domain uses camelCase.
 */

// ============================================================================
// Offset-based Pagination (Page/Limit)
// ============================================================================

/**
 * Raw API response for offset pagination
 */
export interface RawOffsetPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

/**
 * Offset pagination parameters
 */
export interface OffsetPaginationParams {
  page?: number;
  limit?: number;
}

/**
 * Standard offset pagination result
 */
export interface OffsetPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasMore: boolean;
}

// ============================================================================
// Cursor-based Pagination
// ============================================================================

/**
 * Raw API response for cursor pagination
 */
export interface RawCursorPagination {
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
}

/**
 * Cursor pagination parameters
 */
export interface CursorPaginationParams {
  cursor?: string | null;
  limit?: number;
}

/**
 * Standard cursor pagination result
 */
export interface CursorPagination {
  limit: number;
  hasMore: boolean;
  nextCursor: string | null;
}

// ============================================================================
// Generic Paginated Result
// ============================================================================

/**
 * Generic paginated result wrapper
 */
export interface PaginatedResult<T> {
  items: T[];
  pagination: OffsetPagination | CursorPagination;
}

/**
 * Check if pagination is offset-based
 */
export function isOffsetPagination(
  pagination: OffsetPagination | CursorPagination
): pagination is OffsetPagination {
  return "page" in pagination && "total" in pagination;
}

/**
 * Check if pagination is cursor-based
 */
export function isCursorPagination(
  pagination: OffsetPagination | CursorPagination
): pagination is CursorPagination {
  return "nextCursor" in pagination;
}

// ============================================================================
// Transform Functions (Raw API → Domain)
// ============================================================================

/**
 * Transform raw offset pagination to domain
 */
export function offsetPaginationFromRaw(raw: RawOffsetPagination): OffsetPagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.total_pages,
    hasMore: raw.page < raw.total_pages,
  };
}

/**
 * Transform raw cursor pagination to domain
 */
export function cursorPaginationFromRaw(raw: RawCursorPagination): CursorPagination {
  return {
    limit: raw.limit,
    hasMore: raw.has_more,
    nextCursor: raw.next_cursor,
  };
}

// ============================================================================
// Default Values
// ============================================================================

/**
 * Default pagination values
 */
export const DEFAULT_PAGINATION: OffsetPaginationParams = {
  page: 1,
  limit: 20,
};

/**
 * Maximum pagination limits by feature
 */
export const PAGINATION_LIMITS = {
  default: 100,
  logs: 500,
  telemetry: 1000,
  commands: 100,
} as const;

/**
 * Clamp pagination limit to maximum
 */
export function clampLimit(limit: number, max: number = PAGINATION_LIMITS.default): number {
  return Math.min(Math.max(1, limit), max);
}

/**
 * Calculate offset from page and limit
 */
export function calculateOffset(page: number, limit: number): number {
  return (Math.max(1, page) - 1) * limit;
}
