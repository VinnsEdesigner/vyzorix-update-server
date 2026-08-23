import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import { getSettings,
  type OperatorThresholds,
  type NotificationSettings,
  type ClientSettings,
  type SettingsResponseResult,
  type ThresholdsResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useSettings(
  options?: Omit<UseQueryOptions<SettingsResponseResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.settings,
    queryFn: () => getSettings().getMeSettings(),
    ...options,
  });
}

export function useThresholds(
  options?: Omit<UseQueryOptions<ThresholdsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.thresholds,
    queryFn: () => getSettings().getMeThresholds(),
    ...options,
  });
}

export function useNotifications(
  options?: Omit<UseQueryOptions<NotificationSettings>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.notifications,
    queryFn: () => getSettings().getMeNotifications(),
    ...options,
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { client?: Partial<ClientSettings> }) => getSettings().patchMeSettings(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.settings, updated);
    },
  });
}

export function useUpdateThresholds() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<OperatorThresholds>) => getSettings().patchMeThresholds(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.thresholds, updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.settings });
    },
  });
}

export function useUpdateNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<NotificationSettings>) => getSettings().patchMeNotifications(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.notifications, updated);
    },
  });
}

export function useTestWebhook() {
  return useMutation({
    mutationFn: (url: string) => getSettings().postMeNotificationsWebhookTest({ url }),
  });
}

export function useRotateWebhookSecret() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => getSettings().postMeNotificationsWebhookRotate(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
    },
  });
}

export type { OperatorThresholds, NotificationSettings, ClientSettings };
