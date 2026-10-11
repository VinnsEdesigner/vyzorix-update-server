import { usePostDeviceInboxImeiResend } from '@/generated-rq/inbox/device-inbox';

export function useResendInboxApproval() {
  return usePostDeviceInboxImeiResend();
}
