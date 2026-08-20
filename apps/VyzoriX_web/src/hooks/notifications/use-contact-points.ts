import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  contactPoints as api,
  type ContactPointRequest,
} from '@vyzorix/api-client';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

const contactPointKeys = {
  list: (orgId: string | null) => ['notifications', 'contact-points', orgId] as const,
  detail: (orgId: string | null, id: string) => ['notifications', 'contact-points', orgId, id] as const,
};

export function useContactPoints() {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: contactPointKeys.list(organizationId),
    queryFn: () => api.listContactPoints(organizationId ?? undefined),
    enabled: organizationId !== null,
    refetchInterval: 60_000,
  });
}

export function useCreateContactPoint() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (req: ContactPointRequest) => api.createContactPoint(req, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useUpdateContactPoint() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: ContactPointRequest }) =>
      api.updateContactPoint(id, req, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useDeleteContactPoint() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => api.deleteContactPoint(id, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useTestContactPoint() {
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => api.testContactPoint(id, organizationId ?? undefined),
  });
}
