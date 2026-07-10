import { restClient } from "../_shared/rest-client";
import { commandFromRaw, commandListItemFromRaw, sendCommandRequestToRaw } from "@/domain/commands/commands-mappers";
import type { Command, CommandListItem, SendCommandRequest } from "@/domain/commands/commands-entity";
import type { PaginatedResult } from "@/domain/_shared";
import { offsetPaginationFromRaw } from "@/domain/_shared";

export const COMMANDS_PATHS = {
  commandHistory: (imei: string) => `/v1/device/${imei}/commands`,
  sendCommand: (imei: string) => `/v1/device/${imei}/command`,
  cancelCommand: (imei: string, dispatchId: string) => `/v1/device/${imei}/command/${dispatchId}`,
  pendingCommands: (imei: string) => `/v1/device/${imei}/commands/pending`,
  commandStatus: (dispatchId: string) => `/v1/command/${dispatchId}/status`,
  retryCommand: (dispatchId: string) => `/v1/command/${dispatchId}/retry`,
} as const;

export interface CommandHistoryParams {
  status?: "pending" | "delivered" | "failed" | "completed";
  page?: number;
  limit?: number;
  startTime?: number;
  endTime?: number;
}

interface RawCommandHistoryResponse {
  commands: Record<string, unknown>[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
    has_more: boolean;
  };
}

interface RawPendingCommandsResponse {
  commands: Record<string, unknown>[];
}

interface RawSendCommandResponse {
  success: boolean;
  dispatch_id?: string;
  command?: Record<string, unknown>;
  error?: string;
}

export async function fetchCommandHistory(
  imei: string,
  params?: CommandHistoryParams
): Promise<PaginatedResult<CommandListItem[]>> {
  const response = await restClient.get<RawCommandHistoryResponse>(
    COMMANDS_PATHS.commandHistory(imei),
    {
      status: params?.status,
      page: params?.page,
      limit: params?.limit,
      start_time: params?.startTime,
      end_time: params?.endTime,
    }
  );
  
  return {
    items: response.commands.map((raw) => commandListItemFromRaw(raw)),
    pagination: offsetPaginationFromRaw(response.pagination),
  };
}

export async function fetchPendingCommands(imei: string): Promise<CommandListItem[]> {
  const response = await restClient.get<RawPendingCommandsResponse>(
    COMMANDS_PATHS.pendingCommands(imei)
  );
  
  return response.commands.map((raw) => commandListItemFromRaw(raw));
}

export async function fetchCommandByDispatchId(dispatchId: string): Promise<Command | null> {
  const data = await restClient.get<Record<string, unknown>>(
    COMMANDS_PATHS.commandStatus(dispatchId)
  );
  if (!data || Object.keys(data).length === 0) return null;
  return commandFromRaw(data);
}

export async function sendCommand(request: SendCommandRequest): Promise<Command> {
  const payload = sendCommandRequestToRaw(request);
  const response = await restClient.post<RawSendCommandResponse>(
    COMMANDS_PATHS.sendCommand(request.imei),
    payload
  );
  
  if (!response.success) {
    throw new Error(response.error ?? "Failed to send command");
  }
  
  if (response.command) {
    return commandFromRaw(response.command);
  }
  
  return {
    id: response.dispatch_id ?? "",
    dispatchId: response.dispatch_id ?? "",
    type: request.commandType,
    deviceImei: request.imei,
    status: "pending",
    params: request.params ?? {},
    createdAt: new Date(),
  };
}

export async function cancelCommand(
  imei: string,
  dispatchId: string
): Promise<{ success: boolean }> {
  return restClient.delete<{ success: boolean }>(
    COMMANDS_PATHS.cancelCommand(imei, dispatchId)
  );
}

export async function retryCommand(dispatchId: string): Promise<Command> {
  const response = await restClient.post<RawSendCommandResponse>(
    COMMANDS_PATHS.retryCommand(dispatchId)
  );
  
  if (!response.success || !response.command) {
    throw new Error("Failed to retry command");
  }
  
  return commandFromRaw(response.command);
}
