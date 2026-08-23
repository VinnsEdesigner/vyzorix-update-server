import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getDevices, type DeregisterResult } from '@vyzorix/api-client';
import { deregisterViaGraphQL } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface DeregisterDeviceVariables {
  imei: string;
  hard?: boolean;
}

export function useDeregisterRegisteredDevice() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: async ({ imei, hard }: DeregisterDeviceVariables): Promise<DeregisterResult> => {
      try {
        await getDevices().deleteDevicesImei(imei);
        return {
          imei,
          status: 'deregistered',
          deregisteredAt: new Date(),
          retentionUntil: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
        };
      } catch (restError) {
        if (!organizationId) throw restError;
        return deregisterViaGraphQL(organizationId, imei, hard);
      }
    },
    onSuccess: (_, { imei }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationDevices() });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.removeQueries({ queryKey: queryKeys.registrationDevice(imei) });
    },
  });
}

export type { DeregisterResult };
