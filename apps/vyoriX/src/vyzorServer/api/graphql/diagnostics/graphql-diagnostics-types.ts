export type RawIdentityInfo = {
  __typename?: "IdentityInfo";
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
};

export type RawSoftwareInfo = {
  __typename?: "SoftwareInfo";
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  buildId?: string;
};

export type RawRegistrationInfo = {
  __typename?: "RegistrationInfo";
  status: string;
  registeredAt?: string | null;
  fcmTokenValid: boolean;
  fcmTokenRefreshedAt?: string | null;
  commandSecretSet: boolean;
};

export type RawConnectionInfo = {
  __typename?: "ConnectionInfo";
  webSocketStatus?: string;
  connectedAt?: string | null;
  fcmStatus?: string;
  lastSeen?: string | null;
  clientIp?: string;
  protocol?: string;
};

export type RawTelemetryInfo = {
  __typename?: "TelemetryInfo";
  lastTimestamp?: string | null;
  framesToday: number;
  avgLatencyMs?: number;
  totalBytesToday?: number;
  sessionsToday: number;
};

export type RawDeviceInspection = {
  __typename?: "DeviceInspection";
  identity: RawIdentityInfo;
  software: RawSoftwareInfo;
  registration: RawRegistrationInfo;
  connection: RawConnectionInfo;
  telemetry: RawTelemetryInfo;
};

export type RawTimelineEvent = {
  __typename?: "TimelineEvent";
  id: string;
  type: string;
  timestamp: string;
  data?: Record<string, unknown>;
};

export interface RawTimelineConnection {
  events: RawTimelineEvent[];
  hasMore: boolean;
  nextCursor?: string;
}
