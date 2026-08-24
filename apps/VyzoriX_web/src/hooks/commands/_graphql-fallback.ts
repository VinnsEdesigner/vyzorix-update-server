import {
  gqlQuery,
  type Command,
  type CommandListItem,
  type CommandResponse,
} from '@vyzorix/api-client';
import {
  GetPendingCommandsDocument,
  GetCommandDocument,
  type GetPendingCommandsQuery,
  type GetCommandQuery,
} from '@vyzorix/api-client/generated-graphql';

interface RawCommandFields {
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  status: string;
  createdAt?: string;
  deliveredAt?: string;
}

function toDate(value?: string | null): Date {
  if (!value) return new Date();
  return new Date(value);
}

function normalizeCommand(raw: RawCommandFields): Command {
  return {
    id: raw.commandId,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: raw.status as Command['status'],
    failureReason: '',
    args: {},
    createdAt: toDate(raw.createdAt),
    updatedAt: toDate(raw.createdAt),
    deliveredAt: raw.deliveredAt ? toDate(raw.deliveredAt) : undefined,
    completedAt: undefined,
  };
}

function normalizeCommandListItem(raw: RawCommandFields): CommandListItem {
  return {
    id: raw.commandId,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: raw.status as CommandListItem['status'],
    createdAt: toDate(raw.createdAt),
  };
}




export async function fetchPendingCommandsViaGraphQL(
  organizationId: string,
  imei: string,
): Promise<CommandListItem[]> {
  const data = await gqlQuery<GetPendingCommandsQuery>(GetPendingCommandsDocument, { organizationId, deviceId: imei });
  const raw = (data.pendingCommands ?? []).filter((c): c is NonNullable<typeof c> => c != null);
  return raw.map((c) => normalizeCommandListItem(c as RawCommandFields));
}

export async function fetchCommandViaGraphQL(
  organizationId: string,
  dispatchId: string,
): Promise<Command | null> {
  const data = await gqlQuery<GetCommandQuery>(GetCommandDocument, { organizationId, dispatchId });
  const raw = data.command;
  if (!raw?.dispatchId) return null;
  return normalizeCommand(raw as RawCommandFields);
}

/** Maps the REST wire DTO onto the domain Command shape so hook consumers get
 * one type regardless of the data path. */
export function normalizeWireCommand(raw: CommandResponse): Command {
  const createdAt = toDate(raw.serverTime ? String(raw.serverTime) : undefined);
  return {
    id: raw.id ?? raw.command ?? '',
    dispatchId: raw.dispatchId ?? '',
    deviceId: raw.deviceId ?? '',
    command: raw.command ?? '',
    status: (raw.status ?? 'pending') as Command['status'],
    failureReason: '',
    args: {},
    createdAt,
    updatedAt: createdAt,
  };
}
