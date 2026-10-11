import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import {
  patchDeviceInboxImei,
  invalidateGetDeviceInbox,
  invalidateGetDeviceInboxImei,
} from '@/generated-rq/inbox/device-inbox';
import type { AckResult } from '@vyzorix/api-client';

export interface DismissInboxVariables {
  imei: string;
  data: { status: string };
}

export function useDismissInbox() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: async ({ imei, data }: DismissInboxVariables): Promise<AckResult> => {
      try {
        return normalizeAckResult(await patchDeviceInboxImei(imei, data));
      } catch (restError) {
        if (!organizationId) throw restError;
        return acknowledgeViaGraphQL(organizationId, imei, 'reject');
      }
    },
    onSuccess: (_, { imei }) => {
      // GraphQL-fallback path bypasses the generated hook's automatic
      // invalidation, so invalidate the org-scoped generated keys explicitly.
      void invalidateGetDeviceInbox(queryClient);
      void invalidateGetDeviceInboxImei(queryClient, imei);
    },
  });
}
