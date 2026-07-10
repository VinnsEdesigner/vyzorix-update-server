export type ApiKeyScope = "read" | "write" | "admin";

export type ApiKeyStatus = "active" | "expired" | "revoked";

export interface ApiKey {
  id: string;
  operatorId: string;
  name: string;
  keyPrefix: string;
  scope: ApiKeyScope;
  expiresAt: Date | null;
  isActive: boolean;
  requestCount: number;
  lastRequestAt: Date | null;
  createdAt: Date;
  updatedAt: Date;
  revokedAt: Date | null;
}

export interface ApiKeyWithSecret extends ApiKey {
  apiKey: string;
}

export interface CreateApiKeyRequest {
  name: string;
  scope: ApiKeyScope;
  expiresInDays?: number;
}

export interface UpdateApiKeyRequest {
  name?: string;
  scope?: ApiKeyScope;
}

export interface ApiKeyStats {
  monthlyLimit: number;
  keysCreatedThisMonth: number;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface ApiKeyListResponse {
  keys: ApiKey[];
  pagination: Pagination;
  monthlyLimit: number;
  keysCreatedThisMonth: number;
}

export function getApiKeyStatus(key: ApiKey): ApiKeyStatus {
  if (key.revokedAt) return "revoked";
  if (key.expiresAt && key.expiresAt < new Date()) return "expired";
  return "active";
}

export function isKeyUsable(key: ApiKey): boolean {
  return key.isActive && getApiKeyStatus(key) === "active";
}
