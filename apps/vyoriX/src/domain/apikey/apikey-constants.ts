import type { ApiKeyScope } from "./apikey-entity";

export const API_KEY_PREFIX = "vxyz_";

export const MAX_KEYS_PER_MONTH = 20;

export const MAX_NAME_LENGTH = 64;

export const MIN_NAME_LENGTH = 1;

export const MAX_EXPIRY_DAYS = 365;

export const DEFAULT_EXPIRY_DAYS = null;

export const SCOPE_LABELS: Record<ApiKeyScope, string> = {
  read: "Read Only",
  write: "Read/Write",
  admin: "Admin",
};

export const SCOPE_DESCRIPTIONS: Record<ApiKeyScope, string> = {
  read: "Can only make GET requests",
  write: "Can make GET, POST, PUT, PATCH requests",
  admin: "Full access including DELETE",
};

export const SCOPE_PERMISSIONS: Record<ApiKeyScope, string[]> = {
  read: ["GET", "HEAD", "OPTIONS"],
  write: ["GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"],
  admin: ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"],
};

export const RATE_LIMIT_PER_MINUTE = 100;
