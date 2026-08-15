import { useMutation, useQueryClient } from '@tanstack/react-query';
import { registration, type InboxStatus } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useDismissInbox() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (imei: string) =>
      registration.dismissInbox(imei, organizationId ?? undefined),
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.removeQueries({ queryKey: queryKeys.registrationInboxEntry(imei) });
    },
  });
}

export type { InboxStatus };
