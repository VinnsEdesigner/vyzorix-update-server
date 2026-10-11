import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  deleteDevicesImei,
  invalidateGetDevices,
  invalidateGetDevicesImei,
} from '@/generated-rq/devices/device-management';
import type { DeregisterResult } from '@vyzorix/api-client';
import { deregisterViaGraphQL } from './_graphql-fallback';
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
        await deleteDevicesImei(imei);
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
      // Delete is called directly (not through the generated mutation hook) so
      // the GraphQL fallback can run; invalidate the generated device keys
      // explicitly.
      void invalidateGetDevices(queryClient);
      void invalidateGetDevicesImei(queryClient, imei);
    },
  });
}

export type { DeregisterResult };
