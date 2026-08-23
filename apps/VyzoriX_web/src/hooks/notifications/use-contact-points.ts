import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getContactPoints,
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
    queryFn: () => getContactPoints().getNotificationsContactPoints(),
    enabled: organizationId !== null,
    refetchInterval: 60_000,
  });
}

export function useCreateContactPoint() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: ContactPointRequest) => getContactPoints().postNotificationsContactPoints(req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useUpdateContactPoint() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: ContactPointRequest }) =>
      getContactPoints().patchNotificationsContactPointsId(id, req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useDeleteContactPoint() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => getContactPoints().deleteNotificationsContactPointsId(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useTestContactPoint() {
  return useMutation({
    mutationFn: (id: string) => getContactPoints().postNotificationsContactPointsIdTest(id),
  });
}
