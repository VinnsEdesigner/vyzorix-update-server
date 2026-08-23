import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getUpdates, type UpdatePushDetailResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function usePushDetail(
  pushId: string | undefined,
  options?: Omit<UseQueryOptions<UpdatePushDetailResult | null>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.updatePushDetail(pushId ?? ''),
    queryFn: () => getUpdates().getUpdatesHistoryPushId(pushId!),
    enabled: pushId !== undefined && pushId !== '',
    ...options,
  });
}

export type { UpdatePushDetailResult };
