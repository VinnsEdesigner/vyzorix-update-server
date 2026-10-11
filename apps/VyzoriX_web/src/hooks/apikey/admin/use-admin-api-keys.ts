import {
  useGetAdminApiKeys,
  useGetAdminApiKeysOperatorOperatorId,
  useGetAdminApiKeysStats,
  useGetAdminApiKeysStatsOperatorOperatorId,
  useDeleteAdminApiKeysKeyId,
} from '@/generated-rq/admin/admin-operations';

export function useAdminApiKeys(params?: { page?: number; limit?: number; operatorId?: string; search?: string }) {
  return useGetAdminApiKeys({
    page: params?.page,
    limit: params?.limit,
    operator_id: params?.operatorId,
    search: params?.search,
  });
}

export function useAdminOperatorKeys(operatorId: string | undefined) {
  return useGetAdminApiKeysOperatorOperatorId(
    operatorId ?? '',
    undefined,
    { query: { enabled: operatorId !== undefined } },
  );
}

export function useGlobalStats() {
  return useGetAdminApiKeysStats();
}

export function useAdminOperatorKeyStats(operatorId: string | undefined) {
  return useGetAdminApiKeysStatsOperatorOperatorId(
    operatorId ?? '',
    { query: { enabled: operatorId !== undefined } },
  );
}

export function useForceRevokeKey() {
  return useDeleteAdminApiKeysKeyId();
}
