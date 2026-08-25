import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchInboxViaGraphQL } from './_graphql-fallback';
import {
  useGetDeviceInbox,
} from '@/generated-rq/inbox/device-inbox';

import type { InboxStatus } from '@vyzorix/api-client';

export function useInbox(params?: { status?: InboxStatus | 'all'; page?: number; limit?: number }) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceInbox(
    { status: params?.status, page: params?.page, limit: params?.limit },
    {
      query: {
        queryKey: ['inbox', params, organizationId] as const,
        enabled: organizationId !== null,
        // REST primary; fall back to GraphQL on REST failure.
        queryFn: async () => {
          try {
            return await import('@/generated-rq/inbox/device-inbox').then((m) =>
              m.getDeviceInbox({ status: params?.status, page: params?.page, limit: params?.limit }),
            );
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchInboxViaGraphQL(organizationId, params) as unknown as Awaited<ReturnType<typeof import('@/generated-rq/inbox/device-inbox').getDeviceInbox>>;
          }
        },
      },
    },
  );
}





