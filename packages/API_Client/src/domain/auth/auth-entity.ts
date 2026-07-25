
import type { Thresholds, ClientSettings } from "../settings/settings-entity";
import type { OrganizationRole } from "../organization/organization-entity";

export type OperatorRole = "super_admin" | "admin" | "operator" | "viewer";

export interface OrganizationInfo {
  id: string;
  name: string;
  role: OrganizationRole;
}

export interface OrganizationMembership {
  id: string;
  organization_id: string;
  organization_name: string;
  role: OrganizationRole;
  joined_at: string;
}

export interface Operator {
  id: string;
  email: string;
  name: string;
  role?: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  needs_organization: boolean;
  organizations: OrganizationInfo[];
  memberships: OrganizationMembership[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
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
  selectedOrganization?: OrganizationInfo;
  lastOrganizationId?: string;
  organizations: OrganizationInfo[];
  needsOrganization: boolean;
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
  memberships: OrganizationMembership[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
}

export interface LoginWithTokensMFARequiredResponse {
  mfaRequired: true;
  operatorId: string;
  email: string;
  name: string;
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

export interface MFAStatusResponse {
  enabled: boolean;
  backupCodes?: string[];
}

export interface MFAEnrollResponse {
  secret: string;
  uri: string;
}

export interface MFAVerifyResponse {
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
  operator?: {
    id: string;
    email: string;
    name: string;
    mfaEnabled: boolean;
  };
}

export interface MFAEnableResponse {
  success: boolean;
  backupCodes?: string[];
}

export interface MFAVerifySetupResponse {
  verified: boolean;
}

export interface MFADisableResponse {
  success: boolean;
}

export interface MFAVerifyBackupResponse {
  valid: boolean;
}

export interface MFARegenerateCodesResponse {
  backupCodes: string[];
}
