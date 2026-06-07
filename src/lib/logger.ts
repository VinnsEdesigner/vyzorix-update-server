// App-wide log bus. Singleton outside React so any module (api client, ws,
// command dispatch, threshold derivation) can emit without prop-drilling.
// Components subscribe via subscribe() and get a stable, throttled snapshot.

export type LogLevel = "debug" | "info" | "warn" | "error";
export type LogSource =
  | "system"
  | "ws"
  | "api"
  | "command"
  | "update"
  | "device"
  | "alert"
  | "auth";

export interface LogEntry {
  id: number;
  t: number;
  level: LogLevel;
  source: LogSource;
  message: string;
  meta?: Record<string, unknown>;
}

const STORAGE_KEY = "vyzorix.logs.v1";
const DEFAULT_LIMIT = 1000;
const PERSIST_LIMIT = 300; // keep last N across reloads

let nextId = 1;
let buffer: LogEntry[] = [];
let limit = DEFAULT_LIMIT;
const listeners = new Set<() => void>();

// Hydrate from localStorage so logs survive reloads.
const hydrate = (): void => {
  if (typeof window === "undefined") return;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return;
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      buffer = parsed.slice(-limit);
      nextId = (buffer[buffer.length - 1]?.id ?? 0) + 1;
    }
  } catch {
    // ignore
  }
};
hydrate();

let persistTimer: ReturnType<typeof setTimeout> | null = null;
const schedulePersist = (): void => {
  if (typeof window === "undefined") return;
  if (persistTimer) return;
  persistTimer = setTimeout(() => {
    persistTimer = null;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(buffer.slice(-PERSIST_LIMIT)));
    } catch {
      // quota / serialize fail — drop silently
    }
  }, 500);
};

const emit = (): void => {
  for (const l of listeners) l();
};

export const logger = {
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  setLimit(n: number) {
    limit = Math.max(50, Math.min(n, 10000));
    if (buffer.length > limit) buffer = buffer.slice(-limit);
    emit();
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  clear() {
    buffer = [];
    schedulePersist();
    emit();
  },
  snapshot(): LogEntry[] {
    return buffer;
  },
  subscribe(cb: () => void) {
    listeners.add(cb);
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
    return () => listeners.delete(cb);
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  log(level: LogLevel, source: LogSource, message: string, meta?: Record<string, unknown>) {
    const entry: LogEntry = { id: nextId++, t: Date.now(), level, source, message, meta };
    buffer = buffer.length >= limit ? [...buffer.slice(-(limit - 1)), entry] : [...buffer, entry];
    schedulePersist();
    emit();
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  debug(source: LogSource, m: string, meta?: Record<string, unknown>) {
    this.log("debug", source, m, meta);
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  info(source: LogSource, m: string, meta?: Record<string, unknown>) {
    this.log("info", source, m, meta);
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  warn(source: LogSource, m: string, meta?: Record<string, unknown>) {
    this.log("warn", source, m, meta);
  },
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  error(source: LogSource, m: string, meta?: Record<string, unknown>) {
    this.log("error", source, m, meta);
  },
};

// Initial bootstrap line (only the first time after hydrate adds nothing).
if (buffer.length === 0) {
  logger.info("system", "Vyzorix dashboard initialized");
}

export const LOG_SOURCES: LogSource[] = [
  "system",
  "ws",
  "api",
  "command",
  "update",
  "device",
  "alert",
  "auth",
];
export const LOG_LEVELS: LogLevel[] = ["debug", "info", "warn", "error"];
