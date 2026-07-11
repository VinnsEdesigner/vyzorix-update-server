import { restClient } from "../_shared/rest-client";
import {
  thresholdsFromRaw,
  clientSettingsFromRaw,
  notificationSettingsFromRaw,
  thresholdsToRaw,
  clientSettingsToRaw,
  notificationSettingsToRaw,
  type RawThresholds,
  type RawClientSettings,
  type RawNotificationSettings,
  type RawNotificationSettingsResponse,
} from "@/domain/settings";
import { meResponseFromRaw } from "@/domain/auth";
import type { Operator } from "@/domain/auth";
import type {
  Thresholds,
  ClientSettings,
  NotificationSettings,
} from "@/domain/settings";

const PATHS = {
  settings: "/v1/auth/me/settings",
  thresholds: "/v1/auth/me/thresholds",
  notifications: "/v1/auth/me/notifications",
  webhookTest: "/v1/auth/me/notifications/webhook/test",
  webhookRotate: "/v1/auth/me/notifications/webhook/rotate",
} as const;

interface RawSettingsResponse {
  id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  client: RawClientSettings;
  thresholds: RawThresholds;
}

export const settings = {
  async getSettings(): Promise<{ client: ClientSettings; thresholds: Thresholds }> {
    const response = await restClient.get<RawSettingsResponse>(PATHS.settings);
    return {
      client: clientSettingsFromRaw(response.client),
      thresholds: thresholdsFromRaw(response.thresholds),
    };
  },

  async updateSettings(data: { client?: Partial<ClientSettings> }): Promise<{ client: ClientSettings; thresholds: Thresholds }> {
    const response = await restClient.patch<RawSettingsResponse>(PATHS.settings, {
      client: data.client ? clientSettingsToRaw(data.client) : undefined,
    });
    return {
      client: clientSettingsFromRaw(response.client),
      thresholds: thresholdsFromRaw(response.thresholds),
    };
  },

  async getThresholds(): Promise<Thresholds> {
    const response = await restClient.get<RawThresholds>(PATHS.thresholds);
    return thresholdsFromRaw(response);
  },

  async updateThresholds(data: Partial<Thresholds>): Promise<Thresholds> {
    const response = await restClient.patch<RawThresholds>(PATHS.thresholds, thresholdsToRaw(data));
    return thresholdsFromRaw(response);
  },

  async getNotifications(): Promise<NotificationSettings> {
    const response = await restClient.get<RawNotificationSettings>(PATHS.notifications);
    return notificationSettingsFromRaw(response);
  },

  async updateNotifications(data: Partial<NotificationSettings>): Promise<NotificationSettings> {
    const response = await restClient.patch<RawNotificationSettings>(PATHS.notifications, notificationSettingsToRaw(data));
    return notificationSettingsFromRaw(response);
  },

  async testWebhook(url: string): Promise<{
    success: boolean;
    statusCode?: number;
    responseTime?: number;
    error?: string;
  }> {
    return restClient.post<{
      success: boolean;
      statusCode?: number;
      responseTime?: number;
      error?: string;
    }>(PATHS.webhookTest, { url });
  },

  async rotateWebhookSecret(): Promise<{ secret: string }> {
    return restClient.post<{ secret: string }>(PATHS.webhookRotate);
  },
};
