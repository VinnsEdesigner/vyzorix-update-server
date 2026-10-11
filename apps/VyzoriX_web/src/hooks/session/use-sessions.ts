import {
  useGetAuthSessions,
  useGetAuthSessionsConcurrent,
  useDeleteAuthSessionsId,
  useDeleteAuthSessions,
  usePostAuthSessionsRevokeAll,
} from '@/generated-rq/sessions/auth-sessions';

export function useSessions() {
  return useGetAuthSessions();
}

export function useConcurrentSessions() {
  return useGetAuthSessionsConcurrent();
}

export function useRevokeSession() {
  return useDeleteAuthSessionsId();
}

export function useRevokeAllSessions() {
  return useDeleteAuthSessions();
}

export function useRevokeAllDevices() {
  return usePostAuthSessionsRevokeAll();
}
