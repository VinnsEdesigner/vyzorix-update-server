import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import { postDeviceInboxImeiAck } from '@/generated-rq/inbox/device-inbox';
import type { AckResult, AcknowledgeAction } from '@vyzorix/api-client';

export interface AcknowledgeInboxVariables {
  imei: string;
  data: { action: AcknowledgeAction; notes?: string };
}

export function useAcknowledgeInbox() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: async ({ imei, data }: AcknowledgeInboxVariables): Promise<AckResult> => {
      try {
        return normalizeAckResult(await postDeviceInboxImeiAck(imei, { action: data.action, notes: data.notes }));
      } catch (restError) {
        if (!organizationId) throw restError;
        return acknowledgeViaGraphQL(organizationId, imei, data.action, data.notes);
      }
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['inbox'] }),
  });
}
