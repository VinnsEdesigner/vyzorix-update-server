import { usePostDeviceInbox } from '@/generated-rq/inbox/device-inbox';

export function useRegisterDevice() {
  return usePostDeviceInbox();
}
