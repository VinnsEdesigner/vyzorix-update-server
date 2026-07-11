export type LogEventType =
  | "connection"
  | "command"
  | "telemetry"
  | "error"
  | "warning"
  | "info"
  | "update"
  | "registration"
  | "deregistration";

export interface LogEntry {
  id: string;
  deviceId: string;
  eventType: LogEventType;
  timestamp: Date;
  data?: Record<string, unknown>;
}

export interface LogListResult {
  logs: LogEntry[];
  hasMore: boolean;
  nextCursor?: string;
}

export interface LogStats {
  total: number;
  byType: {
    connection: number;
    command: number;
    telemetry: number;
    error: number;
    warning: number;
  };
}

export interface LogFilters {
  eventType?: LogEventType;
  startTime?: Date;
  endTime?: Date;
}

export const LOG_EVENT_TYPES: Record<LogEventType, { label: string }> = {
  connection: { label: "Connection" },
  command: { label: "Command" },
  telemetry: { label: "Telemetry" },
  error: { label: "Error" },
  warning: { label: "Warning" },
  info: { label: "Info" },
  update: { label: "Update" },
  registration: { label: "Registration" },
  deregistration: { label: "Deregistration" },
};

export function getEventTypeLabel(type: LogEventType): string {
  return LOG_EVENT_TYPES[type].label;
}
