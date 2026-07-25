export type RawCommand = {
  __typename?: "Command";
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  args?: Record<string, unknown>;
  status: string;
  createdAt?: string;
  deliveredAt?: string;
};

export type RawCommandListItem = {
  __typename?: "Command";
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  status: string;
  createdAt?: string;
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
  dispatchId: string;
  commandId: string;
  status: string;
  deviceOnline: boolean;
}

export interface RawCommandStatus {
  dispatchId: string;
  status: string;
  updatedAt?: string;
  result?: {
    dispatchId: string;
    commandId: string;
    status: string;
    deviceOnline: boolean;
  };
}

export interface RawSendCommandResponse {
  dispatchId: string;
  commandId: string;
  status: string;
  deviceOnline: boolean;
}

export interface RawCancelCommandResponse {
  dispatchId: string;
  cancelledAt: number;
  status: string;
}
