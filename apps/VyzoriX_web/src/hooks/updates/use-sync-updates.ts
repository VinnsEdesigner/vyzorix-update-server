import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  postUpdatesSync,
  invalidateGetUpdatesVersions,
  invalidateGetUpdatesStatus,
  invalidateGetUpdatesChangelog,
} from '@/generated-rq/updates/update-management';
import { syncUpdatesViaGraphQL } from './_graphql-fallback';

export interface SyncResult {
  status: string;
  startedAt: Date;
  versionsFound?: number;
}

export function useSyncUpdates() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (): Promise<SyncResult> => {
      try {
        const result = await postUpdatesSync();
        return {
          status: result.status ?? 'syncing',
          startedAt: result.startedAt ? new Date(result.startedAt) : new Date(),
          versionsFound: result.versionsFound,
        };
      } catch (restError) {
        if (!organizationId) throw restError;
        return syncUpdatesViaGraphQL(organizationId);
      }
    },
    onSuccess: () => {
      // GraphQL fallback path bypasses generated auto-invalidation.
      void invalidateGetUpdatesVersions(queryClient);
      void invalidateGetUpdatesStatus(queryClient);
      void invalidateGetUpdatesChangelog(queryClient);
    },
  });
}
