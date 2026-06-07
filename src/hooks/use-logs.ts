import { useSyncExternalStore } from "react";

import {
  logger,
  LOG_SOURCES,
  LOG_LEVELS,
  type LogEntry,
  type LogLevel,
  type LogSource,
} from "@/lib/logger";

export function useLogs(): LogEntry[] {
  return useSyncExternalStore(
    (cb) => logger.subscribe(cb),
    () => logger.snapshot(),
    () => logger.snapshot(),
  );
}

export { logger, LOG_SOURCES, LOG_LEVELS };
export type { LogEntry, LogLevel, LogSource };
