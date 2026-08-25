import { useCallback } from 'react';
import { useQueryClient, useMutation } from '@tanstack/react-query';
import { getAuth } from '@vyzorix/api-client';
import type { MeResult, OrganizationInfo, SelectOrganizationResult } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

export function useMe() {
  const setFromMeResponse = useAuthStore((s) => s.setFromMeResponse);
  return { 
    data: undefined as MeResult | undefined,
    // Keep using getAuth().getAuthMe() directly since auth store integration
    // requires custom handling that generated hooks don't support.
    refetch: useCallback(async () => {
      try {
        const me = await getAuth().getAuthMe();
        setFromMeResponse(me);
        return me;
      } catch {
        return null;
      }
    }, [setFromMeResponse]),
  };
}

export function useMyOrganizations() {
  return useQueryClient().getQueryData<{ organizations: OrganizationInfo[] }>(['me', 'organizations']) ?? { organizations: [] };
}

export function useSelectOrganization() {
  const queryClient = useQueryClient();
  const setOrganization = useAuthStore((s) => s.setOrganization);
  return useMutation({
    mutationFn: (organizationId: string) =>
      getAuth().postAuthOrganizationsSelect({ organization_id: organizationId }),
    onSuccess: (org: SelectOrganizationResult) => {
      setOrganization(org.organization_id ?? null);
      queryClient.invalidateQueries({ queryKey: ['me'] });
    },
  });
}
