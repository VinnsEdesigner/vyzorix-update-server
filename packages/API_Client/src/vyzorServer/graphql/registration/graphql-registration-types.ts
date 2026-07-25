export type RawInboxEntry = {
  __typename?: "InboxEntry";
  id: string;
  imei: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  firebaseInstallId?: string;
  status: string;
  notes?: string;
  operatorId?: string;
  createdAt: number;
  approvedAt?: number;
  rejectedAt?: number;
};

export interface RawInboxConnection {
  requests: RawInboxEntry[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
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
