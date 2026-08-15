import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  eventFromRaw,
  eventResultFromRaw,
  eventsFromRaw,
  type RawEvent,
  type RawEventResult,
  type EventParams,
} from "../../../domain/events";
import type { Event, EventResult } from "../../../domain/events";

const PATHS = {
  deviceEvents: (imei: string) => `/v1/dashboard/device/${imei}/events`,
  recentEvents: "/v1/dashboard/events/recent",
  eventsByType: (type: string) => `/v1/dashboard/events/types/${type}`,
  eventById: (id: string) => `/v1/dashboard/events/${id}`,
} as const;

export { type EventParams };

export const events = {
  async getDeviceEvents(imei: string, params?: EventParams & { organizationId?: string }): Promise<EventResult> {
    const response = await restClient.get<RawEventResult>(PATHS.deviceEvents(imei), {
      params: {
        types: params?.types,
        severities: params?.severities,
        limit: params?.limit,
        offset: params?.offset,
        startTime: params?.startTime,
        endTime: params?.endTime,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return eventResultFromRaw(response);
  },

  async getRecentEvents(limit?: number, organizationId?: string): Promise<Event[]> {
    const response = await restClient.get<{ events: RawEvent[] }>(PATHS.recentEvents, {
      params: { limit, organization_id: organizationId || getOrganizationContext() },
    });
    return eventsFromRaw(response.events);
  },

  async getEventsByType(type: string, params?: Omit<EventParams, "types"> & { organizationId?: string }): Promise<EventResult> {
    const response = await restClient.get<RawEventResult>(PATHS.eventsByType(type), {
      params: {
        severities: params?.severities,
        limit: params?.limit,
        offset: params?.offset,
        startTime: params?.startTime,
        endTime: params?.endTime,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return eventResultFromRaw(response);
  },

  async getById(id: string, organizationId?: string): Promise<Event> {
    const response = await restClient.get<RawEvent>(PATHS.eventById(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return eventFromRaw(response);
  },
};
