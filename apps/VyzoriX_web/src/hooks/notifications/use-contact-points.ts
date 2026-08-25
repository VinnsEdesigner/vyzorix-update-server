import { useQueryClient } from '@tanstack/react-query';
import {
  useGetNotificationsContactPoints,
  usePostNotificationsContactPoints,
  usePatchNotificationsContactPointsId,
  useDeleteNotificationsContactPointsId,
  usePostNotificationsContactPointsIdTest,
} from '@/generated-rq/contact-points/notification-contact-points';

export function useContactPoints() {
  return useGetNotificationsContactPoints({ query: { queryKey: ['contact-points'] as const, refetchInterval: 60_000 } });
}

export function useCreateContactPoint() {
  const queryClient = useQueryClient();
  return usePostNotificationsContactPoints({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['contact-points'] }) },
  });
}

export function useUpdateContactPoint() {
  const queryClient = useQueryClient();
  return usePatchNotificationsContactPointsId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['contact-points'] }) },
  });
}

export function useDeleteContactPoint() {
  const queryClient = useQueryClient();
  return useDeleteNotificationsContactPointsId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['contact-points'] }) },
  });
}

export function useTestContactPoint() {
  return usePostNotificationsContactPointsIdTest();
}
