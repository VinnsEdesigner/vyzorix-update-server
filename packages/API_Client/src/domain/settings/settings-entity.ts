export type NotificationEvent =
  | "threshold_breach"
  | "device_offline"
  | "device_online"
  | "update_available"
  | "command_failed"
  | "registration_request";

export interface Thresholds {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

export interface ClientSettings {
  serverUrl: string;
  deviceId: string;
  requestTimeoutMs: number;
  logBufferLimit: number;
  signalHistoryLimit: number;
  autoReconnect: boolean;
  strictHmac: boolean;
  notificationsEnabled: boolean;
}

export interface EmailNotifications {
  thresholdBreach: boolean;
  deviceOffline: boolean;
  deviceOnline: boolean;
  updateAvailable: boolean;
  commandFailed: boolean;
  registrationRequest: boolean;
}

export interface PushNotifications {
  thresholdBreach: boolean;
  deviceOffline: boolean;
  deviceOnline: boolean;
  updateAvailable: boolean;
  commandFailed: boolean;
  registrationRequest: boolean;
}

export interface WebhookNotifications {
  enabled: boolean;
  url: string;
  secret: string;
  types: NotificationEvent[];
}

export interface NotificationSettings {
  enabled: boolean;
  channels: string[];
  email: EmailNotifications;
  push: PushNotifications;
  webhook: WebhookNotifications;
}

export interface SecuritySettings {
  maxConcurrentSessions: number;
  passwordMinAgeDays: number;
  passwordMaxAgeDays: number;
  passwordHistoryCount: number;
  sessionPinRequired: boolean;
}

export const DEFAULT_THRESHOLDS: Thresholds = {
  riskWarn: 70,
  riskCrit: 85,
  thermalWarn: 45,
  thermalCrit: 50,
  bufferWarn: 30,
  bufferCrit: 15,
};

export const DEFAULT_CLIENT_SETTINGS: ClientSettings = {
  serverUrl: "",
  deviceId: "",
  requestTimeoutMs: 8000,
  logBufferLimit: 500,
  signalHistoryLimit: 240,
  autoReconnect: true,
  strictHmac: false,
  notificationsEnabled: true,
};

export const DEFAULT_SECURITY_SETTINGS: SecuritySettings = {
  maxConcurrentSessions: 3,
  passwordMinAgeDays: 0,
  passwordMaxAgeDays: 90,
  passwordHistoryCount: 5,
  sessionPinRequired: false,
};
