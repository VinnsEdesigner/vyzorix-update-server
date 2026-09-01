import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { postUpdatesPush } from '@/generated-rq/updates/update-management';
import type { UpdatePushRequest, InstallType } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useUpdatesStore } from '@/stores/updates-store';
import { pushUpdateViaGraphQL, normalizeWirePushResult } from './_graphql-fallback';

export function usePushUpdate() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (request: UpdatePushRequest) => {
      try {
        return normalizeWirePushResult(await postUpdatesPush(request));
      } catch (restError) {
        if (!organizationId) throw restError;
        if (!request.version || !request.deviceIds || !request.installType) {
          throw new Error('version, deviceIds, and installType are required');
        }
        return pushUpdateViaGraphQL(organizationId, {
          version: request.version,
          deviceIds: request.deviceIds,
          installType: request.installType as InstallType,
          scheduledAt: request.scheduledAt ? new Date(request.scheduledAt) : undefined,
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
