/**
 * Registration GraphQL Types
 * 
 * Raw GraphQL response types for registration.
 */

import type { InboxEntry, Device, TelemetryFrame } from "@/domain/registration";

/**
 * Raw inbox entry response from GraphQL
 */
export type RawInboxEntry = {
  __typename?: "InboxEntry";
} & Omit<InboxEntry, "receivedAt" | "updatedAt" | "acknowledgedAt" | "approvedAt" | "rejectedAt"> & {
  receivedAt: string;
  updatedAt: string;
  acknowledgedAt?: string | null;
  approvedAt?: string | null;
  rejectedAt?: string | null;
};

/**
 * Raw device response from GraphQL
 */
export type RawDevice = {
  __typename?: "Device";
} & Omit<Device, "registeredAt" | "lastSeen"> & {
  registeredAt?: string | null;
  lastSeen?: string | null;
};

/**
 * Raw telemetry frame from GraphQL
 */
export type RawTelemetryFrame = {
  __typename?: "TelemetryFrame";
} & Omit<TelemetryFrame, "timestamp"> & {
  timestamp: string;
};

/**
 * Raw inbox connection response
 */
export interface RawInboxConnection {
  entries: RawInboxEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

/**
 * Raw device connection response
 */
export interface RawDeviceConnection {
  devices: RawDevice[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

/**
 * Raw telemetry connection response
 */
export interface RawTelemetryConnection {
  frames: RawTelemetryFrame[];
  pagination: {
    limit: number;
    total: number;
  };
}

/**
 * Registration request input (device-side)
 */
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

/**
 * Registration request response
 */
export interface RawRegistrationRequestResponse {
  success: boolean;
  status: string;
  messageId?: string;
  error?: string;
}

/**
 * Acknowledge response
 */
export interface RawAcknowledgeResponse {
  success: boolean;
  status: string;
  error?: string;
}

/**
 * Register device response
 */
export interface RawRegisterDeviceResponse {
  success: boolean;
  status: string;
  deviceId?: string;
  message?: string;
  error?: string;
}

/**
 * Confirm registration response
 */
export interface RawConfirmResponse {
  success: boolean;
  status: string;
  deviceId?: string;
  commandSecret?: string;
  error?: string;
}

/**
 * Deregister response
 */
export interface RawDeregisterResponse {
  success: boolean;
  status: string;
  error?: string;
}

/**
 * Dismiss response
 */
export interface RawDismissResponse {
  success: boolean;
  status: string;
  error?: string;
}
