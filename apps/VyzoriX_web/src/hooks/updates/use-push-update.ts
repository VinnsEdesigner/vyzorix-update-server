import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  updates,
  type UpdatePush,
  type PushUpdateRequest,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { useUpdatesStore } from '@/stores/updates-store';
import { pushUpdateViaGraphQL } from './_graphql-fallback';

export function usePushUpdate() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (request: PushUpdateRequest): Promise<UpdatePush> => {
      try {
        return await updates.pushUpdate(request, organizationId ?? undefined);
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return pushUpdateViaGraphQL(organizationId, {
          version: request.version,
          deviceIds: request.deviceIds,
          installType: request.installType,
          scheduledAt: request.scheduledAt,
        });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['updates', 'history'] });
      queryClient.invalidateQueries({ queryKey: queryKeys.updatesStatus(organizationId ?? '') });
      useUpdatesStore.getState().resetDraft();
    },
  });
}
