import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { updates, type UpdatePush } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function usePushDetail(
  pushId: string | undefined,
  options?: Omit<UseQueryOptions<UpdatePush | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updatePushDetail(pushId ?? ''),
    queryFn: () => updates.getPushDetail(pushId!, organizationId ?? undefined),
    enabled: pushId !== undefined && pushId !== '',
    ...options,
  });
}
