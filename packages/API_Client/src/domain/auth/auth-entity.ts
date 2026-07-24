import type { Thresholds, ClientSettings } from "../settings/settings-entity";

export type OperatorRole = "super_admin" | "admin" | "operator" | "viewer";

export interface OrganizationInfo {
  id: string;
  name: string;
  role: string;
}

export interface Operator {
  id: string;
  email: string;
  name: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  needs_organization: boolean;
  organizations: OrganizationInfo[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
}

export interface MeResponse extends Operator {
  thresholds?: Thresholds;
  client?: ClientSettings;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  sessionId: string;
}

export interface LoginResponse {
  operatorId: string;
  email: string;
  name: string;
  role: OperatorRole;
  mfaEnabled: boolean;
}

export interface LoginMFARequiredResponse {
  mfaRequired: true;
  operatorId: string;
}

export interface LoginWithTokensResponse {
  operator_id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  access_token: string;
  refresh_token: string;
  expires_at: number;
  session_id: string;
  needs_organization: boolean;
  organizations: OrganizationInfo[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
}

export interface LoginWithTokensMFARequiredResponse {
  mfaRequired: true;
  operatorId: string;
  email: string;
  name: string;
  role: OperatorRole;
  mfaEnabled: boolean;
}

export interface RegisterResponse {
  operatorId: string;
  email: string;
  name: string;
}

export interface MeResponse extends Operator {
  thresholds?: Thresholds;
  client?: ClientSettings;
}

export interface UpdateNameRequest {
  name: string;
}

export interface ForgotPasswordResponse {
  success: boolean;
}

export interface ResetPasswordResponse {
  success: boolean;
}

export interface VerifyEmailResponse {
  success: boolean;
}

export interface MFAStatusResponse {
  enabled: boolean;
  backupCodes?: string[];
}

export interface MFAEnrollResponse {
  secret: string;
  qrCodeUrl: string;
}

export interface MFAVerifyResponse {
  success: boolean;
}

export interface MFAEnableResponse {
  success: boolean;
  backupCodes?: string[];
}
