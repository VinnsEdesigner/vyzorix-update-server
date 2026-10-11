import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { getAuthMe, usePostAuthOrganizationsSelect } from '@/generated-rq/auth/auth-session';
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
        const me = await getAuthMe();
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
  const setOrganization = useAuthStore((s) => s.setOrganization);
  return usePostAuthOrganizationsSelect({
    mutation: {
      onSuccess: (org) => {
        // The generated mutation already invalidates the `getAuthMe` query tag;
        // this hook only needs its auth-store side effect.
        setOrganization((org as SelectOrganizationResult).organization_id ?? null);
      },
    },
  });
}
