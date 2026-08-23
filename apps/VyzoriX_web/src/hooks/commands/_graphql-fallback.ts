import {
  queryPendingCommands,
  queryCommand,
  type Command,
  type CommandListItem,
  type CommandResponse,
} from '@vyzorix/api-client';

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

// queryPendingCommands / queryCommand return the raw Apollo QueryResult. The
// GraphQL payload lives under its `data` field; unwrap it before extracting.
function unwrapApolloData(response: unknown): Record<string, unknown> | null {
  const r = response as { data?: Record<string, unknown> } | null;
  return r?.data ?? null;
}

function extractArray<T>(response: unknown, key: string): T[] {
  const r = unwrapApolloData(response);
  if (!r) return [];
  const value = r[key];
  return Array.isArray(value) ? (value as T[]) : [];
}

function extractObject<T>(response: unknown, key: string): T | null {
  const r = unwrapApolloData(response);
  if (!r) return null;
  const value = r[key];
  return (value as T | undefined) ?? null;
}

export async function fetchPendingCommandsViaGraphQL(
  organizationId: string,
  imei: string,
): Promise<CommandListItem[]> {
  const response = await queryPendingCommands({ organizationId, deviceId: imei });
  const raw = extractArray<RawCommandFields>(response, 'pendingCommands');
  return raw.map(normalizeCommandListItem);
}

export async function fetchCommandViaGraphQL(
  organizationId: string,
  dispatchId: string,
): Promise<Command | null> {
  const response = await queryCommand({ organizationId, dispatchId });
  const raw = extractObject<RawCommandFields>(response, 'command');
  if (!raw?.dispatchId) return null;
  return normalizeCommand(raw);
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
