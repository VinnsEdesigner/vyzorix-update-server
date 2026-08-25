import { useQueryClient } from '@tanstack/react-query';
import { usePostDeviceInbox } from '@/generated-rq/inbox/device-inbox';

export function useRegisterDevice() {
  const queryClient = useQueryClient();
  return usePostDeviceInbox({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['inbox'] }) },
  });
}
