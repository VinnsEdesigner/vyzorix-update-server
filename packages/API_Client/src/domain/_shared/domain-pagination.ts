






export interface RawOffsetPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}


export interface OffsetPaginationParams {
  page?: number;
  limit?: number;
}


export interface OffsetPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasMore: boolean;
}






export interface RawCursorPagination {
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
}


export interface CursorPaginationParams {
  cursor?: string | null;
  limit?: number;
}


export interface CursorPagination {
  limit: number;
  hasMore: boolean;
  nextCursor: string | null;
}






export interface PaginatedResult<T> {
  items: T[];
  pagination: OffsetPagination | CursorPagination;
}


export function isOffsetPagination(
  pagination: OffsetPagination | CursorPagination
): pagination is OffsetPagination {
  return "page" in pagination && "total" in pagination;
}


export function isCursorPagination(
  pagination: OffsetPagination | CursorPagination
): pagination is CursorPagination {
  return "nextCursor" in pagination;
}






export function offsetPaginationFromRaw(raw: RawOffsetPagination): OffsetPagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.total_pages,
    hasMore: raw.page < raw.total_pages,
  };
}


export function cursorPaginationFromRaw(raw: RawCursorPagination): CursorPagination {
  return {
    limit: raw.limit,
    hasMore: raw.has_more,
    nextCursor: raw.next_cursor,
  };
}






export const DEFAULT_PAGINATION: OffsetPaginationParams = {
  page: 1,
  limit: 20,
};


export const PAGINATION_LIMITS = {
  default: 100,
  logs: 500,
  telemetry: 1000,
  commands: 100,
} as const;


export function clampLimit(limit: number, max: number = PAGINATION_LIMITS.default): number {
  return Math.min(Math.max(1, limit), max);
}


export function calculateOffset(page: number, limit: number): number {
  return (Math.max(1, page) - 1) * limit;
}
