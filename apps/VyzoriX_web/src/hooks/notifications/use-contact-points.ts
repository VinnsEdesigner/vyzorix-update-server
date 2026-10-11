import {
  useGetNotificationsContactPoints,
  usePostNotificationsContactPoints,
  usePatchNotificationsContactPointsId,
  useDeleteNotificationsContactPointsId,
  usePostNotificationsContactPointsIdTest,
} from '@/generated-rq/contact-points/notification-contact-points';

export function useContactPoints() {
  return useGetNotificationsContactPoints({ query: { refetchInterval: 60_000 } });
}

export function useCreateContactPoint() {
  return usePostNotificationsContactPoints();
}

export function useUpdateContactPoint() {
  return usePatchNotificationsContactPointsId();
}

export function useDeleteContactPoint() {
  return useDeleteNotificationsContactPointsId();
}

export function useTestContactPoint() {
  return usePostNotificationsContactPointsIdTest();
}
