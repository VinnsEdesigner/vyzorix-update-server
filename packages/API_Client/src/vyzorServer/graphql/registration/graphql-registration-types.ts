

import type { InboxEntry, Device } from "@/domain/registration";
import type { TelemetryFrame } from "@/domain/telemetry";


export type RawInboxEntry = {
  __typename?: "InboxEntry";
} & Omit<InboxEntry, "receivedAt" | "updatedAt" | "acknowledgedAt" | "approvedAt" | "rejectedAt"> & {
  receivedAt: string;
  updatedAt: string;
  acknowledgedAt?: string | null;
  approvedAt?: string | null;
  rejectedAt?: string | null;
};


export type RawDevice = {
  __typename?: "Device";
} & Omit<Device, "registeredAt" | "lastSeen"> & {
  registeredAt?: string | null;
  lastSeen?: string | null;
};


export type RawTelemetryFrame = {
  __typename?: "TelemetryFrame";
} & Omit<TelemetryFrame, "timestamp"> & {
  timestamp: string;
};


export interface RawInboxConnection {
  entries: RawInboxEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}


export interface RawDeviceConnection {
  devices: RawDevice[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}


export interface RawTelemetryConnection {
  frames: RawTelemetryFrame[];
  pagination: {
    limit: number;
    total: number;
  };
}


export interface RegistrationRequestInput {
  imei: string;
  deviceName: string;
  model?: string;
  manufacturer?: string;
  osVersion: string;
  appVersion: string;
  fcmToken: string;
  firmware?: string;
  securityPatch?: string;
  buildId?: string;
}


export interface RawRegistrationRequestResponse {
  success: boolean;
  status: string;
  messageId?: string;
  error?: string;
}


export interface RawAcknowledgeResponse {
  success: boolean;
  status: string;
  error?: string;
}


export interface RawRegisterDeviceResponse {
  success: boolean;
  status: string;
  deviceId?: string;
  message?: string;
  error?: string;
}


export interface RawConfirmResponse {
  success: boolean;
  status: string;
  deviceId?: string;
  commandSecret?: string;
  error?: string;
}


export interface RawDeregisterResponse {
  success: boolean;
  status: string;
  error?: string;
}


export interface RawDismissResponse {
  success: boolean;
  status: string;
  error?: string;
}
