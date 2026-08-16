import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updates } from '@vyzorix/api-client';
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
        return await updates.sync(organizationId ?? undefined);
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
