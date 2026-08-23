import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getInbox, type InboxEntriesResult, type InboxStatus, type Pagination } from '@vyzorix/api-client';
import { fetchInboxViaGraphQL, normalizeInboxEntry } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface UseInboxParams {
  status?: InboxStatus | 'all';
  page?: number;
  limit?: number;
}

const EMPTY_PAGINATION: Pagination = { page: 1, limit: 20, total: 0, totalPages: 0 };

export function useInbox(
  params?: UseInboxParams,
  options?: Omit<UseQueryOptions<InboxEntriesResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationInbox({ ...params, organizationId }),
    queryFn: async (): Promise<InboxEntriesResult> => {
      try {
        const result = await getInbox().getDeviceInbox({
          status: params?.status === 'all' ? undefined : params?.status,
          page: params?.page,
          limit: params?.limit,
        });
        const rawPagination = result.pagination;
        return {
          requests: (result.requests ?? []).map(normalizeInboxEntry),
          pagination: rawPagination
            ? {
                page: rawPagination.page ?? 1,
                limit: rawPagination.limit ?? 20,
                total: rawPagination.total ?? 0,
                totalPages: rawPagination.total_pages ?? 0,
              }
            : EMPTY_PAGINATION,
        };
      } catch (restError) {
        if (!organizationId) throw restError;
        return fetchInboxViaGraphQL(organizationId, params);
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { InboxEntriesResult, InboxStatus };
