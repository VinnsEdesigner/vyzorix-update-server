// API key domain — generated types + zod structural validation + hand-rolled
// business rules. Entity types and Raw mappers eliminated.
import {
  CreateAPIKeyRequestSchema,
  UpdateAPIKeyRequestSchema,
} from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  APIKey,
  APIKeyWithSecret,
  APIKeyListResult,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  GlobalAPIKeyStatsResult,
  OperatorAPIKeyStatsResult,
  TopOperatorStat,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  APIKey,
  APIKeyWithSecret,
  APIKeyListResult,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  GlobalAPIKeyStatsResult,
  OperatorAPIKeyStatsResult,
  TopOperatorStat,
};

// ---- Constants (hand-rolled, not in OpenAPI) ----

export type ApiKeyScope = 'read' | 'write' | 'admin';
export type ApiKeyStatus = 'active' | 'expired' | 'revoked';

export const API_KEY_PREFIX = 'vxyz_';
export const MAX_KEYS_PER_MONTH = 20;
export const MAX_NAME_LENGTH = 64;
export const MIN_NAME_LENGTH = 1;
export const MAX_EXPIRY_DAYS = 365;
export const DEFAULT_EXPIRY_DAYS = null;
export const RATE_LIMIT_PER_MINUTE = 100;

export const SCOPE_LABELS: Record<ApiKeyScope, string> = {
  read: 'Read Only',
  write: 'Read/Write',
  admin: 'Admin',
};

export const SCOPE_DESCRIPTIONS: Record<ApiKeyScope, string> = {
  read: 'Can only make GET requests',
  write: 'Can make GET, POST, PUT, PATCH requests',
  admin: 'Full access including DELETE',
};

export const SCOPE_PERMISSIONS: Record<ApiKeyScope, string[]> = {
  read: ['GET', 'HEAD', 'OPTIONS'],
  write: ['GET', 'POST', 'PUT', 'PATCH', 'HEAD', 'OPTIONS'],
  admin: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'],
};

const NAME_PATTERN = /^[a-zA-Z0-9\s\-_]+$/;
const VALID_SCOPES: ApiKeyScope[] = ['read', 'write', 'admin'];

// ---- Validators (business rules on generated zod) ----

export const createApiKeyValidator = CreateAPIKeyRequestSchema
  .refine((r) => r.name && r.name.trim().length >= MIN_NAME_LENGTH, {
    message: 'Name is required',
    path: ['name'],
  })
  .refine((r) => !r.name || r.name.length <= MAX_NAME_LENGTH, {
    message: `Name must be ${MAX_NAME_LENGTH} characters or less`,
    path: ['name'],
  })
  .refine((r) => !r.name || NAME_PATTERN.test(r.name), {
    message: 'Name may only contain letters, numbers, spaces, hyphens, and underscores',
    path: ['name'],
  })
  .refine((r) => VALID_SCOPES.includes(r.scope as ApiKeyScope), {
    message: `Scope must be one of: ${VALID_SCOPES.join(', ')}`,
    path: ['scope'],
  })
  .refine((r) => !r.expires_in_days || (Number.isInteger(r.expires_in_days) && r.expires_in_days >= 1 && r.expires_in_days <= MAX_EXPIRY_DAYS), {
    message: `Expiry days must be a positive whole number (max ${MAX_EXPIRY_DAYS})`,
    path: ['expires_in_days'],
  });

export function validateCreateApiKeyRequest(input: unknown) {
  return createApiKeyValidator.safeParse(input);
}

export const updateApiKeyValidator = UpdateAPIKeyRequestSchema
  .refine((r) => !r.name || (r.name.trim().length >= MIN_NAME_LENGTH && r.name.length <= MAX_NAME_LENGTH && NAME_PATTERN.test(r.name)), {
    message: `Name must be ${MIN_NAME_LENGTH}-${MAX_NAME_LENGTH} characters (alphanumeric, spaces, hyphens, underscores)`,
    path: ['name'],
  })
  .refine((r) => !r.scope || VALID_SCOPES.includes(r.scope as ApiKeyScope), {
    message: `Scope must be one of: ${VALID_SCOPES.join(', ')}`,
    path: ['scope'],
  });

export function validateUpdateApiKeyRequest(input: unknown) {
  return updateApiKeyValidator.safeParse(input);
}

// ---- Field-level validators (for forms) ----

export function validateApiKeyName(name: string): string | null {
  if (!name || name.trim().length === 0) return 'Name is required';
  if (name.length > MAX_NAME_LENGTH) return `Name must be ${MAX_NAME_LENGTH} characters or less`;
  if (name.length < MIN_NAME_LENGTH) return `Name must be at least ${MIN_NAME_LENGTH} character`;
  if (!NAME_PATTERN.test(name)) return 'Name may only contain letters, numbers, spaces, hyphens, and underscores';
  return null;
}

export function validateExpiryDays(days: number | null): string | null {
  if (days === null || days === undefined) return null;
  if (!Number.isInteger(days) || days < 1) return 'Expiry days must be a positive whole number';
  if (days > MAX_EXPIRY_DAYS) return `Expiry days cannot exceed ${MAX_EXPIRY_DAYS}`;
  return null;
}

export function validateScope(scope: string): string | null {
  if (!VALID_SCOPES.includes(scope as ApiKeyScope)) return `Scope must be one of: ${VALID_SCOPES.join(', ')}`;
  return null;
}

export function parseScope(scope: string): ApiKeyScope | null {
  return VALID_SCOPES.includes(scope as ApiKeyScope) ? (scope as ApiKeyScope) : null;
}

// ---- Hook-facing input types (camelCase; mapped to wire DTOs by hooks) ----

export interface CreateApiKeyInput {
  name: string;
  scope: ApiKeyScope;
  expiresInDays?: number | null;
}

export interface UpdateApiKeyInput {
  name?: string;
  scope?: ApiKeyScope;
}
