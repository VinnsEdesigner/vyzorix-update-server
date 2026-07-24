import { restClient } from "../_shared/rest-client";
import {
  inboxEntryFromRaw,
  deviceFromRaw,
  paginationFromRaw,
  createInboxRequestToRaw,
  createInboxResultFromRaw,
  confirmDeviceResultFromRaw,
  type RawInboxEntry,
  type RawDevice,
  type RawPagination,
  type RawCreateInboxRequest,
  type RawCreateInboxResponse,
  type RawConfirmDeviceResponse,
} from "@/domain/registration";
import type {
  InboxEntry,
  Device,
  InboxListResult,
  DeviceListResult,
  AckResult,
  DeregisterResult,
  AcknowledgeAction,
  InboxStatus,
  CreateInboxRequest,
  CreateInboxResult,
  ConfirmDeviceResult,
} from "@/domain/registration";

const PATHS = {
  inbox: "/v1/device/inbox",
  inboxEntry: (imei: string) => `/v1/device/inbox/${imei}`,
  inboxAck: (imei: string) => `/v1/device/inbox/${imei}/ack`,
  confirm: "/v1/device/confirm",
  devices: "/v1/devices",
  device: (imei: string) => `/v1/devices/${imei}`,
} as const;

interface RawInboxListResponse {
  requests: RawInboxEntry[];
  pagination: RawPagination;
}

interface RawDeviceListResponse {
  devices: RawDevice[];
  pagination: RawPagination;
}

interface RawAckResponse {
  id: string;
  imei: string;
  status: InboxStatus;
  acknowledgedAt: number | null;
  approvingAt: number | null;
  approvedAt: number | null;
  rejectedAt: number | null;
  commandSecret: string | null;
  fcmPushSent: boolean;
  notes: string | null;
}

interface RawDeregisterResponse {
  imei: string;
  status: string;
  deregisteredAt: number;
  retentionUntil: number;
}

export const registration = {
  
  async createInboxRequest(request: CreateInboxRequest): Promise<CreateInboxResult> {
    const rawRequest = createInboxRequestToRaw(request);
    const response = await restClient.post<RawCreateInboxResponse>(
      PATHS.inbox,
      rawRequest
    );
    return createInboxResultFromRaw(response);
  },

  
  async confirmDevice(imei: string, commandSecret: string): Promise<ConfirmDeviceResult> {
    const response = await restClient.post<RawConfirmDeviceResponse>(PATHS.confirm, {
      imei,
      commandSecret,
    });
    return confirmDeviceResultFromRaw(response);
  },

  async getInbox(params?: {
    status?: InboxStatus | "all";
    page?: number;
    limit?: number;
  }): Promise<InboxListResult> {
    const response = await restClient.get<RawInboxListResponse>(PATHS.inbox, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
      },
    });
    return {
      requests: response.requests.map(inboxEntryFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getInboxEntry(imei: string): Promise<InboxEntry | null> {
    const response = await restClient.get<RawInboxEntry | null>(PATHS.inboxEntry(imei));
    if (!response?.imei) return null;
    return inboxEntryFromRaw(response);
  },

  async acknowledgeInbox(
    imei: string,
    action: AcknowledgeAction,
    notes?: string
  ): Promise<AckResult> {
    const response = await restClient.post<RawAckResponse>(PATHS.inboxAck(imei), {
      action,
      notes,
    });
    return {
      id: response.id,
      imei: response.imei,
      status: response.status,
      acknowledgedAt: response.acknowledgedAt ? new Date(response.acknowledgedAt) : null,
      approvingAt: response.approvingAt ? new Date(response.approvingAt) : null,
      approvedAt: response.approvedAt ? new Date(response.approvedAt) : null,
      rejectedAt: response.rejectedAt ? new Date(response.rejectedAt) : null,
      commandSecret: response.commandSecret,
      fcmPushSent: response.fcmPushSent,
      notes: response.notes,
    };
  },

  async dismissInbox(imei: string): Promise<{ status: InboxStatus }> {
    const response = await restClient.delete<{ status: InboxStatus }>(PATHS.inboxEntry(imei));
    return { status: response.status };
  },

  async getDevices(params?: {
    status?: string;
    page?: number;
    limit?: number;
  }): Promise<DeviceListResult> {
    const response = await restClient.get<RawDeviceListResponse>(PATHS.devices, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
      },
    });
    return {
      devices: response.devices.map(deviceFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getDevice(imei: string): Promise<Device | null> {
    const response = await restClient.get<RawDevice | null>(PATHS.device(imei));
    if (!response?.imei) return null;
    return deviceFromRaw(response);
  },

  async deregisterDevice(imei: string): Promise<DeregisterResult> {
    const response = await restClient.delete<RawDeregisterResponse>(PATHS.device(imei));
    return {
      imei: response.imei,
      status: response.status,
      deregisteredAt: new Date(response.deregisteredAt),
      retentionUntil: new Date(response.retentionUntil),
    };
  },
};

export type {
  InboxStatus,
  AcknowledgeAction,
  CreateInboxRequest,
  CreateInboxResult,
  ConfirmDeviceResult,
};
