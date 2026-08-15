import { useMutation, useQueryClient } from '@tanstack/react-query';
import { registration } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useResendInboxApproval() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (imei: string) =>
      registration.resendInboxApproval(imei, organizationId ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
    },
  });
}
