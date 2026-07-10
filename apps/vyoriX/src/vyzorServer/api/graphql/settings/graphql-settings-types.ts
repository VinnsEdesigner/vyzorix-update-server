import type { 
  Settings, 
  ThresholdSettings, 
  NotificationSettings, 
  OperatorInfo,
  ConnectionSettings,
  AdvancedSettings,
  NotificationEvent 
} from "@/domain/settings";

export type RawSettings = {
  __typename?: "VyzorixSettings";
} & Omit<Settings, "operator" | "connection" | "thresholds" | "notifications" | "advanced"> & {
  operator: RawOperatorInfo;
  connection: RawConnectionSettings;
  thresholds: RawThresholdSettings;
  notifications: RawNotificationSettings;
  advanced: RawAdvancedSettings;
};

export type RawOperatorInfo = {
  __typename?: "OperatorInfo";
} & Omit<OperatorInfo, "createdAt"> & {
  createdAt: string;
};

export type RawConnectionSettings = {
  __typename?: "ConnectionSettings";
} & ConnectionSettings;

export type RawThresholdSettings = {
  __typename?: "ThresholdSettings";
} & ThresholdSettings;

export type RawAdvancedSettings = {
  __typename?: "AdvancedSettings";
} & AdvancedSettings;

export type RawNotificationChannel = {
  enabled: boolean;
  email?: string;
  webhookUrl?: string;
  webhookSecret?: string;
};

export type RawNotificationSettings = {
  __typename?: "NotificationSettings";
} & Omit<NotificationSettings, "email" | "push" | "webhook"> & {
  email?: RawNotificationChannel;
  push?: RawNotificationChannel;
  webhook?: RawNotificationChannel;
  events: Partial<Record<NotificationEvent, boolean>>;
};

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
