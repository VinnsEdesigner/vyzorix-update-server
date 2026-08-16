

import { restClient, setAuthToken, setRefreshToken } from "../_shared/rest-client";
import {
  mfaStatusResponseFromRaw,
  mfaEnrollResponseFromRaw,
  mfaVerifySetupFromRaw,
  mfaEnableFromRaw,
  mfaVerifyBackupFromRaw,
  mfaRegenerateCodesFromRaw,
  mfaVerifyResponseFromRaw,
  mfaCodeRequestToRaw,
  backupCodeVerifyRequestToRaw,
  mfaVerifyRequestToRaw,
  type RawMFAVerifyResponse,
} from "../../../domain/auth";
import type {
  MFAStatusResponse,
  MFAEnrollResponse,
  MFAEnableResponse,
  MFAVerifySetupResponse,
  MFADisableResponse,
  MFAVerifyBackupResponse,
  MFARegenerateCodesResponse,
  MFAVerifyResponse,
} from "../../../domain/auth";

const MFA_PATHS = {
  status: "/v1/auth/mfa/status",
  enroll: "/v1/auth/mfa/enroll",
  verifySetup: "/v1/auth/mfa/verify-setup",
  enable: "/v1/auth/mfa/enable",
  disable: "/v1/auth/mfa/disable",
  verify: "/v1/auth/mfa/verify",
  verifyBackup: "/v1/auth/mfa/verify-backup",
  regenerateBackupCodes: "/v1/auth/mfa/regenerate-backup-codes",
} as const;


export async function getMFAStatus(): Promise<MFAStatusResponse> {
  const raw = await restClient.get<{ mfa_enabled: boolean; backup_codes?: string[] }>(MFA_PATHS.status);
  return mfaStatusResponseFromRaw(raw);
}


export async function enrollMFA(): Promise<MFAEnrollResponse> {
  const raw = await restClient.post<{ secret: string; uri: string }>(MFA_PATHS.enroll);
  return mfaEnrollResponseFromRaw(raw);
}


export async function verifyMFASetup(code: string): Promise<MFAVerifySetupResponse> {
  const raw = await restClient.post<{ verified: boolean }>(MFA_PATHS.verifySetup, { code, token: code });
  return mfaVerifySetupFromRaw(raw);
}


export async function enableMFA(code: string): Promise<MFAEnableResponse> {
  const raw = await restClient.post<{ success: boolean; backup_codes?: string[] }>(MFA_PATHS.enable, { code, token: code });
  return mfaEnableFromRaw(raw);
}


export async function disableMFA(code: string): Promise<MFADisableResponse> {
  const raw = await restClient.post<{ success: boolean }>(MFA_PATHS.disable, mfaCodeRequestToRaw(code));
  return { success: raw.success };
}


export async function verifyBackupCode(code: string): Promise<MFAVerifyBackupResponse> {
  const raw = await restClient.post<{ valid: boolean }>(MFA_PATHS.verifyBackup, backupCodeVerifyRequestToRaw(code));
  return mfaVerifyBackupFromRaw(raw);
}


export async function regenerateBackupCodes(): Promise<MFARegenerateCodesResponse> {
  const raw = await restClient.post<{ backup_codes: string[] }>(MFA_PATHS.regenerateBackupCodes);
  return mfaRegenerateCodesFromRaw(raw);
}


export async function verifyMFA(operatorId: string, code: string): Promise<MFAVerifyResponse> {
  const raw = await restClient.post<RawMFAVerifyResponse>(MFA_PATHS.verify, mfaVerifyRequestToRaw(operatorId, code));
  const result = mfaVerifyResponseFromRaw(raw);
  if (result.success && result.accessToken && result.refreshToken) {
    setAuthToken(result.accessToken);
    setRefreshToken(result.refreshToken);
  }
  return result;
}
