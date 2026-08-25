import { useQueryClient } from '@tanstack/react-query';
import {
  useGetMeSettings,
  usePatchMeSettings,
  useGetMeThresholds,
  usePatchMeThresholds,
  useGetMeNotifications,
  usePatchMeNotifications,
  usePostMeNotificationsWebhookTest,
  usePostMeNotificationsWebhookRotate,
} from '@/generated-rq/settings/operator-settings';

export function useSettings() {
  return useGetMeSettings({ query: { queryKey: ['me-settings'] as const } });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return usePatchMeSettings({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me-settings'] }) },
  });
}

export function useThresholds() {
  return useGetMeThresholds({ query: { queryKey: ['me-thresholds'] as const } });
}

export function useUpdateThresholds() {
  const queryClient = useQueryClient();
  return usePatchMeThresholds({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me-thresholds'] }) },
  });
}

export function useNotifications() {
  return useGetMeNotifications({ query: { queryKey: ['me-notifications'] as const } });
}

export function useUpdateNotifications() {
  const queryClient = useQueryClient();
  return usePatchMeNotifications({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me-notifications'] }) },
  });
}

export function useTestWebhook() {
  return usePostMeNotificationsWebhookTest();
}

export function useRotateWebhookSecret() {
  const queryClient = useQueryClient();
  return usePostMeNotificationsWebhookRotate({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me-notifications'] }) },
  });
}
