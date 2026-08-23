import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getInbox, type AckResult, type AcknowledgeAction } from '@vyzorix/api-client';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface AcknowledgeInboxVariables {
  imei: string;
  action: AcknowledgeAction;
  notes?: string;
}

export function useAcknowledgeInbox() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: async ({ imei, action, notes }: AcknowledgeInboxVariables): Promise<AckResult> => {
      try {
        return normalizeAckResult(await getInbox().postDeviceInboxImeiAck(imei, { action, notes }));
      } catch (restError) {
        if (!organizationId) throw restError;
        return acknowledgeViaGraphQL(organizationId, imei, action, notes);
      }
    },
    onSuccess: (_, { imei }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInboxEntry(imei) });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
}

export type { AckResult, AcknowledgeAction };
