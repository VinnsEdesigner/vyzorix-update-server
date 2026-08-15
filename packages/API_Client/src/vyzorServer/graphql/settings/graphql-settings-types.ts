import type { Thresholds, NotificationEvent } from "../../../domain/settings";

export interface RawOperatorInfo {
  __typename?: "OperatorInfo";
  id: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface RawConnectionSettings {
  __typename?: "ConnectionSettings";
  serverUrl: string;
  deviceId: string;
}

export type RawThresholdSettings = {
  __typename?: "ThresholdSettings";
} & Thresholds;

export interface RawAdvancedSettings {
  __typename?: "AdvancedSettings";
  logBufferLimit: number;
}

export interface RawNotificationChannel {
  enabled: boolean;
  email?: string;
  webhookUrl?: string;
  webhookSecret?: string;
}

export interface RawNotificationSettings {
  __typename?: "NotificationSettings";
  email?: RawNotificationChannel;
  push?: RawNotificationChannel;
  webhook?: RawNotificationChannel;
  events: Partial<Record<NotificationEvent, boolean>>;
}

export interface RawSettings {
  __typename?: "VyzorixSettings";
  operator: RawOperatorInfo;
  connection: RawConnectionSettings;
  thresholds: RawThresholdSettings;
  notifications: RawNotificationSettings;
  advanced: RawAdvancedSettings;
}

export interface RawUpdateSettingsResponse {
  success: boolean;
  settings?: RawSettings;
  error?: string;
}

export interface RawUpdateThresholdsResponse {
  success: boolean;
  thresholds?: RawThresholdSettings;
  error?: string;
}
