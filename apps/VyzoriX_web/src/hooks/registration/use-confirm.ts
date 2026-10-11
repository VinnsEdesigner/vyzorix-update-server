import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  postDeviceConfirm,
  invalidateGetDevices,
  invalidateGetDashboardDevices,
  invalidateGetDashboardDevicesOperator,
  invalidateGetDashboardEventsRecent,
  invalidateGetDeviceCount,
} from '@/generated-rq/devices/device-management';
import type { DeviceConfirmResult } from '@vyzorix/api-client';

export interface ConfirmDeviceVariables {
  imei: string;
  commandSecret: string;
}

export function useConfirmDevice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ imei, commandSecret }: ConfirmDeviceVariables) =>
      postDeviceConfirm({ imei, commandSecret }),
    onSuccess: () => {
      // Confirm carries no GraphQL fallback, but it is wired by hand because the
      // variables shape (`{ imei, commandSecret }`) predates the generated hook.
      // Invalidate the same org-scoped keys the generated `postDeviceConfirm`
      // mutation tag declares so confirm and
      // the generated surface stay in lockstep.
      void invalidateGetDevices(queryClient);
      void invalidateGetDashboardDevices(queryClient);
      void invalidateGetDashboardDevicesOperator(queryClient);
      void invalidateGetDashboardEventsRecent(queryClient);
      void invalidateGetDeviceCount(queryClient);
    },
  });
}

export type { DeviceConfirmResult };
