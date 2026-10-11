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
  return useGetMeSettings();
}

export function useUpdateSettings() {
  return usePatchMeSettings();
}

export function useThresholds() {
  return useGetMeThresholds();
}

export function useUpdateThresholds() {
  return usePatchMeThresholds();
}

export function useNotifications() {
  return useGetMeNotifications();
}

export function useUpdateNotifications() {
  return usePatchMeNotifications();
}

export function useTestWebhook() {
  return usePostMeNotificationsWebhookTest();
}

export function useRotateWebhookSecret() {
  return usePostMeNotificationsWebhookRotate();
}
