import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import { getCommands } from '@vyzorix/api-client';
import type {
  Command,
  CommandListItem,
  CommandResponse,
  CommandStatus,
  CommandParams,
  CommandHistoryResult,
  PresetCommandType,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { useCommandDispatchStore } from '@/stores';
import { fetchPendingCommandsViaGraphQL, fetchCommandViaGraphQL, normalizeWireCommand } from './_graphql-fallback';

const TERMINAL_STATUSES: CommandStatus[] = ['completed' as CommandStatus, 'failed' as CommandStatus, 'cancelled' as CommandStatus];

/** Callbacks for command status polling (previously provided by the deleted
 * hand-rolled command status poller). */
export interface CommandPollingOptions {
  onStateChange?: (prev: CommandStatus, current: CommandStatus, command: CommandResponse) => void;
  onCompleted?: (command: CommandResponse) => void;
  onFailed?: (command: CommandResponse) => void;
  onCancelled?: (command: CommandResponse) => void;
}

interface CommandHistoryParams {
  status?: string;
  page?: number;
  limit?: number;
  startTime?: number;
  endTime?: number;
}

export function useCommandHistory(
  imei: string | undefined,
  params?: CommandHistoryParams,
  options?: Omit<
    UseQueryOptions<CommandHistoryResult>,
    'queryKey' | 'queryFn'
  >,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.commands(imei ?? '', { ...params, organizationId }),
    queryFn: () =>
      getCommands().getDashboardDeviceImeiCommands(imei!, {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        startTime: params?.startTime,
        endTime: params?.endTime,
      }),
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export function useCommand(
  dispatchId: string | undefined,
  options?: Omit<UseQueryOptions<Command | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.command(dispatchId ?? ''),
    queryFn: async (): Promise<Command | null> => {
      try {
        return normalizeWireCommand(await getCommands().getCommandDispatchIdStatus(dispatchId!));
      } catch (restError) {
        if (!organizationId || !dispatchId) throw restError;
        return fetchCommandViaGraphQL(organizationId, dispatchId);
      }
    },
    enabled: dispatchId !== undefined && dispatchId !== '' && organizationId !== null,
    ...options,
  });
}

export function usePendingCommands(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.pendingCommands(imei ?? ''),
    queryFn: async (): Promise<CommandListItem[]> => {
      try {
        const result = await getCommands().getDeviceImeiCommandsPending(imei!);
        return (result.commands ?? []).map((c) => {
          const normalized = normalizeWireCommand(c);
          return {
            id: normalized.id,
            dispatchId: normalized.dispatchId,
            deviceId: normalized.deviceId,
            command: normalized.command,
            status: normalized.status,
            createdAt: normalized.createdAt,
          };
        });
      } catch (restError) {
        if (!organizationId || !imei) throw restError;
        return fetchPendingCommandsViaGraphQL(organizationId, imei);
      }
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    refetchInterval: 10_000,
  });
}

export function useSendCommand() {
  const queryClient = useQueryClient();
  const addPending = useCommandDispatchStore((s) => s.addPending);

  return useMutation({
    mutationFn: ({ imei, commandType, params }: { imei: string; commandType: PresetCommandType; params?: CommandParams }) =>
      getCommands().postDeviceImeiCommand(imei, { command: commandType, args: params }),
    onSuccess: (command, { imei, commandType, params }) => {
      addPending({
        dispatchId: command.dispatchId ?? '',
        imei,
        commandType,
        params: params ?? {},
        createdAt: Date.now(),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.commands(imei) });
      queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(imei) });
      queryClient.setQueryData(queryKeys.command(command.dispatchId ?? ''), command);
    },
  });
}

export function usePollCommandStatus() {
  const queryClient = useQueryClient();
  const removePending = useCommandDispatchStore((s) => s.removePending);

  return (dispatchId: string, options?: CommandPollingOptions) => {
    const imei = useCommandDispatchStore.getState().getPending(dispatchId)?.imei;
    return {
      onStateChange: (_prev: CommandStatus, current: CommandStatus, command: CommandResponse) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
        if (TERMINAL_STATUSES.includes(current)) {
          removePending(dispatchId);
          if (imei) {
            queryClient.invalidateQueries({ queryKey: queryKeys.commands(imei) });
            queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(imei) });
          }
        }
      },
      onCompleted: (command: CommandResponse) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
      onFailed: (command: CommandResponse) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
      onCancelled: (command: CommandResponse) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
      ...options,
    };
  };
}

export function useCancelCommand() {
  const queryClient = useQueryClient();
  const removePending = useCommandDispatchStore((s) => s.removePending);
  return useMutation({
    mutationFn: (dispatchId: string) => getCommands().deleteCommandDispatchId(dispatchId),
    onSuccess: (_, dispatchId) => {
      removePending(dispatchId);
      queryClient.invalidateQueries({ queryKey: queryKeys.command(dispatchId) });
    },
  });
}

export function useRetryCommand() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (dispatchId: string) => getCommands().postCommandDispatchIdRetry(dispatchId),
    onSuccess: (command, dispatchId) => {
      queryClient.setQueryData(queryKeys.command(dispatchId), command);
      queryClient.invalidateQueries({ queryKey: queryKeys.commands(command.dispatchId ?? '') });
    },
  });
}

export type {
  CommandResponse,
  CommandStatus,
  CommandHistoryParams,
};
