import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import { getSessions,
  type SessionListResult,
  type ConcurrentSessionsResult,
  type SuccessResult,
  type RevokeResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useSessions(
  options?: Omit<UseQueryOptions<SessionListResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.sessions,
    queryFn: () => getSessions().getAuthSessions(),
    ...options,
  });
}

export function useConcurrentSessions(
  options?: Omit<UseQueryOptions<ConcurrentSessionsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.concurrentSessions,
    queryFn: () => getSessions().getAuthSessionsConcurrent(),
    ...options,
  });
}

export function useRevokeSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) => getSessions().deleteAuthSessionsId(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      queryClient.invalidateQueries({ queryKey: queryKeys.concurrentSessions });
    },
  });
}

export function useRevokeAllSessions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => getSessions().deleteAuthSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
    },
  });
}

export function useRevokeAllDevices() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => getSessions().postAuthSessionsRevokeAll(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      queryClient.invalidateQueries({ queryKey: queryKeys.concurrentSessions });
    },
  });
}

export type { SessionListResult, ConcurrentSessionsResult, SuccessResult, RevokeResult };
