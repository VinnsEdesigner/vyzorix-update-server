import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getDevices, type DeviceConfirmResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export interface ConfirmDeviceVariables {
  imei: string;
  commandSecret: string;
}

export function useConfirmDevice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ imei, commandSecret }: ConfirmDeviceVariables) =>
      getDevices().postDeviceConfirm({ imei, commandSecret }),
    onSuccess: (_, { imei }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInboxEntry(imei) });
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationDevice(imei) });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
}

export type { DeviceConfirmResult };
