import { usePostDeviceInbox } from '@/generated-rq/inbox/device-inbox';

export function useCreateInboxRequest() {
  return usePostDeviceInbox();
}
