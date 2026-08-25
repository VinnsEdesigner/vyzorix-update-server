import { useQueryClient } from '@tanstack/react-query';
import { usePostDeviceInboxImeiResend } from '@/generated-rq/inbox/device-inbox';

export function useResendInboxApproval() {
  const queryClient = useQueryClient();
  return usePostDeviceInboxImeiResend({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['inbox'] }) },
  });
}
