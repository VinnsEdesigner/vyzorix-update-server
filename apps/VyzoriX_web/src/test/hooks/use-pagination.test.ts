import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  usePagination,
  usePaginationInfo,
  parsePaginationFromURL,
  updatePaginationURL,
} from '@/hooks/_shared/use-pagination';
import type { OffsetPagination } from '@vyzorix/api-client';

vi.mock('@vyzorix/api-client', () => ({}));

describe('usePagination', () => {
  it('starts with defaults (page 1, limit 20)', () => {
    const { result } = renderHook(() => usePagination());
    expect(result.current.page).toBe(1);
    expect(result.current.limit).toBe(20);
    expect(result.current.params).toEqual({ page: 1, limit: 20 });
  });

  it('respects initialPage and initialLimit', () => {
    const { result } = renderHook(() => usePagination({ initialPage: 3, initialLimit: 50 }));
    expect(result.current.page).toBe(3);
    expect(result.current.limit).toBe(50);
  });

  it('clamps limit to maxLimit', () => {
    const { result } = renderHook(() => usePagination({ initialLimit: 200, maxLimit: 100 }));
    expect(result.current.limit).toBe(100);
  });

  it('setPage does not go below 1', () => {
    const { result } = renderHook(() => usePagination());
    act(() => result.current.setPage(0));
    expect(result.current.page).toBe(1);
    act(() => result.current.setPage(-5));
    expect(result.current.page).toBe(1);
  });

  it('nextPage increments page', () => {
    const { result } = renderHook(() => usePagination());
    act(() => result.current.nextPage());
    expect(result.current.page).toBe(2);
  });

  it('prevPage decrements but not below 1', () => {
    const { result } = renderHook(() => usePagination({ initialPage: 3 }));
    act(() => result.current.prevPage());
    expect(result.current.page).toBe(2);
    act(() => result.current.prevPage());
    act(() => result.current.prevPage());
    expect(result.current.page).toBe(1);
  });

  it('setLimit resets page to 1', () => {
    const { result } = renderHook(() => usePagination({ initialPage: 5 }));
    act(() => result.current.setLimit(50));
    expect(result.current.limit).toBe(50);
    expect(result.current.page).toBe(1);
  });

  it('setLimit clamps to maxLimit', () => {
    const { result } = renderHook(() => usePagination({ maxLimit: 100 }));
    act(() => result.current.setLimit(500));
    expect(result.current.limit).toBe(100);
  });

  it('resetPagination restores initial values', () => {
    const { result } = renderHook(() => usePagination({ initialPage: 2, initialLimit: 30 }));
    act(() => result.current.setPage(10));
    act(() => result.current.resetPagination());
    expect(result.current.page).toBe(2);
    expect(result.current.limit).toBe(30);
  });

  it('canGoPrev returns false on page 1', () => {
    const { result } = renderHook(() => usePagination());
    expect(result.current.canGoPrev()).toBe(false);
  });

  it('canGoPrev returns true on page > 1', () => {
    const { result } = renderHook(() => usePagination({ initialPage: 2 }));
    expect(result.current.canGoPrev()).toBe(true);
  });

  it('canGoNext respects pagination total', () => {
    const { result } = renderHook(() => usePagination());
    const pagination: OffsetPagination = { page: 1, limit: 20, total: 40, totalPages: 2, hasMore: true };
    expect(result.current.canGoNext(pagination)).toBe(true);
    expect(result.current.canGoNext({ ...pagination, page: 2, hasMore: false })).toBe(false);
  });

  it('calls onPageChange callback', () => {
    const onPageChange = vi.fn();
    const { result } = renderHook(() => usePagination({ onPageChange }));
    act(() => result.current.setPage(5));
    expect(onPageChange).toHaveBeenCalledWith(5);
  });

  it('calls onLimitChange callback', () => {
    const onLimitChange = vi.fn();
    const { result } = renderHook(() => usePagination({ onLimitChange }));
    act(() => result.current.setLimit(50));
    expect(onLimitChange).toHaveBeenCalledWith(50);
  });
});

describe('usePaginationInfo', () => {
  it('returns zeros when no pagination', () => {
    const { result } = renderHook(() => usePaginationInfo(undefined, 1, 20));
    expect(result.current).toEqual({ showingFrom: 0, showingTo: 0, total: 0, hasPagination: false });
  });

  it('calculates showingFrom and showingTo', () => {
    const pagination: OffsetPagination = { page: 2, limit: 20, total: 50, totalPages: 3, hasMore: true };
    const { result } = renderHook(() => usePaginationInfo(pagination, 2, 20));
    expect(result.current.showingFrom).toBe(21);
    expect(result.current.showingTo).toBe(40);
    expect(result.current.total).toBe(50);
    expect(result.current.hasPagination).toBe(true);
  });

  it('showingTo is capped at total', () => {
    const pagination: OffsetPagination = { page: 3, limit: 20, total: 50, totalPages: 3, hasMore: true };
    const { result } = renderHook(() => usePaginationInfo(pagination, 3, 20));
    expect(result.current.showingTo).toBe(50);
  });

  it('hasPagination is false when totalPages is 1', () => {
    const pagination: OffsetPagination = { page: 1, limit: 20, total: 5, totalPages: 1, hasMore: false };
    const { result } = renderHook(() => usePaginationInfo(pagination, 1, 20));
    expect(result.current.hasPagination).toBe(false);
  });
});

describe('parsePaginationFromURL', () => {
  it('defaults to page 1, limit 20 when no params', () => {
    expect(parsePaginationFromURL(new URLSearchParams())).toEqual({ page: 1, limit: 20 });
  });

  it('parses page and limit', () => {
    expect(parsePaginationFromURL(new URLSearchParams('page=3&limit=50'))).toEqual({ page: 3, limit: 50 });
  });
});

describe('updatePaginationURL', () => {
  it('omits defaults', () => {
    const params = updatePaginationURL(1, 20);
    expect(params.toString()).toBe('');
  });

  it('includes page when > 1', () => {
    const params = updatePaginationURL(3, 20);
    expect(params.get('page')).toBe('3');
    expect(params.get('limit')).toBeNull();
  });

  it('includes limit when != 20', () => {
    const params = updatePaginationURL(1, 50);
    expect(params.get('limit')).toBe('50');
    expect(params.get('page')).toBeNull();
  });
});
