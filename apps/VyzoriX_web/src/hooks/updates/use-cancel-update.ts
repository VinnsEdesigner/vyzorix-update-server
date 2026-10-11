import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { cancelUpdateViaGraphQL, normalizeWireCancelResult } from './_graphql-fallback';
import {
  postUpdatesHistoryPushIdCancel,
  invalidateGetUpdatesHistory,
  invalidateGetUpdatesHistoryPushId,
} from '@/generated-rq/updates/update-management';
import type { UpdatePush } from '@vyzorix/api-client';

export function useCancelUpdate() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: async ({ pushId }: { pushId: string }): Promise<UpdatePush> => {
      try {
        return normalizeWireCancelResult(await postUpdatesHistoryPushIdCancel(pushId));
      } catch (restError) {
        if (!organizationId) throw restError;
        return cancelUpdateViaGraphQL(organizationId, pushId);
      }
    },
    onSuccess: (_, { pushId }) => {
      // GraphQL fallback path bypasses generated auto-invalidation.
      void invalidateGetUpdatesHistory(queryClient);
      void invalidateGetUpdatesHistoryPushId(queryClient, pushId);
    },
  });
}
