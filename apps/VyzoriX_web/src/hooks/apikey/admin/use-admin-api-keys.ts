import { useQueryClient } from '@tanstack/react-query';
import {
  useGetAdminApiKeys,
  useGetAdminApiKeysOperatorOperatorId,
  useGetAdminApiKeysStats,
  useGetAdminApiKeysStatsOperatorOperatorId,
  useDeleteAdminApiKeysKeyId,
} from '@/generated-rq/admin/admin-operations';

export function useAdminApiKeys(params?: { page?: number; limit?: number; operatorId?: string; search?: string }) {
  return useGetAdminApiKeys(
    { page: params?.page, limit: params?.limit, operator_id: params?.operatorId, search: params?.search },
    { query: { queryKey: ['admin', 'api-keys', params] as const } },
  );
}

export function useAdminOperatorKeys(operatorId: string | undefined) {
  return useGetAdminApiKeysOperatorOperatorId(
    operatorId ?? '',
    undefined,
    { query: { queryKey: ['admin', 'api-keys', 'operator', operatorId] as const, enabled: operatorId !== undefined } },
  );
}

export function useGlobalStats() {
  return useGetAdminApiKeysStats(
    { query: { queryKey: ['admin', 'api-keys', 'stats'] as const } },
  );
}

export function useAdminOperatorKeyStats(operatorId: string | undefined) {
  return useGetAdminApiKeysStatsOperatorOperatorId(
    operatorId ?? '',
    { query: { queryKey: ['admin', 'api-keys', 'stats', operatorId] as const, enabled: operatorId !== undefined } },
  );
}

export function useForceRevokeKey() {
  const queryClient = useQueryClient();
  return useDeleteAdminApiKeysKeyId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'api-keys'] }) },
  });
}
