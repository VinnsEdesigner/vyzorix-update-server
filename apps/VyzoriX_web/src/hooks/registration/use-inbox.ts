import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { registration, type InboxListResult, type InboxStatus } from '@vyzorix/api-client';
import { fetchInboxViaGraphQL } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface UseInboxParams {
  status?: InboxStatus | 'all';
  page?: number;
  limit?: number;
}

export function useInbox(
  params?: UseInboxParams,
  options?: Omit<UseQueryOptions<InboxListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationInbox({ ...params, organizationId }),
    queryFn: async () => {
      const org = organizationId ?? undefined;
      try {
        return await registration.getInbox({ ...params, organizationId: org });
      } catch (restError) {
        if (!organizationId) throw restError;
        return fetchInboxViaGraphQL(organizationId, params);
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { InboxListResult, InboxStatus };
