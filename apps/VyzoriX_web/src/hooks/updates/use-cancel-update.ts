import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getUpdates, type UpdatePush } from '@vyzorix/api-client';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { cancelUpdateViaGraphQL, normalizeWireCancelResult } from './_graphql-fallback';

export function useCancelUpdate() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (pushId: string): Promise<UpdatePush> => {
      try {
        return normalizeWireCancelResult(await getUpdates().postUpdatesHistoryPushIdCancel(pushId));
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return cancelUpdateViaGraphQL(organizationId, pushId);
      }
    },
    onSuccess: (push) => {
      queryClient.invalidateQueries({ queryKey: ['updates', 'history'] });
      queryClient.setQueryData(['updates', 'push', push.id], push);
    },
  });
}
