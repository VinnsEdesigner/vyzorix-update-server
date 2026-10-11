import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import {
  postDeviceInboxImeiAck,
  invalidateGetDeviceInbox,
  invalidateGetDeviceInboxImei,
} from '@/generated-rq/inbox/device-inbox';
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
    onSuccess: (_, { imei }) => {
      // Org-scoped generated keys (VYZORIX_FRONTEND_ARCHITECTURE §1) — the
      // mutation ran through the GraphQL fallback, so the generated hook's
      // automatic invalidation did not fire.
      void invalidateGetDeviceInbox(queryClient);
      void invalidateGetDeviceInboxImei(queryClient, imei);
    },
  });
}
