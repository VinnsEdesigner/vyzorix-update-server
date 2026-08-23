import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getInbox } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useResendInboxApproval() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (imei: string) =>
      getInbox().postDeviceInboxImeiResend(imei),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
    },
  });
}
