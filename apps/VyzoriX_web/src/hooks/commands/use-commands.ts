import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  commands,
  pollCommandStatus,
  type Command,
  type CommandListItem,
  type CommandStatus,
  type SendCommandRequest,
  type CommandHistoryParams,
  type CommandPollingOptions,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { useCommandDispatchStore } from '@/stores';
import { fetchPendingCommandsViaGraphQL, fetchCommandViaGraphQL } from './_graphql-fallback';

const TERMINAL_STATUSES: CommandStatus[] = ['completed', 'failed', 'cancelled'];

export function useCommandHistory(
  imei: string | undefined,
  params?: Omit<CommandHistoryParams, 'organizationId'>,
  options?: Omit<
    UseQueryOptions<Awaited<ReturnType<typeof commands.getHistory>>>,
    'queryKey' | 'queryFn'
  >,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.commands(imei ?? '', { ...params, organizationId }),
    queryFn: () =>
      commands.getHistory(imei!, {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        startTime: params?.startTime,
        endTime: params?.endTime,
        organizationId: organizationId ?? undefined,
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
    queryFn: async () => {
      try {
        return await commands.getByDispatchId(dispatchId!, organizationId ?? undefined);
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
    queryFn: async () => {
      try {
        return await commands.getPending(imei!, organizationId ?? undefined);
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
  const organizationId = useCurrentOrganizationId();
  const addPending = useCommandDispatchStore((s) => s.addPending);

  return useMutation({
    mutationFn: (request: SendCommandRequest) =>
      commands.send(request, organizationId ?? undefined),
    onSuccess: (command, request) => {
      addPending({
        dispatchId: command.dispatchId,
        imei: command.deviceId,
        commandType: request.commandType,
        params: request.params,
        createdAt: Date.now(),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.commands(command.deviceId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(command.deviceId) });
      queryClient.setQueryData(queryKeys.command(command.dispatchId), command);
    },
  });
}

export function usePollCommandStatus() {
  const queryClient = useQueryClient();
  const removePending = useCommandDispatchStore((s) => s.removePending);

  return (dispatchId: string, options?: CommandPollingOptions) => {
    const imei = useCommandDispatchStore.getState().getPending(dispatchId)?.imei;
    return pollCommandStatus(dispatchId, options, {
      onStateChange: (_prev, current, command) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
        if (TERMINAL_STATUSES.includes(current)) {
          removePending(dispatchId);
          if (imei) {
            queryClient.invalidateQueries({ queryKey: queryKeys.commands(imei) });
            queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(imei) });
          }
        }
      },
      onCompleted: (command) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
      onFailed: (command) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
      onCancelled: (command) => {
        queryClient.setQueryData(queryKeys.command(dispatchId), command);
      },
    });
  };
}

export function useCancelCommand() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  const removePending = useCommandDispatchStore((s) => s.removePending);
  return useMutation({
    mutationFn: (dispatchId: string) => commands.cancel(dispatchId, organizationId ?? undefined),
    onSuccess: (_, dispatchId) => {
      removePending(dispatchId);
      queryClient.invalidateQueries({ queryKey: queryKeys.command(dispatchId) });
    },
  });
}

export function useRetryCommand() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (dispatchId: string) => commands.retry(dispatchId, organizationId ?? undefined),
    onSuccess: (command, dispatchId) => {
      queryClient.setQueryData(queryKeys.command(dispatchId), command);
      queryClient.invalidateQueries({ queryKey: queryKeys.commands(command.deviceId) });
    },
  });
}

export type {
  Command,
  CommandListItem,
  CommandStatus,
  SendCommandRequest,
  CommandHistoryParams,
};
