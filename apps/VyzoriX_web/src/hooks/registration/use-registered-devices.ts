import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { registration, type RegisteredDeviceListResult } from '@vyzorix/api-client';
import { fetchRegisteredDevicesViaGraphQL } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface UseRegisteredDevicesParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useRegisteredDevices(
  params?: UseRegisteredDevicesParams,
  options?: Omit<UseQueryOptions<RegisteredDeviceListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationDevices({ ...params, organizationId }),
    queryFn: async () => {
      const org = organizationId ?? undefined;
      try {
        return await registration.getDevices({ ...params, organizationId: org });
      } catch (restError) {
        if (!organizationId) throw restError;
        return fetchRegisteredDevicesViaGraphQL(organizationId, params);
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { RegisteredDeviceListResult };
