import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  deviceInspectionFromRaw,
  timelineResultFromRaw,
  type RawDeviceInspection,
  type RawTimelineResult,
} from "../../../domain/diagnostics";
import type {
  DeviceInspection,
  TimelineResult,
  TimelineEventType,
} from "../../../domain/diagnostics";

const PATHS = {
  inspect: (imei: string) => `/v1/device/${imei}/inspect`,
  timeline: (imei: string) => `/v1/device/${imei}/timeline`,
} as const;

export const diagnostics = {
  async inspectDevice(imei: string, organizationId?: string): Promise<DeviceInspection> {
    const response = await restClient.get<RawDeviceInspection>(PATHS.inspect(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return deviceInspectionFromRaw(response);
  },

  async getTimeline(imei: string, params?: {
    eventType?: TimelineEventType;
    startTime?: number;
    endTime?: number;
    cursor?: string;
    limit?: number;
    organizationId?: string;
  }): Promise<TimelineResult> {
    const response = await restClient.get<RawTimelineResult>(PATHS.timeline(imei), {
      params: {
        event_type: params?.eventType,
        start_time: params?.startTime,
        end_time: params?.endTime,
        cursor: params?.cursor,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return timelineResultFromRaw(response);
  },
};
