/**
 * Use Pagination Hook
 * 
 * Generic pagination hook for offset-based pagination.
 * Works with any list data that uses page/limit/total.
 */

import { useState, useCallback, useMemo } from "react";
import type {
  OffsetPaginationParams,
  OffsetPagination,
} from "@/domain/_shared";

// ============================================================================
// Types
// ============================================================================

/**
 * Pagination state
 */
export interface PaginationState {
  page: number;
  limit: number;
}

/**
 * Use pagination options
 */
export interface UsePaginationOptions {
  initialPage?: number;
  initialLimit?: number;
  maxLimit?: number;
  onPageChange?: (page: number) => void;
  onLimitChange?: (limit: number) => void;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Generic pagination hook
 */
export function usePagination(options: UsePaginationOptions = {}) {
  const {
    initialPage = 1,
    initialLimit = 20,
    maxLimit = 100,
    onPageChange,
    onLimitChange,
  } = options;

  const [state, setState] = useState<PaginationState>({
    page: initialPage,
    limit: Math.min(initialLimit, maxLimit),
  });

  // Set page
  const setPage = useCallback(
    (page: number) => {
      const newPage = Math.max(1, page);
      setState((prev) => ({ ...prev, page: newPage }));
      onPageChange?.(newPage);
    },
    [onPageChange]
  );

  // Set limit
  const setLimit = useCallback(
    (limit: number) => {
      const newLimit = Math.min(Math.max(1, limit), maxLimit);
      setState((prev) => ({ ...prev, limit: newLimit, page: 1 })); // Reset to page 1
      onLimitChange?.(newLimit);
    },
    [maxLimit, onLimitChange]
  );

  // Navigation helpers
  const nextPage = useCallback(() => {
    setPage(state.page + 1);
  }, [state.page, setPage]);

  const prevPage = useCallback(() => {
    setPage(state.page - 1);
  }, [state.page, setPage]);

  const resetPagination = useCallback(() => {
    setState({ page: initialPage, limit: Math.min(initialLimit, maxLimit) });
  }, [initialPage, initialLimit, maxLimit]);

  // Derived state
  const params = useMemo<OffsetPaginationParams>(
    () => ({
      page: state.page,
      limit: state.limit,
    }),
    [state.page, state.limit]
  );

  return {
    // State
    page: state.page,
    limit: state.limit,
    params,

    // Setters
    setPage,
    setLimit,

    // Navigation
    nextPage,
    prevPage,
    resetPagination,

    // Helpers
    canGoNext: (pagination?: OffsetPagination) =>
      pagination ? pagination.page < pagination.totalPages : true,
    canGoPrev: () => state.page > 1,
  };
}

// ============================================================================
// Pagination Info Hook
// ============================================================================

/**
 * Calculate pagination display info
 */
export function usePaginationInfo(
  pagination: OffsetPagination | undefined,
  page: number,
  limit: number
) {
  return useMemo(() => {
    if (!pagination) {
      return {
        showingFrom: 0,
        showingTo: 0,
        total: 0,
        hasPagination: false,
      };
    }

    const showingFrom = (pagination.page - 1) * pagination.limit + 1;
    const showingTo = Math.min(
      pagination.page * pagination.limit,
      pagination.total
    );

    return {
      showingFrom,
      showingTo,
      total: pagination.total,
      hasPagination: pagination.totalPages > 1,
    };
  }, [pagination, page, limit]);
}

// ============================================================================
// URL Sync Helper
// ============================================================================

/**
 * Parse pagination from URL search params
 */
export function parsePaginationFromURL(
  searchParams: URLSearchParams
): OffsetPaginationParams {
  const page = searchParams.get("page");
  const limit = searchParams.get("limit");

  return {
    page: page ? parseInt(page, 10) : 1,
    limit: limit ? parseInt(limit, 10) : 20,
  };
}

/**
 * Update pagination in URL
 */
export function updatePaginationURL(
  page: number,
  limit: number
): URLSearchParams {
  const params = new URLSearchParams();
  if (page > 1) params.set("page", String(page));
  if (limit !== 20) params.set("limit", String(limit));
  return params;
}