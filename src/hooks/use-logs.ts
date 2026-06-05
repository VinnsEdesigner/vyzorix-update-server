import { useSyncExternalStore } from "react";
import { logger, type LogEntry } from "@/lib/logger";

export function useLogs(): LogEntry[] {
  return useSyncExternalStore(
    (cb) => logger.subscribe(cb),
    () => logger.snapshot(),
    () => logger.snapshot(),
  );
}

export { logger, LOG_SOURCES, LOG_LEVELS };
export type { LogEntry, LogLevel, LogSource };