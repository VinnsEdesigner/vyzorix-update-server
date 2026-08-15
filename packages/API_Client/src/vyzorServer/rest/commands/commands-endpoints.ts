import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  commandFromRaw,
  commandListItemFromRaw,
  paginationFromRaw,
  sendCommandRequestToRaw,
  type RawCommand,
  type RawCommandListItem,
  type RawCommandHistoryResult,
} from "../../../domain/commands";
import type { Command, CommandListItem, SendCommandRequest, CommandStatus } from "../../../domain/commands";

const PATHS = {
  history: (imei: string) => `/v1/dashboard/device/${imei}/commands`,
  send: (imei: string) => `/v1/device/${imei}/command`,
  cancel: (dispatchId: string) => `/v1/command/${dispatchId}`,
  pending: (imei: string) => `/v1/device/${imei}/commands/pending`,
  status: (dispatchId: string) => `/v1/command/${dispatchId}/status`,
  retry: (dispatchId: string) => `/v1/command/${dispatchId}/retry`,
} as const;

export interface CommandHistoryParams {
  status?: CommandStatus;
  page?: number;
  limit?: number;
  startTime?: number;
  endTime?: number;
  organizationId?: string;
}

export const commands = {
  async getHistory(imei: string, params?: CommandHistoryParams) {
    const response = await restClient.get<RawCommandHistoryResult>(PATHS.history(imei), {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        start_time: params?.startTime,
        end_time: params?.endTime,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return {
      commands: response.commands.map(commandListItemFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getPending(imei: string, organizationId?: string): Promise<CommandListItem[]> {
    const response = await restClient.get<{ commands: RawCommandListItem[] }>(PATHS.pending(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return response.commands.map(commandListItemFromRaw);
  },

  async getByDispatchId(dispatchId: string, organizationId?: string): Promise<Command | null> {
    const response = await restClient.get<RawCommand | null>(PATHS.status(dispatchId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    if (!response?.id) return null;
    return commandFromRaw(response);
  },

  async send(request: SendCommandRequest, organizationId?: string): Promise<Command> {
    const payload = sendCommandRequestToRaw(request);
    const response = await restClient.post<{
      success: boolean;
      dispatchId?: string;
      command?: RawCommand;
      error?: string;
    }>(PATHS.send(request.imei), payload, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });

    if (!response.success || !response.command) {
      throw new Error(response.error ?? "Failed to send command");
    }

    return commandFromRaw(response.command);
  },

  async cancel(dispatchId: string, organizationId?: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.cancel(dispatchId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async retry(dispatchId: string, organizationId?: string): Promise<Command> {
    const response = await restClient.post<{
      success: boolean;
      command?: RawCommand;
      error?: string;
    }>(PATHS.retry(dispatchId), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });

    if (!response.success || !response.command) {
      throw new Error(response.error ?? "Failed to retry command");
    }

    return commandFromRaw(response.command);
  },
};
