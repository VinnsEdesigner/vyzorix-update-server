

import { restClient } from "../_shared/rest-client";
import {
  mfaStatusFromRaw,
  mfaEnrollFromRaw,
  mfaEnableFromRaw,
  mfaVerifySetupFromRaw,
  mfaDisableFromRaw,
  mfaVerifyBackupFromRaw,
  mfaRegenerateCodesFromRaw,
  mfaVerifyFromRaw,
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
  const response = await restClient.get<{ mfa_enabled: boolean }>(MFA_PATHS.status);
  return mfaStatusFromRaw(response);
}


export async function enrollMFA(): Promise<MFAEnrollResponse> {
  const response = await restClient.post<{ secret: string; uri: string }>(MFA_PATHS.enroll);
  return mfaEnrollFromRaw(response);
}


export async function verifyMFASetup(code: string): Promise<MFAVerifySetupResponse> {
  const response = await restClient.post<{ verified: boolean }>(MFA_PATHS.verifySetup, { code });
  return mfaVerifySetupFromRaw(response);
}


export async function enableMFA(code: string): Promise<MFAEnableResponse> {
  const response = await restClient.post<{ success: boolean; backup_codes?: string[] }>(MFA_PATHS.enable, { code });
  return mfaEnableFromRaw(response);
}


export async function disableMFA(code: string): Promise<MFADisableResponse> {
  const response = await restClient.post<{ success: boolean }>(MFA_PATHS.disable, { code });
  return mfaDisableFromRaw(response);
}


export async function verifyBackupCode(code: string): Promise<MFAVerifyBackupResponse> {
  const response = await restClient.post<{ valid: boolean }>(MFA_PATHS.verifyBackup, { code });
  return mfaVerifyBackupFromRaw(response);
}


export async function regenerateBackupCodes(): Promise<MFARegenerateCodesResponse> {
  const response = await restClient.post<{ backup_codes: string[] }>(MFA_PATHS.regenerateBackupCodes);
  return mfaRegenerateCodesFromRaw(response);
}


export async function verifyMFA(operatorId: string, code: string): Promise<MFAVerifyResponse> {
  const response = await restClient.post<{
    success: boolean;
    session_id?: string;
    access_token?: string;
    refresh_token?: string;
    expires_at?: number;
    operator?: {
      id: string;
      email: string;
      name: string;
      role: string;
      mfa_enabled: boolean;
    };
  }>(MFA_PATHS.verify, {
    operator_id: operatorId,
    code,
  });
  return mfaVerifyFromRaw(response);
}
