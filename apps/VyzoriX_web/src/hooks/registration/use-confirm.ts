import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  registration,
  type ConfirmDeviceResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface ConfirmDeviceVariables {
  imei: string;
  commandSecret: string;
}

export function useConfirmDevice() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: ({ imei, commandSecret }: ConfirmDeviceVariables) =>
      registration.confirmDevice(imei, commandSecret, organizationId ?? undefined),
    onSuccess: (_, { imei }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInboxEntry(imei) });
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationDevice(imei) });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
}

export type { ConfirmDeviceResult };
