import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getUpdates } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
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
        const result = await getUpdates().postUpdatesSync();
        return {
          status: result.status ?? 'syncing',
          startedAt: result.startedAt ? new Date(result.startedAt) : new Date(),
          versionsFound: result.versionsFound,
        };
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return syncUpdatesViaGraphQL(organizationId);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.updateVersions() });
      queryClient.invalidateQueries({ queryKey: queryKeys.updatesStatus(organizationId ?? '') });
      queryClient.invalidateQueries({ queryKey: queryKeys.updateChangelog() });
    },
  });
}
