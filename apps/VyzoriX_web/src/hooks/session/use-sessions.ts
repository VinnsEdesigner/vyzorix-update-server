import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import {
  sessions,
  type SessionListResponse,
  type ConcurrentSessionsResponse,
  type RevokeSessionResponse,
  type RevokeAllSessionsResponse,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useSessions(
  options?: Omit<UseQueryOptions<SessionListResponse>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.sessions,
    queryFn: () => sessions.listSessions(),
    ...options,
  });
}

export function useConcurrentSessions(
  options?: Omit<UseQueryOptions<ConcurrentSessionsResponse>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.concurrentSessions,
    queryFn: () => sessions.getConcurrent(),
    ...options,
  });
}

export function useRevokeSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) => sessions.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      queryClient.invalidateQueries({ queryKey: queryKeys.concurrentSessions });
    },
  });
}

export function useRevokeAllSessions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => sessions.revokeAllExceptCurrent(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
    },
  });
}

export function useRevokeAllDevices() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => sessions.revokeAllDevices(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      queryClient.invalidateQueries({ queryKey: queryKeys.concurrentSessions });
    },
  });
}

export type { SessionListResponse, ConcurrentSessionsResponse, RevokeSessionResponse, RevokeAllSessionsResponse };
