// Registration domain — generated types + hand-rolled business rules.
// Device name length, semver validation, IMEI, FCM token, inbox status
// transitions — all genuine business logic.
import { InboxRequestSchema } from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  InboxRequest,
  InboxEntryResponse,
  InboxListResult,
  InboxAckRequest,
  InboxAckResult,
  InboxResendResult,
  UpdateInboxEntryRequest,
} from '../../generated/vyzorixUpdateServerAPI.schemas';
import type { DeviceStatus, Pagination } from '../_shared';

export type {
  InboxRequest,
  InboxEntryResponse,
  InboxListResult,
  InboxAckRequest,
  InboxAckResult,
  InboxResendResult,
  UpdateInboxEntryRequest,
};

// ---- Domain types (not in OpenAPI, used by hooks) ----

/** Normalized inbox entry (Date|null timestamps) used by hooks/UI. The wire
 * DTO is the generated InboxEntryResponse (epoch-millis numbers). */
export interface InboxEntry {
  id: string;
  imei: string;
  deviceName: string;
  deviceClass: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  fcmToken: string;
  firebaseInstallId: string;
  status: InboxStatus;
  acknowledgedAt: Date | null;
  approvingAt: Date | null;
  approvedAt: Date | null;
  rejectedAt: Date | null;
  notes: string | null;
  operatorId: string | null;
  createdAt: Date;
}

/** Normalized inbox list returned by hooks (domain InboxEntry[], not the
 * wire InboxEntryResponse[]). */
export interface InboxEntriesResult {
  requests: InboxEntry[];
  pagination: Pagination;
}

/** Device-submitted registration request. Required fields are enforced by
 * registrationValidator; the generated InboxRequest keeps them optional. */
export interface CreateInboxRequest {
  imei: string;
  deviceName?: string;
  deviceClass?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  fcmToken: string;
  firebaseInstallId: string;
  idempotencyKey?: string;
}

export interface RegisteredDevice {
  id: string;
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  status: DeviceStatus;
  registeredAt: Date | null;
  lastSeen: Date | null;
  online: boolean;
}

export interface RegisteredDeviceListResult {
  devices: RegisteredDevice[];
  pagination: Pagination;
}

export interface DeregisterResult {
  imei: string;
  status: string;
  deregisteredAt: Date;
  retentionUntil: Date;
}

export interface AckResult {
  id: string;
  imei: string;
  status: InboxStatus;
  acknowledgedAt: Date | null;
  approvingAt: Date | null;
  approvedAt: Date | null;
  rejectedAt: Date | null;
  commandSecret: string | null;
  fcmPushSent: boolean;
  notes: string | null;
}

export interface ConfirmDeviceResult {
  device_id: string;
  imei: string;
  confirmed: boolean;
  online: boolean;
  registered_at?: number;
  server_time: number;
}

export interface CreateInboxResult {
  id: string;
  imei: string;
  status: InboxStatus;
  createdAt: Date;
}

// ---- Constants (hand-rolled, not in OpenAPI) ----

export type InboxStatus = 'pending' | 'acknowledged' | 'approving' | 'approved' | 'rejected' | 'expired';
export type AcknowledgeAction = 'acknowledge' | 'approve' | 'reject';

const VALID_ACKNOWLEDGE_ACTIONS: AcknowledgeAction[] = ['acknowledge', 'approve', 'reject'];

const VALID_TRANSITIONS: Record<string, InboxStatus[]> = {
  pending: ['acknowledged', 'rejected'],
  acknowledged: ['approving', 'rejected'],
  approving: ['approved', 'rejected'],
  approved: [],
  rejected: [],
  expired: [],
};

// ---- Validators (business rules on generated zod) ----

export const registrationValidator = InboxRequestSchema
  .refine((r) => r.imei && /^\d{15}$/.test(r.imei), {
    message: 'IMEI must be 15 digits',
    path: ['imei'],
  })
  .refine((r) => !r.deviceName || (r.deviceName.length >= 1 && r.deviceName.length <= 100), {
    message: 'Device name must be 1-100 characters',
    path: ['deviceName'],
  })
  .refine((r) => r.fcmToken && r.fcmToken.length >= 100, {
    message: 'FCM token is required and must be at least 100 characters',
    path: ['fcmToken'],
  });

export function validateRegistrationRequest(input: unknown) {
  return registrationValidator.safeParse(input);
}

// ---- Field validators (for forms) ----

export function validateDeviceName(name: string): string | null {
  if (!name) return 'Device name is required';
  if (name.length < 1 || name.length > 100) return 'Device name must be 1-100 characters';
  return null;
}

export function validateVersion(version: string, fieldName = 'version'): string | null {
  if (!version) return `${fieldName} is required`;
  if (!/^\d+\.\d+\.\d+$/.test(version)) return `${fieldName} must be in semver format (e.g., 1.2.3)`;
  return null;
}

export function validateAcknowledgeAction(action: string): string | null {
  if (!action) return 'Action is required';
  if (!VALID_ACKNOWLEDGE_ACTIONS.includes(action as AcknowledgeAction)) return `Invalid action: ${action}. Valid actions: ${VALID_ACKNOWLEDGE_ACTIONS.join(', ')}`;
  return null;
}

// ---- Status transition state machine (genuine algorithm) ----

export function isValidStatusTransition(currentStatus: string, newStatus: string): boolean {
  const validTargets = VALID_TRANSITIONS[currentStatus] ?? [];
  return validTargets.includes(newStatus as InboxStatus);
}

export function validateStatusTransition(currentStatus: InboxStatus, newStatus: InboxStatus): string | null {
  if (!isValidStatusTransition(currentStatus, newStatus)) {
    return `Cannot transition from ${currentStatus} to ${newStatus}`;
  }
  return null;
}
