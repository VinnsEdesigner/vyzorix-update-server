import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { acknowledgeViaGraphQL, normalizeAckResult } from './_graphql-fallback';
import { patchDeviceInboxImei } from '@/generated-rq/inbox/device-inbox';
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
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['inbox'] }),
  });
}
