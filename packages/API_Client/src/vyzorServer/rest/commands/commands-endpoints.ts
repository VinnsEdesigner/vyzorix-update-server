import { restClient } from "../_shared/rest-client";
import {
  commandFromRaw,
  commandListItemFromRaw,
  paginationFromRaw,
  sendCommandRequestToRaw,
  type RawCommand,
  type RawCommandListItem,
  type RawCommandHistoryResult,
} from "@/domain/commands";
import type { Command, CommandListItem, SendCommandRequest, CommandStatus } from "@/domain/commands";

const PATHS = {
  history: (imei: string) => `/v1/device/${imei}/commands`,
  send: (imei: string) => `/v1/device/${imei}/command`,
  cancel: (imei: string, dispatchId: string) => `/v1/device/${imei}/command/${dispatchId}`,
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
      },
    });
    return {
      commands: response.commands.map(commandListItemFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async getPending(imei: string): Promise<CommandListItem[]> {
    const response = await restClient.get<{ commands: RawCommandListItem[] }>(PATHS.pending(imei));
    return response.commands.map(commandListItemFromRaw);
  },

  async getByDispatchId(dispatchId: string): Promise<Command | null> {
    const response = await restClient.get<RawCommand | null>(PATHS.status(dispatchId));
    if (!response?.id) return null;
    return commandFromRaw(response);
  },

  async send(request: SendCommandRequest): Promise<Command> {
    const payload = sendCommandRequestToRaw(request);
    const response = await restClient.post<{
      success: boolean;
      dispatchId?: string;
      command?: RawCommand;
      error?: string;
    }>(PATHS.send(request.imei), payload);

    if (!response.success || !response.command) {
      throw new Error(response.error ?? "Failed to send command");
    }

    return commandFromRaw(response.command);
  },

  async cancel(imei: string, dispatchId: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.cancel(imei, dispatchId));
  },

  async retry(dispatchId: string): Promise<Command> {
    const response = await restClient.post<{
      success: boolean;
      command?: RawCommand;
      error?: string;
    }>(PATHS.retry(dispatchId), {});

    if (!response.success || !response.command) {
      throw new Error(response.error ?? "Failed to retry command");
    }

    return commandFromRaw(response.command);
  },
};
