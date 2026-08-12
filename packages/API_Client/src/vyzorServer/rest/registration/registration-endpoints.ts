import { restClient, getCSRFToken, fetchAndSetCSRFToken, getOrganizationContext } from "../_shared/rest-client";
import {
  inboxEntryFromRaw,
  registrationDeviceFromRaw,
  paginationFromRaw,
  createInboxRequestToRaw,
  createInboxResultFromRaw,
  confirmDeviceResultFromRaw,
  type RawInboxEntry,
  type RawRegisteredDevice,
  type RawPagination,
  type RawCreateInboxResponse,
  type RawConfirmDeviceResponse,
} from "@/domain/registration";
import type {
  InboxEntry,
  RegisteredDevice,
  InboxListResult,
  RegisteredDeviceListResult,
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
  inboxEntry: (imei: string) => `/v1/inbox/${imei}`,
  inboxAck: (imei: string) => `/v1/inbox/${imei}/ack`,
  inboxResend: (imei: string) => `/v1/inbox/${imei}/resend`,
  confirm: "/v1/device/confirm",
  devices: "/v1/devices",
  device: (imei: string) => `/v1/devices/${imei}`,
} as const;

async function ensureCSRFToken(): Promise<void> {
  if (!getCSRFToken()) {
    await fetchAndSetCSRFToken();
  }
}

interface RawInboxListResponse {
  requests: RawInboxEntry[];
  pagination: RawPagination;
}

interface RawDeviceListResponse {
  devices: RawRegisteredDevice[];
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
  
  async createInboxRequest(request: CreateInboxRequest, organizationId?: string): Promise<CreateInboxResult> {
    await ensureCSRFToken();
    const rawRequest = createInboxRequestToRaw(request);
    const response = await restClient.post<RawCreateInboxResponse>(
      PATHS.inbox,
      rawRequest,
      { params: { organization_id: organizationId || getOrganizationContext() } }
    );
    return createInboxResultFromRaw(response);
  },

  
  async confirmDevice(imei: string, commandSecret: string, organizationId?: string): Promise<ConfirmDeviceResult> {
    await ensureCSRFToken();
    const response = await restClient.post<RawConfirmDeviceResponse>(PATHS.confirm, {
      imei,
      commandSecret,
    }, { params: { organization_id: organizationId || getOrganizationContext() } });
    return confirmDeviceResultFromRaw(response);
  },

  async getInbox(params?: {
    status?: InboxStatus | "all";
    page?: number;
    limit?: number;
    organizationId?: string;
  }): Promise<InboxListResult> {
    const response = await restClient.get<RawInboxListResponse>(PATHS.inbox, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return {
      requests: response.requests.map(inboxEntryFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getInboxEntry(imei: string, organizationId?: string): Promise<InboxEntry | null> {
    const response = await restClient.get<RawInboxEntry | null>(PATHS.inboxEntry(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    if (!response?.imei) return null;
    return inboxEntryFromRaw(response);
  },

  async acknowledgeInbox(
    imei: string,
    action: AcknowledgeAction,
    notes?: string,
    organizationId?: string
  ): Promise<AckResult> {
    const response = await restClient.post<RawAckResponse>(PATHS.inboxAck(imei), {
      action,
      notes,
    }, { params: { organization_id: organizationId || getOrganizationContext() } });
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

  async resendInboxApproval(imei: string, organizationId?: string): Promise<{ success: boolean; message: string }> {
    return restClient.post<{ success: boolean; message: string }>(PATHS.inboxResend(imei), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async dismissInbox(imei: string, organizationId?: string): Promise<{ status: InboxStatus }> {
    const response = await restClient.delete<{ status: InboxStatus }>(PATHS.inboxEntry(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return { status: response.status };
  },

  async getDevices(params?: {
    status?: string;
    page?: number;
    limit?: number;
    organizationId?: string;
  }): Promise<RegisteredDeviceListResult> {
    const response = await restClient.get<RawDeviceListResponse>(PATHS.devices, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return {
      devices: response.devices.map(registrationDeviceFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getDevice(imei: string, organizationId?: string): Promise<RegisteredDevice | null> {
    const response = await restClient.get<RawRegisteredDevice | null>(PATHS.device(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    if (!response?.imei) return null;
    return registrationDeviceFromRaw(response);
  },

  async deregisterDevice(imei: string, organizationId?: string): Promise<DeregisterResult> {
    const response = await restClient.delete<RawDeregisterResponse>(PATHS.device(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
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
