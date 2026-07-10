import type { Command, CommandListItem, CommandStatus, PresetCommandType } from "@/domain/commands";

export type RawCommand = {
  __typename?: "Command";
} & Omit<Command, "createdAt" | "sentAt" | "acknowledgedAt" | "completedAt"> & {
  createdAt: string;
  sentAt?: string | null;
  acknowledgedAt?: string | null;
  completedAt?: string | null;
};

export type RawCommandListItem = {
  __typename?: "Command";
} & Omit<CommandListItem, "createdAt"> & {
  createdAt: string;
};

export interface RawCommandConnection {
  commands: RawCommandListItem[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

export interface RawCommandResult {
  success: boolean;
  message?: string;
  output?: string;
  error?: string;
}

export interface RawCommandStatus {
  dispatchId: string;
  status: CommandStatus;
  updatedAt: string;
  result?: RawCommandResult;
}

export interface RawSendCommandResponse {
  success: boolean;
  dispatchId?: string;
  status?: CommandStatus;
  error?: string;
}

export interface RawCancelCommandResponse {
  success: boolean;
  dispatchId?: string;
  error?: string;
}
