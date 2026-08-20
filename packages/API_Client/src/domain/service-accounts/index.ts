export interface ServiceAccount {
  id: string;
  orgId: string;
  name: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ServiceAccountToken {
  id: string;
  serviceId: string;
  name: string;
  keyPrefix: string;
  scopes: string[];
  valid: boolean;
  expiresAt: string | null;
  createdAt: string;
  revokedAt: string | null;
  fullKey?: string;
}

export interface CreateServiceAccountTokenRequest {
  name?: string;
  scopes?: string[];
  expiresAt?: string;
}

interface RawServiceAccount {
  id: string;
  org_id: string;
  name: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface RawServiceAccountToken {
  id: string;
  service_id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  valid: boolean;
  expires_at: string | null;
  created_at: string;
  revoked_at: string | null;
  full_key?: string;
}

export const serviceAccountFromRaw = (raw: RawServiceAccount): ServiceAccount => ({
  id: raw.id,
  orgId: raw.org_id,
  name: raw.name,
  enabled: raw.enabled,
  createdAt: raw.created_at,
  updatedAt: raw.updated_at,
});

export const serviceAccountsFromRaw = (raw: RawServiceAccount[]): ServiceAccount[] => raw.map(serviceAccountFromRaw);

export const serviceAccountTokenFromRaw = (raw: RawServiceAccountToken): ServiceAccountToken => ({
  id: raw.id,
  serviceId: raw.service_id,
  name: raw.name,
  keyPrefix: raw.key_prefix,
  scopes: raw.scopes ?? [],
  valid: raw.valid,
  expiresAt: raw.expires_at,
  createdAt: raw.created_at,
  revokedAt: raw.revoked_at,
  fullKey: raw.full_key,
});

export const serviceAccountTokensFromRaw = (raw: RawServiceAccountToken[]): ServiceAccountToken[] =>
  raw.map(serviceAccountTokenFromRaw);
