import { useQueryClient } from '@tanstack/react-query';
import {
  useGetAuthSessions,
  useGetAuthSessionsConcurrent,
  useDeleteAuthSessionsId,
  useDeleteAuthSessions,
  usePostAuthSessionsRevokeAll,
} from '@/generated-rq/sessions/auth-sessions';

export function useSessions() {
  return useGetAuthSessions({ query: { queryKey: ['sessions'] as const } });
}

export function useConcurrentSessions() {
  return useGetAuthSessionsConcurrent({ query: { queryKey: ['concurrent-sessions'] as const } });
}

export function useRevokeSession() {
  const queryClient = useQueryClient();
  return useDeleteAuthSessionsId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['sessions'] });
        queryClient.invalidateQueries({ queryKey: ['concurrent-sessions'] });
      },
    },
  });
}

export function useRevokeAllSessions() {
  const queryClient = useQueryClient();
  return useDeleteAuthSessions({
    mutation: {
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['sessions'] }),
    },
  });
}

export function useRevokeAllDevices() {
  const queryClient = useQueryClient();
  return usePostAuthSessionsRevokeAll({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['sessions'] });
        queryClient.invalidateQueries({ queryKey: ['concurrent-sessions'] });
      },
    },
  });
}
