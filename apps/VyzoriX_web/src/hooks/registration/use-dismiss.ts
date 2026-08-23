import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getInbox, type AckResult } from '@vyzorix/api-client';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useDismissInbox() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();

  return useMutation({
    mutationFn: async (imei: string): Promise<AckResult> => {
      // Dismiss = reject the inbox entry (the old DELETE /v1/inbox/:imei route
      // no longer exists on the server).
      try {
        return normalizeAckResult(await getInbox().postDeviceInboxImeiAck(imei, { action: 'reject' }));
      } catch (restError) {
        if (!organizationId) throw restError;
        return acknowledgeViaGraphQL(organizationId, imei, 'reject');
      }
    },
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.removeQueries({ queryKey: queryKeys.registrationInboxEntry(imei) });
    },
  });
}
