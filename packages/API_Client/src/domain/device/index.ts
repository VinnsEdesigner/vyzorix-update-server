// Device domain — generated types + hand-rolled business rules.
// IMEI Luhn checksum validation and FCM token validation are genuine
// algorithms that can't be derived from the OpenAPI spec.
import type {
  DeviceListItem,
  DeviceListResult,
  DeviceDetailResult,
  DeviceCountResult,
  DeviceTagsResult,
  DeviceTagAddedResult,
  DeviceTagRemovedResult,
  DeviceConfirmResult,
  DeviceConfirmRequest,
  DeviceFCMTokenRequest,
  SetDeviceTagsRequest,
  DeviceTransferRequest,
  DeviceTransferResult,
  DeviceSettingsResult,
  UpdateDeviceSettingsRequest,
  ThresholdsResult,
  ConnectionStatusResult,
  ConnectionListResult,
  ConnectionMetricsResult,
  DeviceDisconnectResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  DeviceListItem,
  DeviceListResult,
  DeviceDetailResult,
  DeviceCountResult,
  DeviceTagsResult,
  DeviceTagAddedResult,
  DeviceTagRemovedResult,
  DeviceConfirmResult,
  DeviceConfirmRequest,
  DeviceFCMTokenRequest,
  SetDeviceTagsRequest,
  DeviceTransferRequest,
  DeviceTransferResult,
  DeviceSettingsResult,
  UpdateDeviceSettingsRequest,
  ThresholdsResult,
  ConnectionStatusResult,
  ConnectionListResult,
  ConnectionMetricsResult,
  DeviceDisconnectResult,
};

// ---- Constants (hand-rolled) ----

export type DeviceStatus = 'online' | 'offline' | 'deregistered';

// ---- IMEI validation (Luhn checksum — genuine algorithm) ----

export function validateIMEI(imei: string): boolean {
  if (!/^\d{15}$/.test(imei)) return false;
  return validateIMEIChecksum(imei);
}

export function validateIMEIChecksum(imei: string): boolean {
  if (!/^\d{15}$/.test(imei)) return false;
  let sum = 0;
  let isEven = false;
  for (let i = imei.length - 1; i >= 0; i--) {
    const ch = imei[i];
    if (ch === undefined) return false;
    let digit = parseInt(ch, 10);
    if (isEven) {
      digit *= 2;
      if (digit > 9) digit -= 9;
    }
    sum += digit;
    isEven = !isEven;
  }
  return sum % 10 === 0;
}

// ---- FCM token validation (business rule: length constraints) ----

export function validateFCMToken(token: string): string | null {
  if (!token) return 'FCM token is required';
  if (token.length < 100 || token.length > 4000) return 'FCM token has invalid length';
  return null;
}
