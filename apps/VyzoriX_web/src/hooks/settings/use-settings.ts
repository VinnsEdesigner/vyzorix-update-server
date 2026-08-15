import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import {
  settings,
  type Thresholds,
  type NotificationSettings,
  type ClientSettings,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

interface UserSettings {
  client: ClientSettings;
  thresholds: Thresholds;
}

export function useSettings(
  options?: Omit<UseQueryOptions<UserSettings>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.settings,
    queryFn: () => settings.getSettings(),
    ...options,
  });
}

export function useThresholds(
  options?: Omit<UseQueryOptions<Thresholds>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.thresholds,
    queryFn: () => settings.getThresholds(),
    ...options,
  });
}

export function useNotifications(
  options?: Omit<UseQueryOptions<NotificationSettings>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.notifications,
    queryFn: () => settings.getNotifications(),
    ...options,
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { client?: Partial<ClientSettings> }) => settings.updateSettings(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.settings, updated);
      queryClient.setQueryData(queryKeys.thresholds, updated.thresholds);
    },
  });
}

export function useUpdateThresholds() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Thresholds>) => settings.updateThresholds(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.thresholds, updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.settings });
    },
  });
}

export function useUpdateNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<NotificationSettings>) => settings.updateNotifications(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.notifications, updated);
    },
  });
}

export function useTestWebhook() {
  return useMutation({
    mutationFn: (url: string) => settings.testWebhook(url),
  });
}

export function useRotateWebhookSecret() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => settings.rotateWebhookSecret(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
    },
  });
}

export type { Thresholds, NotificationSettings, ClientSettings };
