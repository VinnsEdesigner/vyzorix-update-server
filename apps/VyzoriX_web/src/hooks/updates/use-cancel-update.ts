import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updates, type UpdatePush } from '@vyzorix/api-client';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { cancelUpdateViaGraphQL } from './_graphql-fallback';

export function useCancelUpdate() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (pushId: string): Promise<UpdatePush> => {
      try {
        return await updates.cancelPush(pushId, organizationId ?? undefined);
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
