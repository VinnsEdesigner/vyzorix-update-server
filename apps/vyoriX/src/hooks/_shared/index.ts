/**
 * Hooks Shared Index
 * 
 * Re-exports all shared hooks for convenient imports.
 */

// Pagination
export {
  usePagination,
  usePaginationInfo,
  parsePaginationFromURL,
  updatePaginationURL,
  type PaginationState,
  type UsePaginationOptions,
} from "./use-pagination";

// Time Range
export {
  useTimeRange,
  useResolution,
  calculateResolution,
  parseTimeRangeFromURL,
  updateTimeRangeURL,
  formatTimeRange,
  formatTimestamp,
  TIME_RANGES,
  type TimeRange,
  type TimeRangeKey,
  type TimeRangeValue,
  type UseTimeRangeOptions,
} from "./use-time-range";

// Debounce
export {
  useDebounce,
  useDebouncedCallback,
  useDebounceAsync,
  useThrottle,
  useMediaQuery,
  useLocalStorage,
  type UseDebounceOptions,
} from "./use-debounce";