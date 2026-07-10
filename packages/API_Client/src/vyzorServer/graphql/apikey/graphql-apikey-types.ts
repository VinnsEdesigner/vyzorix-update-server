import type { ApiKeyScope } from "@/domain/apikey";

export type RawApiKey = {
  __typename?: "ApiKey";
  id: string;
  operatorId: string;
  name: string;
  keyPrefix: string;
  scope: ApiKeyScope;
  expiresAt: string | null;
  isActive: boolean;
  requestCount: number;
  lastRequestAt: string | null;
  createdAt: string;
  updatedAt: string;
  revokedAt: string | null;
};

export type RawApiKeyWithSecret = RawApiKey & {
  apiKey: string;
};

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface RawApiKeyConnection {
  keys: RawApiKey[];
  pagination: RawPagination;
  monthlyLimit: number;
  keysCreatedThisMonth: number;
}

export interface RawCreateApiKeyResponse {
  success: boolean;
  key?: RawApiKeyWithSecret;
  error?: string;
}

export interface RawUpdateApiKeyResponse {
  success: boolean;
  key?: RawApiKey;
  error?: string;
}

export interface RawRevokeApiKeyResponse {
  success: boolean;
  error?: string;
}

export interface RawRotateApiKeyResponse {
  success: boolean;
  key?: RawApiKeyWithSecret;
  error?: string;
}
