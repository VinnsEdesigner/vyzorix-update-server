

import type {
  Event,
  EventResult,
  EventType,
  Severity,
} from "./events-entity";


export interface RawEvent {
  id: string;
  deviceId: string;
  operatorId?: string;
  type: string;
  severity: string;
  timestamp: string | number;
  data?: Record<string, unknown>;
  source?: string;
}


export interface RawEventResult {
  events: RawEvent[];
  hasMore: boolean;
  count: number;
  totalCount?: number;
}


function parseTimestamp(value: string | number | undefined): Date {
  if (!value) return new Date();
  if (typeof value === "number") {
        return new Date(value > 1e12 ? value : value * 1000);
  }
  return new Date(value);
}


export function eventFromRaw(raw: RawEvent): Event {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    operatorId: raw.operatorId,
    type: raw.type as EventType,
    severity: (raw.severity || "info") as Severity,
    timestamp: parseTimestamp(raw.timestamp),
    data: raw.data,
    source: (raw.source as Event["source"]) || "server",
  };
}


export function eventResultFromRaw(raw: RawEventResult): EventResult {
  return {
    events: raw.events.map(eventFromRaw),
    hasMore: raw.hasMore,
    count: raw.count,
    totalCount: raw.totalCount,
  };
}


export function eventsFromRaw(raw: RawEvent[]): Event[] {
  return raw.map(eventFromRaw);
}
