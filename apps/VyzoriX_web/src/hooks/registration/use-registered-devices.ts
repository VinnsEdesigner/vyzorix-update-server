import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getDevices, type RegisteredDeviceListResult } from '@vyzorix/api-client';
import { fetchRegisteredDevicesViaGraphQL, normalizeRegisteredDevice } from './_graphql-fallback';
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
    queryFn: async (): Promise<RegisteredDeviceListResult> => {
      try {
        const result = await getDevices().getDevices({ page: params?.page, limit: params?.limit });
        const limit = params?.limit ?? 20;
        const total = result.total ?? result.devices?.length ?? 0;
        return {
          devices: (result.devices ?? []).map(normalizeRegisteredDevice),
          pagination: {
            page: params?.page ?? 1,
            limit,
            total,
            totalPages: Math.ceil(total / limit),
          },
        };
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
