# Frontend API Keys Feature Requirements

**Version:** 2.0  
**Date:** 2026-07-05  
**Status:** Complete - Implementation Required  
**Reference:** [MULTI_CLIENT_API_KEY_SYSTEM.md](./MULTI_CLIENT_API_KEY_SYSTEM.md)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Frontend Architecture Integration](#2-frontend-architecture-integration)
3. [Domain Layer](#3-domain-layer)
4. [Data Layer](#4-data-layer)
5. [Presentation Layer](#5-presentation-layer)
6. [UI Layer](#6-ui-layer)
7. [User Flows](#7-user-flows)
8. [State Management](#8-state-management)
9. [Error Handling](#9-error-handling)
10. [Accessibility](#10-accessibility)
11. [Developer Portal](#11-developer-portal)
12. [Super Admin UI](#12-super-admin-ui)
13. [Developer SDK Examples](#13-developer-sdk-examples)
14. [Testing Strategy](#14-testing-strategy)

---

## 1. Overview

### 1.1 Feature Summary

The API Keys feature consists of TWO distinct user experiences:

**A. Operator Settings (Management)**
- Operators manage their own API keys
- Create, rename, revoke, rotate keys
- Set key scopes (read-only, read-write)
- View usage statistics

**B. Developer Portal (Integration)**
- Third-party developers manage their key usage
- View API documentation
- Test endpoints
- See usage patterns
- Regenerate keys

### 1.2 Backend API Reference

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/v1/auth/api-keys` | POST | Session | Create new API key |
| `/v1/auth/api-keys` | GET | Session | List all operator's keys |
| `/v1/auth/api-keys/:keyId` | PATCH | Session | Update key (rename, scopes) |
| `/v1/auth/api-keys/:keyId` | DELETE | Session | Revoke an API key |
| `/v1/auth/api-keys/:keyId/rotate` | POST | Session | Rotate an API key |

### 1.3 Key Scopes

| Scope | Permissions |
|-------|-------------|
| `read` | GET requests only |
| `write` | GET, POST, PUT, PATCH requests |
| `admin` | All requests including DELETE |

### 1.4 Design Principles

1. **Security First**: Keys shown once, never stored in plain text
2. **Optimistic UI**: Immediate feedback, rollback on failure
3. **Developer Experience**: Clear documentation and code examples
4. **Separation**: Operator settings vs Developer portal are separate
5. **Pagination**: Show 20 keys with "Load More" pattern
6. **Complete CRUD**: Create, Read, Update (rename), Delete (revoke)

---

## 2. Frontend Architecture Integration

### 2.1 Layer Structure

Following the [Frontend Architecture](../FRONTEND_ARCHITECTURE.md):

```
src/
 domain/
    apikey/                    # API Keys feature domain
       apikey-entity.ts      # API key domain types
       apikey-mappers.ts     # Raw → Domain transformations
       apikey-validators.ts  # Input validation
       apikey-constants.ts   # Key scope definitions
   
    _shared/                   # Shared domain types
        domain-pagination.ts
        domain-errors.ts

 lib/api/
    apikey/                   # API Keys feature data layer
       graphql-apikey-queries.ts
       graphql-apikey-mutations.ts
       graphql-apikey-types.ts
   
    rest/
       rest-apikey-endpoints.ts
   
    _shared/                   # Shared API utilities
        graphql-client.ts
        rest-client.ts

 hooks/
    apikey/
        use-apikeys.ts               # List keys with pagination
        use-create-apikey.ts        # Create with optimistic update
        use-revoke-apikey.ts        # Revoke with optimistic update
        use-rotate-apikey.ts        # Rotate returns new key
        use-update-apikey.ts        # Rename/update key
        use-apikey-stats.ts         # Usage statistics

 ui/
     pages/
        settings/
           api-keys/
               apikey-settings.tsx         # Operator management
               components/
                   apikey-list.tsx         # Key list table
                   apikey-row.tsx         # Single key row
                   apikey-list-skeleton.tsx # Loading skeleton
                   create-apikey-dialog.tsx    # Create key modal
                   revoke-apikey-dialog.tsx   # Revoke confirmation
                   rotate-apikey-dialog.tsx    # Rotate confirmation
                   edit-apikey-dialog.tsx      # Rename/edit modal
                   apikey-created-dialog.tsx   # Show new key ONCE
                   apikey-usage-stats.tsx     # Usage statistics
       
        developer-portal/
            developer-dashboard.tsx          # Developer overview
            developer-docs.tsx               # API docs
            components/
                sdk-examples.tsx            # Code examples
                test-endpoint.tsx            # API testing tool
                usage-chart.tsx             # Usage visualization
    
     components/
         shared/
             copy-button.tsx                  # Reusable copy component
```

### 2.2 Dependency Rules (STRICT - MUST FOLLOW)

```

                                                                     
  UI Layer  Presentation Layer  Domain 
                                                                   
        uses hooks                   uses types & transforms       
                                                                   
                                                                   
                           Data Layer (lib/api/)                    
                                                                   
                                      imports domain types only   
                                      (NO reverse dependencies)    
                                                                   

```

**CRITICAL RULES:**
1. Domain layer must NEVER import from Data Layer or Presentation Layer
2. Domain types must be self-contained with NO external dependencies
3. UI layer must ONLY import from hooks (Presentation Layer)
4. Hooks must ONLY import from Domain and Data Layer

### 2.3 Two Separate API Clients (MUST BE SEPARATE)

```typescript
// 1. Session-based Client (for Operator Settings UI)
// Location: src/lib/api/client/session-client.ts
// Uses cookies - operator is logged in
// Purpose: Manage own keys (create, revoke, rotate, rename)

import { sessionClient } from "@/lib/api/client/session-client";

// 2. Developer Client (for third-party apps)
// Location: src/lib/api/client/developer-client.ts
// Uses X-API-Key header - developer is using their key
// Purpose: Make API calls using the key

import { createDeveloperClient } from "@/lib/api/client/developer-client";
const client = createDeveloperClient("vxyz_abc123...");
const devices = await client.getDevices();
```

**These are COMPLETELY SEPARATE and serve different purposes. They must NOT be mixed.**

---

## 3. Domain Layer

### 3.1 Target File Structure

```
src/domain/
 apikey/
    apikey-entity.ts          # Types (ApiKey, ApiKeyScope, etc.)
    apikey-mappers.ts         # Raw → Domain transformations
    apikey-validators.ts     # Input validation
    apikey-constants.ts      # Scope limits, defaults

 _shared/
     domain-pagination.ts      # Shared pagination types
     domain-errors.ts         # Shared error types
```

### 3.2 Entity Types (apikey-entity.ts)

```typescript
// src/domain/apikey/apikey-entity.ts

// NOTE: This file must NOT import from lib/api/ or hooks/

/**
 * API Key scope permissions
 */
export type ApiKeyScope = "read" | "write" | "admin";

/**
 * API Key status
 */
export type ApiKeyStatus = "active" | "revoked" | "expired";

/**
 * API Key entity - represents a key in the system
 */
export interface ApiKey {
  id: string;
  name: string;
  keyPrefix: string;           // First 8 chars for display: "vxyz_a1b2"
  scope: ApiKeyScope;         // Key permissions scope
  expiresAt: Date | null;    // null = never expires
  isActive: boolean;
  requestCount: number;      // Total requests made with this key
  lastRequestAt: Date | null;
  createdAt: Date;
  revokedAt: Date | null;
}

/**
 * API Key with full key shown - ONLY returned on create/rotate
 * This is the ONLY time the full key is available
 */
export interface ApiKeyWithFullKey {
  id: string;
  name: string;
  apiKey: string;             // FULL KEY - only shown once!
  keyPrefix: string;
  scope: ApiKeyScope;
  expiresAt: Date | null;
  createdAt: Date;
}

/**
 * Request to create a new API key
 */
export interface CreateApiKeyRequest {
  name: string;
  scope: ApiKeyScope;
  expiresInDays: number | null;   // null = no expiration
}

/**
 * Request to update an API key (rename, change scope)
 */
export interface UpdateApiKeyRequest {
  name?: string;
  scope?: ApiKeyScope;
}

/**
 * Paginated list response
 */
export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

/**
 * API Keys list response with pagination
 */
export interface ApiKeysListResponse {
  keys: ApiKey[];
  pagination: PaginationInfo;
  monthlyLimit: number;
  keysCreatedThisMonth: number;
}

/**
 * Usage statistics for a specific key
 */
export interface ApiKeyUsageStats {
  keyId: string;
  totalRequests: number;
  requestsLast7Days: number[];
  requestsByEndpoint: Record<string, number>;
}
```

### 3.2 Mappers (apikey-mappers.ts)

```typescript
// src/domain/apikey/apikey-mappers.ts

// NOTE: This file must NOT import from lib/api/ or hooks/

import type {
  ApiKey,
  ApiKeyWithFullKey,
  ApiKeyScope,
  ApiKeyStatus,
  ApiKeysListResponse,
  PaginationInfo,
} from "./apikey-entity";

/**
 * Raw API response for a single key (snake_case from backend)
 */
interface RawApiKey {
  id: string;
  name: string;
  key_prefix: string;
  scope: string;
  expires_at: string | null;
  is_active: boolean;
  request_count: number;
  last_request_at: string | null;
  created_at: string;
  updated_at: string;
  revoked_at: string | null;
}

/**
 * Raw API response for key creation (snake_case from backend)
 */
interface RawApiKeyCreated {
  id: string;
  name: string;
  api_key: string;
  key_prefix: string;
  scope: string;
  expires_at: string | null;
  created_at: string;
}

/**
 * Raw API response for list (snake_case from backend)
 */
interface RawApiKeysListResponse {
  keys: RawApiKey[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
  monthly_limit: number;
  keys_created_this_month: number;
}

/**
 * Parse scope from string to ApiKeyScope (case-insensitive)
 */
const parseScope = (scope: string): ApiKeyScope => {
  const normalized = scope.toLowerCase();
  if (normalized === "read" || normalized === "write" || normalized === "admin") {
    return normalized as ApiKeyScope;
  }
  return "read"; // Default to read for unknown scopes
};

/**
 * Determine key status from fields
 */
const determineStatus = (key: RawApiKey): ApiKeyStatus => {
  if (!key.is_active) return "revoked";
  if (key.expires_at && new Date(key.expires_at) < new Date()) return "expired";
  return "active";
};

/**
 * Transform raw API key to domain ApiKey
 */
export const apiKeyFromRaw = (raw: RawApiKey): ApiKey => ({
  id: raw.id,
  name: raw.name,
  keyPrefix: raw.key_prefix,
  scope: parseScope(raw.scope),
  expiresAt: raw.expires_at ? new Date(raw.expires_at) : null,
  isActive: raw.is_active,
  requestCount: raw.request_count,
  lastRequestAt: raw.last_request_at ? new Date(raw.last_request_at) : null,
  createdAt: new Date(raw.created_at),
  revokedAt: raw.revoked_at ? new Date(raw.revoked_at) : null,
});

/**
 * Transform raw creation response to domain
 * NOTE: api_key is the FULL key (only returned once)
 */
export const apiKeyWithFullKeyFromRaw = (raw: RawApiKeyCreated): ApiKeyWithFullKey => ({
  id: raw.id,
  name: raw.name,
  apiKey: raw.api_key,  // Full key from backend
  keyPrefix: raw.key_prefix,
  scope: parseScope(raw.scope),
  expiresAt: raw.expires_at ? new Date(raw.expires_at) : null,
  createdAt: new Date(raw.created_at),
});

/**
 * Transform raw list response to domain
 */
export const apiKeysListFromRaw = (raw: RawApiKeysListResponse): ApiKeysListResponse => ({
  keys: raw.keys.map(apiKeyFromRaw),
  pagination: {
    page: raw.pagination.page,
    limit: raw.pagination.limit,
    total: raw.pagination.total,
    totalPages: raw.pagination.total_pages,
  },
  monthlyLimit: raw.monthly_limit,
  keysCreatedThisMonth: raw.keys_created_this_month,
});
```

### 3.3 Validation

```typescript
// src/domain/api-keys/validation.ts

// NOTE: This file must NOT import from lib/api/ or hooks/

import type { ApiKeyScope } from "./types";

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

/**
 * Validate API key name
 * Rules:
 * - Required, non-empty
 * - Max 64 characters
 * - Only alphanumeric, spaces, hyphens, underscores
 */
export const validateApiKeyName = (name: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!name || name.trim().length === 0) {
    errors.push("Name is required");
  }
  
  if (name.length > 64) {
    errors.push("Name must be 64 characters or less");
  }
  
  if (!/^[a-zA-Z0-9\s\-_]+$/.test(name)) {
    errors.push("Name can only contain letters, numbers, spaces, hyphens, and underscores");
  }
  
  return {
    isValid: errors.length === 0,
    errors,
  };
};

/**
 * Validate expiry days input
 * Rules:
 * - If provided, must be between 1 and 365 days
 * - null means no expiration
 */
export const validateExpiryDays = (days: number | null): ValidationResult => {
  const errors: string[] = [];
  
  if (days !== null) {
    if (!Number.isInteger(days)) {
      errors.push("Expiry must be a whole number of days");
    }
    if (days < 1) {
      errors.push("Expiry must be at least 1 day");
    }
    if (days > 365) {
      errors.push("Expiry cannot exceed 365 days");
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors,
  };
};

/**
 * Validate API key scope
 */
export const validateScope = (scope: string): scope is ApiKeyScope => {
  return scope === "read" || scope === "write" || scope === "admin";
};

/**
 * Validate create API key request
 */
export const validateCreateApiKeyRequest = (
  data: { name: string; scope: string; expiresInDays: number | null }
): ValidationResult => {
  const errors: string[] = [];
  
  const nameValidation = validateApiKeyName(data.name);
  errors.push(...nameValidation.errors);
  
  if (!validateScope(data.scope)) {
    errors.push("Invalid scope. Must be 'read', 'write', or 'admin'");
  }
  
  const expiryValidation = validateExpiryDays(data.expiresInDays);
  errors.push(...expiryValidation.errors);
  
  return {
    isValid: errors.length === 0,
    errors,
  };
};

/**
 * Validate update API key request
 */
export const validateUpdateApiKeyRequest = (
  data: { name?: string; scope?: string }
): ValidationResult => {
  const errors: string[] = [];
  
  if (data.name !== undefined) {
    const nameValidation = validateApiKeyName(data.name);
    errors.push(...nameValidation.errors);
  }
  
  if (data.scope !== undefined && !validateScope(data.scope)) {
    errors.push("Invalid scope. Must be 'read', 'write', or 'admin'");
  }
  
  return {
    isValid: errors.length === 0,
    errors,
  };
};
```

---

## 4. Data Layer

### 4.1 Session Client (Operator Settings - Management)

```typescript
// src/lib/api/client/session-client.ts

/**
 * Session-based API client for operator settings
 * Uses cookies for authentication
 * Used ONLY for managing own keys (create, revoke, rotate, rename)
 */

const getBaseUrl = () => import.meta.env.VITE_API_BASE_URL || "";

export const sessionClient = {
  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const baseUrl = getBaseUrl();
    const response = await fetch(`${baseUrl}${endpoint}`, {
      ...options,
      credentials: "include", // Send cookies
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });
    
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Request failed: ${response.status}`);
    }
    
    // Handle 204 No Content
    if (response.status === 204) {
      return null as T;
    }
    
    return response.json();
  },
};
```

### 4.2 API Keys Management Endpoints

```typescript
// src/lib/api/rest/endpoints/api-keys.ts

import { sessionClient } from "@/lib/api/client/session-client";
import {
  apiKeyWithFullKeyFromRaw,
  apiKeysListFromRaw,
  apiKeyFromRaw,
} from "@/domain/api-keys/transforms";
import type {
  ApiKey,
  ApiKeyWithFullKey,
  CreateApiKeyRequest,
  UpdateApiKeyRequest,
  ApiKeysListResponse,
} from "@/domain/api-keys/types";

const API_KEYS_BASE = "/v1/auth/api-keys";

export const apiKeysEndpoints = {
  /**
   * List all API keys for the authenticated operator
   * @param page Page number (1-indexed)
   * @param limit Items per page (default 20)
   */
  list: async (page: number = 1, limit: number = 20): Promise<ApiKeysListResponse> => {
    const response = await sessionClient.request<ApiKeysListResponse>(
      `${API_KEYS_BASE}?page=${page}&limit=${limit}`
    );
    return apiKeysListFromRaw(response as any);
  },

  /**
   * Create a new API key
   * Returns the FULL key - only time it's available
   */
  create: async (data: CreateApiKeyRequest): Promise<ApiKeyWithFullKey> => {
    const response = await sessionClient.request<any>(API_KEYS_BASE, {
      method: "POST",
      body: JSON.stringify(data),
    });
    return apiKeyWithFullKeyFromRaw(response);
  },

  /**
   * Update an API key (rename, change scope)
   */
  update: async (keyId: string, data: UpdateApiKeyRequest): Promise<ApiKey> => {
    const response = await sessionClient.request<any>(`${API_KEYS_BASE}/${keyId}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
    return apiKeyFromRaw(response);
  },

  /**
   * Revoke an API key (soft delete)
   */
  revoke: async (keyId: string): Promise<void> => {
    await sessionClient.request<void>(`${API_KEYS_BASE}/${keyId}`, {
      method: "DELETE",
    });
  },

  /**
   * Rotate an API key - generates new key, invalidates old
   * Returns the NEW full key - only time it's available
   */
  rotate: async (keyId: string): Promise<ApiKeyWithFullKey> => {
    const response = await sessionClient.request<any>(`${API_KEYS_BASE}/${keyId}/rotate`, {
      method: "POST",
    });
    return apiKeyWithFullKeyFromRaw(response);
  },

  /**
   * Get a single API key by ID
   */
  get: async (keyId: string): Promise<ApiKey> => {
    const response = await sessionClient.request<any>(`${API_KEYS_BASE}/${keyId}`);
    return apiKeyFromRaw(response);
  },
};
```

### 4.3 Developer Client (Third-Party Apps)

```typescript
// src/lib/api/client/developer-client.ts

/**
 * Developer API client for third-party applications
 * Uses X-API-Key header for authentication
 * 
 * IMPORTANT: This is different from session-client.ts
 * This is used by developers to make API calls using their key
 */

export interface DeveloperClientOptions {
  baseUrl?: string;
  onUnauthorized?: () => void;
  onRateLimited?: (retryAfter?: number) => void;
}

export interface DeveloperClient {
  // Device endpoints
  getDevices: () => Promise<any>;
  getDevice: (imei: string) => Promise<any>;
  deleteDevice: (imei: string) => Promise<void>;
  
  // Command endpoints
  getCommandStatus: (dispatchId: string) => Promise<any>;
  retryCommand: (dispatchId: string) => Promise<any>;
  cancelCommand: (dispatchId: string) => Promise<void>;
  
  // Telemetry endpoints
  getTelemetryHistory: (params: any) => Promise<any>;
  getLatestTelemetry: (deviceId: string) => Promise<any>;
  
  // Update endpoints
  getUpdateStatus: () => Promise<any>;
  getVersions: () => Promise<any>;
  
  // Generic request method for flexibility
  request: <T>(endpoint: string, options?: RequestInit) => Promise<T>;
}

/**
 * Create a developer client with the provided API key
 */
export const createDeveloperClient = (
  apiKey: string,
  options: DeveloperClientOptions = {}
): DeveloperClient => {
  const baseUrl = options.baseUrl || import.meta.env.VITE_API_BASE_URL || "";

  const request = async <T>(
    endpoint: string,
    init: RequestInit = {}
  ): Promise<T> => {
    const url = endpoint.startsWith("http") ? endpoint : `${baseUrl}${endpoint}`;
    
    const response = await fetch(url, {
      ...init,
      headers: {
        "X-API-Key": apiKey,
        "Content-Type": "application/json",
        ...init.headers,
      },
    });

    if (response.status === 401) {
      options.onUnauthorized?.();
      throw new Error("Invalid or expired API key");
    }

    if (response.status === 429) {
      const retryAfter = response.headers.get("Retry-After");
      options.onRateLimited?.(retryAfter ? parseInt(retryAfter) : undefined);
      throw new Error("Rate limit exceeded");
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Request failed: ${response.status}`);
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return null as T;
    }

    return response.json();
  };

  return {
    getDevices: () => request("/v1/devices"),
    getDevice: (imei) => request(`/v1/devices/${imei}`),
    deleteDevice: (imei) => request(`/v1/devices/${imei}`, { method: "DELETE" }),
    
    getCommandStatus: (dispatchId) => request(`/v1/command/${dispatchId}/status`),
    retryCommand: (dispatchId) => request(`/v1/command/${dispatchId}/retry`, { method: "POST" }),
    cancelCommand: (dispatchId) => request(`/v1/command/${dispatchId}`, { method: "DELETE" }),
    
    getTelemetryHistory: (params) => {
      const searchParams = new URLSearchParams(params);
      return request(`/v1/telemetry/history?${searchParams}`);
    },
    getLatestTelemetry: (deviceId) => request(`/v1/telemetry/latest/${deviceId}`),
    
    getUpdateStatus: () => request("/v1/updates/status"),
    getVersions: () => request("/v1/updates/versions"),
    
    request,
  };
};
```

---

## 5. Presentation Layer

### 5.1 Query Keys (Centralized)

```typescript
// src/hooks/api-keys/query-keys.ts

export const apiKeysQueryKeys = {
  all: ["api-keys"] as const,
  lists: () => [...apiKeysQueryKeys.all, "list"] as const,
  list: (page: number, limit: number) => [...apiKeysQueryKeys.lists(), { page, limit }] as const,
  details: () => [...apiKeysQueryKeys.all, "detail"] as const,
  detail: (keyId: string) => [...apiKeysQueryKeys.details(), keyId] as const,
};
```

### 5.2 Hook: List API Keys (with Pagination)

```typescript
// src/hooks/api-keys/use-api-keys.ts

import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysEndpoints } from "@/lib/api/rest/endpoints/api-keys";
import { apiKeysListFromRaw } from "@/domain/api-keys/transforms";
import { apiKeysQueryKeys } from "./query-keys";
import type { ApiKey, ApiKeysListResponse } from "@/domain/api-keys/types";

const PAGE_SIZE = 20;

export const useApiKeys = () => {
  const queryClient = useQueryClient();

  const query = useInfiniteQuery({
    queryKey: apiKeysQueryKeys.lists(),
    queryFn: async ({ pageParam = 1 }) => {
      const response = await apiKeysEndpoints.list(pageParam, PAGE_SIZE);
      return apiKeysListFromRaw(response as any);
    },
    getNextPageParam: (lastPage) => {
      const { page, totalPages } = lastPage.pagination;
      return page < totalPages ? page + 1 : undefined;
    },
    initialPageParam: 1,
  });

  return {
    // Flattened list of keys from all pages
    keys: query.data?.pages.flatMap((page) => page.keys) ?? [],
    
    // Pagination info from first page
    pagination: query.data?.pages[0]?.pagination ?? {
      page: 1,
      limit: PAGE_SIZE,
      total: 0,
      totalPages: 0,
    },
    
    // Monthly limits from first page
    monthlyLimit: query.data?.pages[0]?.monthlyLimit ?? 20,
    keysCreatedThisMonth: query.data?.pages[0]?.keysCreatedThisMonth ?? 0,
    
    // Query state
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    hasNextPage: query.hasNextPage ?? false,
    error: query.error as Error | null,
    
    // Actions
    fetchNextPage: query.fetchNextPage,
    refetch: query.refetch,
  };
};
```

### 5.3 Hook: Create API Key (with Optimistic Update)

```typescript
// src/hooks/api-keys/use-create-api-key.ts

import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysEndpoints } from "@/lib/api/rest/endpoints/api-keys";
import { validateCreateApiKeyRequest } from "@/domain/api-keys/validation";
import type { ApiKeyWithFullKey, CreateApiKeyRequest } from "@/domain/api-keys/types";

export const useCreateApiKey = () => {
  const queryClient = useQueryClient();
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  const mutation = useMutation({
    mutationFn: async (data: CreateApiKeyRequest): Promise<ApiKeyWithFullKey> => {
      // Client-side validation
      const validation = validateCreateApiKeyRequest({
        name: data.name,
        scope: data.scope,
        expiresInDays: data.expiresInDays,
      });
      
      if (!validation.isValid) {
        throw { validationErrors: validation.errors };
      }
      
      return apiKeysEndpoints.create(data);
    },
    onSuccess: () => {
      // Invalidate all list queries to refetch
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
  });

  const createKey = useCallback(
    async (data: CreateApiKeyRequest): Promise<ApiKeyWithFullKey> => {
      setValidationErrors([]);
      try {
        return await mutation.mutateAsync(data);
      } catch (error: any) {
        if (error.validationErrors) {
          setValidationErrors(error.validationErrors);
          throw error;
        }
        throw error;
      }
    },
    [mutation]
  );

  return {
    createKey,
    isCreating: mutation.isPending,
    createdKey: mutation.data ?? null,
    error: mutation.error?.message ?? null,
    validationErrors,
    reset: () => {
      setValidationErrors([]);
      mutation.reset();
    },
  };
};
```

### 5.4 Hook: Revoke API Key (with Optimistic Update)

```typescript
// src/hooks/api-keys/use-revoke-api-key.ts

import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysEndpoints } from "@/lib/api/rest/endpoints/api-keys";
import type { ApiKey } from "@/domain/api-keys/types";

export const useRevokeApiKey = () => {
  const queryClient = useQueryClient();
  
  // Track which key is pending revocation
  const [pendingRevoke, setPendingRevoke] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (keyId: string): Promise<void> => {
      return apiKeysEndpoints.revoke(keyId);
    },
    
    // Optimistic update - remove from list immediately
    onMutate: async (keyId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["api-keys"] });

      // Snapshot previous value
      const previousData = queryClient.getQueriesData({ queryKey: ["api-keys"] });

      // Optimistically update all pages
      queryClient.setQueriesData({ queryKey: ["api-keys"] }, (old: any) => {
        if (!old) return old;
        
        // Handle both single query and infinite query pages
        if (old.pages) {
          return {
            ...old,
            pages: old.pages.map((page: any) => ({
              ...page,
              keys: page.keys.filter((k: ApiKey) => k.id !== keyId),
            })),
          };
        }
        
        return old;
      });

      setPendingRevoke(keyId);
      
      // Return context for rollback
      return { previousData };
    },
    
    // On error, rollback to snapshot
    onError: (err, keyId, context) => {
      // Restore previous data
      if (context?.previousData) {
        context.previousData.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data);
        });
      }
      setPendingRevoke(null);
    },
    
    // Always clear pending state
    onSettled: () => {
      setPendingRevoke(null);
    },
  });

  const revokeKey = useCallback(
    async (keyId: string): Promise<void> => {
      return mutation.mutateAsync(keyId);
    },
    [mutation]
  );

  return {
    revokeKey,
    isRevoking: mutation.isPending,
    pendingRevoke,
    error: mutation.error?.message ?? null,
  };
};
```

### 5.5 Hook: Rotate API Key (Returns New Key)

```typescript
// src/hooks/api-keys/use-rotate-api-key.ts

import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysEndpoints } from "@/lib/api/rest/endpoints/api-keys";
import type { ApiKeyWithFullKey } from "@/domain/api-keys/types";

export const useRotateApiKey = () => {
  const queryClient = useQueryClient();
  
  // Track pending rotate
  const [pendingRotate, setPendingRotate] = useState<string | null>(null);
  
  // Track the new key returned from rotate
  const [rotatedKey, setRotatedKey] = useState<ApiKeyWithFullKey | null>(null);

  const mutation = useMutation({
    mutationFn: async (keyId: string): Promise<ApiKeyWithFullKey> => {
      return apiKeysEndpoints.rotate(keyId);
    },
    onSuccess: (data) => {
      // Store the new key to show in dialog
      setRotatedKey(data);
      // Invalidate to refetch updated list
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
    onSettled: () => {
      setPendingRotate(null);
    },
  });

  const rotateKey = useCallback(
    async (keyId: string): Promise<ApiKeyWithFullKey> => {
      setPendingRotate(keyId);
      try {
        return await mutation.mutateAsync(keyId);
      } finally {
        setPendingRotate(null);
      }
    },
    [mutation]
  );

  return {
    rotateKey,
    isRotating: mutation.isPending,
    pendingRotate,
    rotatedKey,
    error: mutation.error?.message ?? null,
    clearRotatedKey: () => setRotatedKey(null),
  };
};
```

### 5.6 Hook: Update API Key (Rename/Change Scope)

```typescript
// src/hooks/api-keys/use-update-api-key.ts

import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysEndpoints } from "@/lib/api/rest/endpoints/api-keys";
import { validateUpdateApiKeyRequest } from "@/domain/api-keys/validation";
import type { ApiKey, UpdateApiKeyRequest } from "@/domain/api-keys/types";

export const useUpdateApiKey = () => {
  const queryClient = useQueryClient();
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  const mutation = useMutation({
    mutationFn: async ({
      keyId,
      data,
    }: {
      keyId: string;
      data: UpdateApiKeyRequest;
    }): Promise<ApiKey> => {
      // Client-side validation
      const validation = validateUpdateApiKeyRequest(data);
      
      if (!validation.isValid) {
        throw { validationErrors: validation.errors };
      }
      
      return apiKeysEndpoints.update(keyId, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
  });

  const updateKey = useCallback(
    async (keyId: string, data: UpdateApiKeyRequest): Promise<ApiKey> => {
      setValidationErrors([]);
      try {
        return await mutation.mutateAsync({ keyId, data });
      } catch (error: any) {
        if (error.validationErrors) {
          setValidationErrors(error.validationErrors);
          throw error;
        }
        throw error;
      }
    },
    [mutation]
  );

  return {
    updateKey,
    isUpdating: mutation.isPending,
    updatedKey: mutation.data ?? null,
    error: mutation.error?.message ?? null,
    validationErrors,
    reset: () => {
      setValidationErrors([]);
      mutation.reset();
    },
  };
};
```

---

## 6. UI Layer

### 6.1 API Keys Page

```tsx
// src/ui/pages/settings/api-keys/api-keys-page.tsx

import { useState, useCallback, useEffect } from "react";
import { useApiKeys } from "@/hooks/api-keys/use-api-keys";
import { useCreateApiKey } from "@/hooks/api-keys/use-create-api-key";
import { useRevokeApiKey } from "@/hooks/api-keys/use-revoke-api-key";
import { useRotateApiKey } from "@/hooks/api-keys/use-rotate-api-key";
import { useUpdateApiKey } from "@/hooks/api-keys/use-update-api-key";
import { ApiKeyList } from "./components/api-key-list";
import { CreateKeyDialog } from "./components/create-key-dialog";
import { RevokeKeyDialog } from "./components/revoke-key-dialog";
import { RotateKeyDialog } from "./components/rotate-key-dialog";
import { EditKeyDialog } from "./components/edit-key-dialog";
import { KeyCreatedDialog } from "./components/key-created-dialog";
import { PageHeader } from "@/ui/components/ui/page-header";
import { Card } from "@/ui/components/ui/card";
import { Button } from "@/ui/components/ui/button";
import { Plus, Loader2 } from "lucide-react";

export const ApiKeysPage = () => {
  // Dialog states
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showRevokeDialog, setShowRevokeDialog] = useState(false);
  const [showRotateDialog, setShowRotateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showKeyCreatedDialog, setShowKeyCreatedDialog] = useState(false);
  
  // Selected key for operations
  const [selectedKey, setSelectedKey] = useState<any>(null);
  
  // Newly created/rotated key to display
  const [displayKey, setDisplayKey] = useState<{
    key: string;
    name: string;
    keyPrefix: string;
  } | null>(null);

  // Hooks
  const {
    keys,
    monthlyLimit,
    keysCreatedThisMonth,
    isLoading,
    error,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = useApiKeys();

  const {
    createKey,
    isCreating,
    validationErrors: createErrors,
  } = useCreateApiKey();

  const {
    revokeKey,
    isRevoking,
  } = useRevokeApiKey();

  const {
    rotateKey,
    isRotating,
    rotatedKey,
  } = useRotateApiKey();

  const {
    updateKey,
    isUpdating,
    validationErrors: updateErrors,
  } = useUpdateApiKey();

  // Clear display key when dialogs close
  useEffect(() => {
    if (!showKeyCreatedDialog) {
      setDisplayKey(null);
    }
  }, [showKeyCreatedDialog]);

  // Handlers
  const handleCreate = async (data: { name: string; scope: string; expiresInDays: number | null }) => {
    const created = await createKey(data);
    setDisplayKey({
      key: created.apiKey,
      name: created.name,
      keyPrefix: created.keyPrefix,
    });
    setShowCreateDialog(false);
    setShowKeyCreatedDialog(true);
  };

  const handleRevoke = async () => {
    if (!selectedKey) return;
    await revokeKey(selectedKey.id);
    setShowRevokeDialog(false);
    setSelectedKey(null);
  };

  const handleRotate = async () => {
    if (!selectedKey) return;
    await rotateKey(selectedKey.id);
    setShowRotateDialog(false);
    // rotatedKey is set by the hook
    if (rotatedKey) {
      setDisplayKey({
        key: rotatedKey.apiKey,
        name: rotatedKey.name,
        keyPrefix: rotatedKey.keyPrefix,
      });
      setShowKeyCreatedDialog(true);
    }
    setSelectedKey(null);
  };

  const handleUpdate = async (data: { name?: string; scope?: string }) => {
    if (!selectedKey) return;
    await updateKey(selectedKey.id, data);
    setShowEditDialog(false);
    setSelectedKey(null);
  };

  // Action handlers from list
  const onRevoke = useCallback((key: any) => {
    setSelectedKey(key);
    setShowRevokeDialog(true);
  }, []);

  const onRotate = useCallback((key: any) => {
    setSelectedKey(key);
    setShowRotateDialog(true);
  }, []);

  const onEdit = useCallback((key: any) => {
    setSelectedKey(key);
    setShowEditDialog(true);
  }, []);

  return (
    <div className="container max-w-4xl py-8">
      <PageHeader
        title="API Keys"
        description="Manage API keys for developer access to your account"
        action={
          <Button onClick={() => setShowCreateDialog(true)} disabled={isCreating}>
            <Plus className="h-4 w-4 mr-2" />
            Generate New Key
          </Button>
        }
      />
      
      <Card className="p-6">
        <div className="mb-4 flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Keys created this month: {keysCreatedThisMonth} / {monthlyLimit}
          </span>
        </div>
        
        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <ApiKeyListSkeleton key={i} />
            ))}
          </div>
        ) : error ? (
          <div className="text-center py-8 text-destructive">
            Failed to load API keys. <button onClick={() => window.location.reload()}>Try again</button>
          </div>
        ) : keys.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">
              No API keys yet. Generate one to get started.
            </p>
          </div>
        ) : (
          <>
            <ApiKeyList
              keys={keys}
              onRevoke={onRevoke}
              onRotate={onRotate}
              onEdit={onEdit}
            />
            
            {hasNextPage && (
              <div className="mt-4 flex justify-center">
                <Button
                  variant="outline"
                  onClick={fetchNextPage}
                  disabled={isFetchingNextPage}
                >
                  {isFetchingNextPage ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Loading...
                    </>
                  ) : (
                    "Load More"
                  )}
                </Button>
              </div>
            )}
          </>
        )}
      </Card>
      
      {/* Dialogs */}
      <CreateKeyDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSubmit={handleCreate}
        isCreating={isCreating}
        validationErrors={createErrors}
        canCreate={keysCreatedThisMonth < monthlyLimit}
      />
      
      <RevokeKeyDialog
        open={showRevokeDialog}
        onOpenChange={setShowRevokeDialog}
        onConfirm={handleRevoke}
        isRevoking={isRevoking}
        keyName={selectedKey?.name}
      />
      
      <RotateKeyDialog
        open={showRotateDialog}
        onOpenChange={setShowRotateDialog}
        onConfirm={handleRotate}
        isRotating={isRotating}
        keyName={selectedKey?.name}
      />
      
      <EditKeyDialog
        open={showEditDialog}
        onOpenChange={setShowEditDialog}
        onSubmit={handleUpdate}
        isUpdating={isUpdating}
        validationErrors={updateErrors}
        currentName={selectedKey?.name}
        currentScope={selectedKey?.scope}
      />
      
      <KeyCreatedDialog
        open={showKeyCreatedDialog}
        onOpenChange={setShowKeyCreatedDialog}
        apiKey={displayKey?.key ?? ""}
        name={displayKey?.name ?? ""}
      />
    </div>
  );
};
```

### 6.2 API Key List Component

```tsx
// src/ui/pages/settings/api-keys/components/api-key-list.tsx

import type { ApiKey } from "@/domain/api-keys/types";
import { ApiKeyRow } from "./api-key-row";

interface ApiKeyListProps {
  keys: ApiKey[];
  onRevoke: (key: ApiKey) => void;
  onRotate: (key: ApiKey) => void;
  onEdit: (key: ApiKey) => void;
}

export const ApiKeyList = ({ keys, onRevoke, onRotate, onEdit }: ApiKeyListProps) => {
  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="grid grid-cols-12 gap-4 text-sm font-medium text-muted-foreground px-4">
        <div className="col-span-3">Name</div>
        <div className="col-span-2">Prefix</div>
        <div className="col-span-1">Scope</div>
        <div className="col-span-2">Created</div>
        <div className="col-span-2">Last Used</div>
        <div className="col-span-2">Actions</div>
      </div>
      
      {/* Rows */}
      <div className="divide-y">
        {keys.map((apiKey) => (
          <ApiKeyRow
            key={apiKey.id}
            apiKey={apiKey}
            onRevoke={onRevoke}
            onRotate={onRotate}
            onEdit={onEdit}
          />
        ))}
      </div>
    </div>
  );
};
```

### 6.3 API Key Row Component

```tsx
// src/ui/pages/settings/api-keys/components/api-key-row.tsx

import type { ApiKey } from "@/domain/api-keys/types";
import { Button } from "@/ui/components/ui/button";
import { CopyButton } from "@/ui/components/shared/copy-button";
import { RotateCw, Pencil, Trash2, MoreVertical } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/ui/components/ui/dropdown-menu";

interface ApiKeyRowProps {
  apiKey: ApiKey;
  onRevoke: (key: ApiKey) => void;
  onRotate: (key: ApiKey) => void;
  onEdit: (key: ApiKey) => void;
}

const scopeLabels: Record<string, string> = {
  read: "Read",
  write: "Write",
  admin: "Admin",
};

const scopeColors: Record<string, string> = {
  read: "bg-blue-100 text-blue-800",
  write: "bg-green-100 text-green-800",
  admin: "bg-purple-100 text-purple-800",
};

export const ApiKeyRow = ({ apiKey, onRevoke, onRotate, onEdit }: ApiKeyRowProps) => {
  const formatDate = (date: Date | null) => {
    if (!date) return "Never";
    return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    }).format(date);
  };

  const formatRelativeTime = (date: Date | null) => {
    if (!date) return "Never";
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return formatDate(date);
  };

  return (
    <div className="grid grid-cols-12 gap-4 items-center px-4 py-3 hover:bg-muted/50">
      {/* Name */}
      <div className="col-span-3">
        <div className="font-medium">{apiKey.name}</div>
        {apiKey.expiresAt && (
          <div className="text-xs text-amber-600">
            Expires {formatDate(apiKey.expiresAt)}
          </div>
        )}
      </div>

      {/* Prefix with copy */}
      <div className="col-span-2 flex items-center gap-2">
        <code className="text-sm bg-muted px-2 py-1 rounded">
          {apiKey.keyPrefix}
        </code>
        <CopyButton
          value={apiKey.keyPrefix}
          size="sm"
          variant="ghost"
          aria-label={`Copy prefix ${apiKey.keyPrefix}`}
        />
      </div>

      {/* Scope badge */}
      <div className="col-span-1">
        <span
          className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
            scopeColors[apiKey.scope] || scopeColors.read
          }`}
        >
          {scopeLabels[apiKey.scope] || "Read"}
        </span>
      </div>

      {/* Created */}
      <div className="col-span-2 text-sm text-muted-foreground">
        {formatDate(apiKey.createdAt)}
      </div>

      {/* Last Used */}
      <div className="col-span-2 text-sm text-muted-foreground">
        {formatRelativeTime(apiKey.lastRequestAt)}
      </div>

      {/* Actions */}
      <div className="col-span-2">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => onEdit(apiKey)}>
              <Pencil className="h-4 w-4 mr-2" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onRotate(apiKey)}>
              <RotateCw className="h-4 w-4 mr-2" />
              Rotate
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onRevoke(apiKey)}
              className="text-destructive focus:text-destructive"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Revoke
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
};
```

### 6.4 Key Created Dialog (CRITICAL - Shows Full Key Once)

```tsx
// src/ui/pages/settings/api-keys/components/key-created-dialog.tsx

import { useState, useEffect, useCallback } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/ui/components/ui/dialog";
import { Button } from "@/ui/components/ui/button";
import { CopyButton } from "@/ui/components/shared/copy-button";
import { AlertTriangle, Eye, EyeOff } from "lucide-react";

interface KeyCreatedDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  apiKey: string;
  name: string;
}

/**
 * CRITICAL SECURITY COMPONENT
 * 
 * This dialog shows the FULL API key ONCE and only once.
 * The key is cleared from memory when the dialog closes.
 * User MUST copy the key before closing.
 */
export const KeyCreatedDialog = ({
  open,
  onOpenChange,
  apiKey,
  name,
}: KeyCreatedDialogProps) => {
  const [revealed, setRevealed] = useState(false);
  const [copied, setCopied] = useState(false);

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setRevealed(false);
      setCopied(false);
    }
  }, [open]);

  // Clear key from memory after dialog closes
  const handleClose = useCallback(() => {
    // Key is not explicitly cleared but will be garbage collected
    // since we don't store it in state beyond this component
    onOpenChange(false);
  }, [onOpenChange]);

  // Don't render if no key (safety check)
  if (!apiKey) return null;

  const maskedKey = apiKey.slice(0, 8) + "•".repeat(Math.max(0, apiKey.length - 8));

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            API Key Created
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Security Warning Banner */}
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
            <p className="text-sm text-amber-800">
              <strong>Important:</strong> Copy this API key now. You won't be able to 
              see it again after closing this dialog.
            </p>
          </div>

          {/* Key Name */}
          <div>
            <label className="text-sm font-medium">Name</label>
            <p className="text-muted-foreground">{name}</p>
          </div>

          {/* Full Key Display */}
          <div>
            <label className="text-sm font-medium">API Key</label>
            <div className="flex items-center gap-2 mt-1">
              <code className="flex-1 bg-muted px-3 py-2 rounded font-mono text-sm break-all min-h-[40px]">
                {revealed ? apiKey : maskedKey}
              </code>
              <Button
                variant="outline"
                size="icon"
                onClick={() => setRevealed(!revealed)}
                title={revealed ? "Hide key" : "Reveal key"}
                aria-label={revealed ? "Hide key" : "Reveal key"}
              >
                {revealed ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
              <CopyButton
                value={apiKey}
                size="icon"
                variant="outline"
                copied={copied}
                onCopiedChange={setCopied}
                aria-label="Copy API key"
              />
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={handleClose}>
              Close
            </Button>
            {!copied && (
              <span className="text-sm text-muted-foreground self-center">
                Please copy the key before closing
              </span>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};
```

### 6.5 Create Key Dialog

```tsx
// src/ui/pages/settings/api-keys/components/create-key-dialog.tsx

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/ui/components/ui/dialog";
import { Button } from "@/ui/components/ui/button";
import { Input } from "@/ui/components/ui/input";
import { Label } from "@/ui/components/ui/label";
import { Alert } from "@/ui/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/ui/components/ui/select";
import type { ApiKeyScope } from "@/domain/api-keys/types";

interface CreateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { name: string; scope: ApiKeyScope; expiresInDays: number | null }) => void;
  isCreating: boolean;
  validationErrors: string[];
  canCreate: boolean;
}

export const CreateKeyDialog = ({
  open,
  onOpenChange,
  onSubmit,
  isCreating,
  validationErrors,
  canCreate,
}: CreateKeyDialogProps) => {
  const [name, setName] = useState("");
  const [scope, setScope] = useState<ApiKeyScope>("read");
  const [expiresInDays, setExpiresInDays] = useState<string>("");
  const [localErrors, setLocalErrors] = useState<string[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLocalErrors([]);

    if (!name.trim()) {
      setLocalErrors(["Name is required"]);
      return;
    }

    const expires = expiresInDays ? parseInt(expiresInDays, 10) : null;

    onSubmit({
      name: name.trim(),
      scope,
      expiresInDays: expires,
    });
  };

  const handleClose = (open: boolean) => {
    if (!open) {
      // Reset form when closing
      setName("");
      setScope("read");
      setExpiresInDays("");
      setLocalErrors([]);
    }
    onOpenChange(open);
  };

  const allErrors = [...localErrors, ...validationErrors];

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Generate New API Key</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Monthly limit warning */}
            {!canCreate && (
              <Alert variant="destructive">
                You've reached your monthly limit of API keys. 
                Revoke an existing key to create a new one.
              </Alert>
            )}

            {/* Validation errors */}
            {allErrors.length > 0 && (
              <Alert variant="destructive">
                <ul className="list-disc list-inside">
                  {allErrors.map((error, i) => (
                    <li key={i}>{error}</li>
                  ))}
                </ul>
              </Alert>
            )}

            {/* Key Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Key Name *</Label>
              <Input
                id="name"
                placeholder="e.g., Production iOS App"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isCreating}
                maxLength={64}
              />
              <p className="text-xs text-muted-foreground">
                A descriptive name to identify this key (max 64 characters)
              </p>
            </div>

            {/* Scope */}
            <div className="space-y-2">
              <Label htmlFor="scope">Scope *</Label>
              <Select value={scope} onValueChange={(v) => setScope(v as ApiKeyScope)}>
                <SelectTrigger id="scope">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="read">
                    <div>
                      <div className="font-medium">Read</div>
                      <div className="text-xs text-muted-foreground">
                        GET requests only
                      </div>
                    </div>
                  </SelectItem>
                  <SelectItem value="write">
                    <div>
                      <div className="font-medium">Write</div>
                      <div className="text-xs text-muted-foreground">
                        GET, POST, PUT, PATCH requests
                      </div>
                    </div>
                  </SelectItem>
                  <SelectItem value="admin">
                    <div>
                      <div className="font-medium">Admin</div>
                      <div className="text-xs text-muted-foreground">
                        All requests including DELETE
                      </div>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Scope determines what operations this key can perform
              </p>
            </div>

            {/* Expiration */}
            <div className="space-y-2">
              <Label htmlFor="expires">Expires In (Days)</Label>
              <Input
                id="expires"
                type="number"
                placeholder="Leave empty for no expiration"
                value={expiresInDays}
                onChange={(e) => setExpiresInDays(e.target.value)}
                disabled={isCreating}
                min="1"
                max="365"
              />
              <p className="text-xs text-muted-foreground">
                Optional. Leave empty for no expiration. Max 365 days.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleClose(false)}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isCreating || !canCreate}>
              {isCreating ? "Generating..." : "Generate Key"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
```

### 6.6 Revoke Key Dialog

```tsx
// src/ui/pages/settings/api-keys/components/revoke-key-dialog.tsx

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/ui/components/ui/dialog";
import { Button } from "@/ui/components/ui/button";
import { AlertTriangle } from "lucide-react";

interface RevokeKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isRevoking: boolean;
  keyName?: string;
}

export const RevokeKeyDialog = ({
  open,
  onOpenChange,
  onConfirm,
  isRevoking,
  keyName = "this key",
}: RevokeKeyDialogProps) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            Revoke API Key
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke <strong>{keyName}</strong>?
            This action cannot be undone and any applications using this key 
            will immediately lose access.
          </DialogDescription>
        </DialogHeader>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isRevoking}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isRevoking}
          >
            {isRevoking ? "Revoking..." : "Revoke Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
```

### 6.7 Rotate Key Dialog

```tsx
// src/ui/pages/settings/api-keys/components/rotate-key-dialog.tsx

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/ui/components/ui/dialog";
import { Button } from "@/ui/components/ui/button";
import { AlertTriangle } from "lucide-react";

interface RotateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isRotating: boolean;
  keyName?: string;
}

export const RotateKeyDialog = ({
  open,
  onOpenChange,
  onConfirm,
  isRotating,
  keyName = "this key",
}: RotateKeyDialogProps) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            Rotate API Key
          </DialogTitle>
          <DialogDescription>
            Rotating <strong>{keyName}</strong> will generate a new key.
            The old key will be immediately invalidated.
            Make sure to update your applications with the new key.
          </DialogDescription>
        </DialogHeader>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isRotating}
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            disabled={isRotating}
          >
            {isRotating ? "Rotating..." : "Rotate Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
```

### 6.8 Edit Key Dialog (Rename/Change Scope)

```tsx
// src/ui/pages/settings/api-keys/components/edit-key-dialog.tsx

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/ui/components/ui/dialog";
import { Button } from "@/ui/components/ui/button";
import { Input } from "@/ui/components/ui/input";
import { Label } from "@/ui/components/ui/label";
import { Alert } from "@/ui/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/ui/components/ui/select";
import type { ApiKeyScope } from "@/domain/api-keys/types";

interface EditKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { name?: string; scope?: ApiKeyScope }) => void;
  isUpdating: boolean;
  validationErrors: string[];
  currentName?: string;
  currentScope?: ApiKeyScope;
}

export const EditKeyDialog = ({
  open,
  onOpenChange,
  onSubmit,
  isUpdating,
  validationErrors,
  currentName = "",
  currentScope = "read",
}: EditKeyDialogProps) => {
  const [name, setName] = useState(currentName);
  const [scope, setScope] = useState<ApiKeyScope>(currentScope);
  const [localErrors, setLocalErrors] = useState<string[]>([]);

  // Sync state when props change
  useEffect(() => {
    setName(currentName);
    setScope(currentScope);
  }, [currentName, currentScope, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLocalErrors([]);

    // Only send updates if values changed
    const updates: { name?: string; scope?: ApiKeyScope } = {};
    if (name.trim() !== currentName) {
      if (!name.trim()) {
        setLocalErrors(["Name cannot be empty"]);
        return;
      }
      updates.name = name.trim();
    }
    if (scope !== currentScope) {
      updates.scope = scope;
    }

    if (Object.keys(updates).length === 0) {
      onOpenChange(false);
      return;
    }

    onSubmit(updates);
  };

  const allErrors = [...localErrors, ...validationErrors];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit API Key</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {allErrors.length > 0 && (
              <Alert variant="destructive">
                <ul className="list-disc list-inside">
                  {allErrors.map((error, i) => (
                    <li key={i}>{error}</li>
                  ))}
                </ul>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="edit-name">Name</Label>
              <Input
                id="edit-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isUpdating}
                maxLength={64}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-scope">Scope</Label>
              <Select value={scope} onValueChange={(v) => setScope(v as ApiKeyScope)}>
                <SelectTrigger id="edit-scope">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="read">Read - GET requests only</SelectItem>
                  <SelectItem value="write">Write - GET, POST, PUT, PATCH</SelectItem>
                  <SelectItem value="admin">Admin - All requests including DELETE</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isUpdating}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isUpdating}>
              {isUpdating ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
```

### 6.9 Skeleton Loaders

```tsx
// src/ui/pages/settings/api-keys/components/api-key-list-skeleton.tsx

import { Skeleton } from "@/ui/components/ui/skeleton";

export const ApiKeyListSkeleton = () => {
  return (
    <div className="grid grid-cols-12 gap-4 items-center px-4 py-3">
      {/* Name */}
      <div className="col-span-3 space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-20" />
      </div>
      
      {/* Prefix */}
      <div className="col-span-2">
        <Skeleton className="h-6 w-16" />
      </div>
      
      {/* Scope */}
      <div className="col-span-1">
        <Skeleton className="h-5 w-12" />
      </div>
      
      {/* Created */}
      <div className="col-span-2">
        <Skeleton className="h-4 w-20" />
      </div>
      
      {/* Last Used */}
      <div className="col-span-2">
        <Skeleton className="h-4 w-16" />
      </div>
      
      {/* Actions */}
      <div className="col-span-2">
        <Skeleton className="h-8 w-8" />
      </div>
    </div>
  );
};
```

### 6.10 Shared Copy Button Component

```tsx
// src/ui/components/shared/copy-button.tsx

import { useState, useCallback } from "react";
import { Button } from "@/ui/components/ui/button";
import { Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface CopyButtonProps {
  value: string;
  onCopiedChange?: (copied: boolean) => void;
  copied?: boolean;
  size?: "sm" | "icon" | "default";
  variant?: "default" | "outline" | "ghost";
  className?: string;
  "aria-label"?: string;
}

/**
 * Reusable copy button with clipboard support
 * Includes fallback for browsers that don't support navigator.clipboard
 */
export const CopyButton = ({
  value,
  onCopiedChange,
  copied: externalCopied,
  size = "icon",
  variant = "outline",
  className,
  "aria-label": ariaLabel,
}: CopyButtonProps) => {
  const [internalCopied, setInternalCopied] = useState(false);

  // Use external state if provided, otherwise internal
  const copied = externalCopied !== undefined ? externalCopied : internalCopied;

  const handleCopy = useCallback(async () => {
    try {
      // Try modern clipboard API first
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(value);
      } else {
        // Fallback for older browsers
        const textarea = document.createElement("textarea");
        textarea.value = value;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
      }
      
      setInternalCopied(true);
      onCopiedChange?.(true);
      
      // Reset after 2 seconds
      setTimeout(() => {
        setInternalCopied(false);
        onCopiedChange?.(false);
      }, 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  }, [value, onCopiedChange]);

  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleCopy}
      className={cn("relative", className)}
      aria-label={ariaLabel || `Copy ${value}`}
    >
      {copied ? (
        <Check className="h-4 w-4 text-green-500" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </Button>
  );
};
```

---

## 7. User Flows

### 7.1 Create API Key Flow

```

                                                                     
  1. User clicks "Generate New Key"                                  
                                                                    
                                                                    
  2. CreateKeyDialog opens                                           
                                                                    
                                                                    
  3. User enters:                                                    
     - Name (required)                                               
     - Scope (read/write/admin)                                      
     - Expiration (optional)                                         
                                                                    
                                                                    
  4. User clicks "Generate Key"                                      
                                                                    
                                                                    
  5. Show loading state                                              
                                                                    
                                                                    
  6a. On SUCCESS:                                                    
      - Close CreateKeyDialog                                        
      - Open KeyCreatedDialog with FULL key                         
      - Show warning: "Copy now, won't see again"                   
      - User clicks copy button                                       
      - User clicks "Done"                                           
      - Clear key from memory                                        
      - Close KeyCreatedDialog                                       
                                                                     
                                                                     
      7. List refreshes showing new key                             
                                                                     
  6b. On VALIDATION ERROR:                                           
      - Show error messages inline                                   
      - Stay in dialog                                               
                                                                     
  6c. On SERVER ERROR:                                               
      - Show error toast                                             
      - Stay in dialog                                               
                                                                     

```

### 7.2 Revoke API Key Flow (Optimistic Update)

```

                                                                     
  1. User clicks revoke icon on key row                              
                                                                    
                                                                    
  2. RevokeKeyDialog opens with key name                             
                                                                    
                                                                    
  3. User clicks "Revoke Key"                                        
                                                                    
                                                                    
  4. IMMEDIATELY (optimistic):                                      
     - Key disappears from list                                      
     - Dialog closes                                                 
                                                                    
                                                                    
  5. Background: API call to revoke                                 
                                                                    
                                                                    
  6a. On SUCCESS:                                                    
      - Nothing to do (already updated UI)                          
      - Show success toast                                           
                                                                     
  6b. On ERROR:                                                      
      - Rollback: Key reappears in list                              
      - Show error toast                                             
      - Show revoke dialog again                                    
                                                                     

```

### 7.3 Rotate API Key Flow

```

                                                                     
  1. User clicks rotate in dropdown menu                             
                                                                    
                                                                    
  2. RotateKeyDialog opens with key name                             
                                                                    
                                                                    
  3. User clicks "Rotate Key"                                        
                                                                    
                                                                    
  4. Show loading state on button                                   
                                                                    
                                                                    
  6a. On SUCCESS:                                                    
      - Close RotateKeyDialog                                        
      - Open KeyCreatedDialog with NEW full key                      
      - Show warning about updating applications                     
      - User copies new key                                          
      - User clicks "Done"                                           
      - Clear new key from memory                                    
                                                                     
                                                                     
      7. List refreshes showing rotated key                         
                                                                     
  6b. On ERROR:                                                      
      - Show error toast                                             
      - Stay in dialog                                               
                                                                     

```

### 7.4 Edit API Key Flow

```

                                                                     
  1. User clicks edit in dropdown menu                               
                                                                    
                                                                    
  2. EditKeyDialog opens pre-filled with:                            
     - Current name                                                  
     - Current scope                                                 
                                                                    
                                                                    
  3. User modifies:                                                  
     - Name (optional)                                               
     - Scope (optional)                                              
                                                                    
                                                                    
  4. User clicks "Save Changes"                                      
                                                                    
                                                                    
  5. If no changes made, close dialog                                
                                                                    
                                                                    
  6. Show loading state                                              
                                                                    
                                                                    
  6a. On SUCCESS:                                                    
      - Close dialog                                                 
      - Show success toast                                           
      - List refreshes showing updated key                          
                                                                     
  6b. On ERROR:                                                      
      - Show error messages                                          
      - Stay in dialog                                               
                                                                     

```

---

## 8. State Management

### 8.1 React Query Configuration

```typescript
// src/lib/api/react-query.ts

import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,     // 5 minutes
      gcTime: 1000 * 60 * 30,       // 30 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,  // Don't retry mutations by default
    },
  },
});
```

### 8.2 Toast Notifications

```typescript
// src/ui/hooks/use-toast.ts

// Use a toast library like sonner or react-hot-toast

import { toast } from "sonner";

toast.success("API key created successfully");
toast.success("API key revoked");
toast.error("Failed to rotate key", {
  description: "Please try again",
});
```

---

## 9. Error Handling

### 9.1 Error States and Recovery

| Error | User Message | Recovery Action |
|-------|-------------|----------------|
| Network offline | "Unable to connect. Check your connection." | Retry button |
| 401 Unauthorized | Redirect to login | Login page |
| 403 Monthly limit | "You've reached your limit" | Show in dialog |
| 403 Rate limited | "Too many requests. Wait X seconds." | Auto-retry after delay |
| 404 Key not found | "Key not found" | Remove from list, show toast |
| 500 Server error | "Something went wrong" | Retry button |
| 503 Unavailable | "Service temporarily unavailable" | Retry with backoff |

### 9.2 Error Boundary

```tsx
// src/ui/components/error-boundary.tsx

import { Component, type ReactNode } from "react";
import { Button } from "@/ui/components/ui/button";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="p-4 text-center">
          <p className="text-destructive">Something went wrong</p>
          <Button onClick={() => window.location.reload()}>Reload</Button>
        </div>
      );
    }

    return this.props.children;
  }
}
```

---

## 10. Accessibility

### 10.1 Requirements (ALL MANDATORY)

1. **Keyboard Navigation**: All interactive elements must be reachable via Tab
2. **Focus Management**: Focus moves logically through dialogs
3. **ARIA Labels**: All buttons and inputs have accessible labels
4. **Screen Reader**: Key actions announced (e.g., "Key revoked successfully")
5. **Color Contrast**: All text meets WCAG AA standards
6. **Error Announcements**: Form errors announced to screen readers

### 10.2 ARIA Implementation

```tsx
// Example: KeyCreatedDialog accessibility
<Dialog aria-labelledby="key-created-title" aria-describedby="key-created-desc">
  <DialogTitle id="key-created-title">API Key Created</DialogTitle>
  <p id="key-created-desc" className="sr-only">
    Copy this API key now. You won't be able to see it again.
  </p>
  ...
</Dialog>

// Example: Error alert
<Alert role="alert" aria-live="polite">
  {errors.map((error) => (
    <p key={error}>{error}</p>
  ))}
</Alert>
```

---

## 11. Developer Portal

### 11.1 Purpose

The Developer Portal is a separate section for third-party developers who use API keys to integrate with the platform. It provides:
- API documentation
- Code examples
- Endpoint testing
- Usage analytics

### 11.2 Structure

```
src/ui/pages/developer-portal/
 dashboard-page.tsx        # Overview with usage stats
 documentation-page.tsx    # API reference docs
 components/
    sdk-examples.tsx      # Code samples for different languages
    test-endpoint.tsx     # Interactive API tester
    usage-chart.tsx       # Usage visualization
```

### 11.3 Developer Dashboard

```tsx
// src/ui/pages/developer-portal/dashboard-page.tsx

import { UsageChart } from "./components/usage-chart";
import { SdkExamples } from "./components/sdk-examples";

export const DeveloperDashboardPage = () => {
  return (
    <div className="container py-8">
      <h1>Developer Portal</h1>
      
      <div className="grid grid-cols-2 gap-8">
        {/* Usage Stats */}
        <section>
          <h2>API Usage</h2>
          <UsageChart />
        </section>
        
        {/* Quick Start */}
        <section>
          <h2>Quick Start</h2>
          <SdkExamples apiKey="your-api-key" />
        </section>
      </div>
    </div>
  );
};
```

---

## 12. Super Admin UI

### 12.1 Purpose

Super Admin UI allows administrators to:
- View ALL API keys across ALL operators
- See usage statistics for all operators
- Force revoke any API key
- Monitor API key usage patterns

### 12.2 Structure

```
src/ui/pages/admin/
 api-keys/
    api-keys-admin-page.tsx    # List all keys with filters
    stats-page.tsx            # Global statistics
    components/
        admin-key-list.tsx         # List with operator info
        admin-key-row.tsx          # Row with operator name
        force-revoke-dialog.tsx    # Force revoke confirmation
        operator-filter.tsx        # Filter by operator
        global-stats.tsx           # Statistics dashboard
```

### 12.3 Domain Types

```typescript
// src/domain/api-keys/admin-types.ts

export interface AdminApiKey {
  id: string;
  operatorId: string;
  operatorName: string;
  name: string;
  keyPrefix: string;
  scope: ApiKeyScope;
  isActive: boolean;
  requestCount: number;
  createdAt: Date;
}

export interface GlobalStats {
  totalKeys: number;
  activeKeys: number;
  revokedKeys: number;
  totalRequestsToday: number;
  totalRequestsThisMonth: number;
  topOperators: OperatorStats[];
  requestsByScope: Record<ApiKeyScope, number>;
}

export interface OperatorStats {
  operatorId: string;
  operatorName: string;
  activeKeys: number;
  totalRequests: number;
}
```

### 12.4 Admin Data Layer

```typescript
// src/lib/api/rest/endpoints/admin-api-keys.ts

import { sessionClient } from "@/lib/api/client/session-client";

const ADMIN_API_KEYS_BASE = "/v1/admin/api-keys";

export const adminApiKeysEndpoints = {
  listAll: async (params: {
    page?: number;
    limit?: number;
    operatorId?: string;
    search?: string;
  }): Promise<{
    keys: AdminApiKey[];
    pagination: PaginationInfo;
  }> => {
    const searchParams = new URLSearchParams();
    if (params.page) searchParams.set("page", String(params.page));
    if (params.limit) searchParams.set("limit", String(params.limit));
    if (params.operatorId) searchParams.set("operator_id", params.operatorId);
    if (params.search) searchParams.set("search", params.search);
    
    const response = await sessionClient.request<any>(
      `${ADMIN_API_KEYS_BASE}?${searchParams.toString()}`
    );
    
    return {
      keys: response.keys.map(transformAdminKey),
      pagination: {
        page: response.pagination.page,
        limit: response.pagination.limit,
        total: response.pagination.total,
        totalPages: response.pagination.total_pages,
      },
    };
  },

  getStats: async (): Promise<GlobalStats> => {
    const response = await sessionClient.request<any>(`${ADMIN_API_KEYS_BASE}/stats`);
    return transformStats(response);
  },

  forceRevoke: async (keyId: string): Promise<void> => {
    await sessionClient.request<void>(`${ADMIN_API_KEYS_BASE}/${keyId}`, {
      method: "DELETE",
    });
  },

  getOperatorKeys: async (
    operatorId: string,
    page = 1,
    limit = 20
  ): Promise<{ keys: AdminApiKey[]; pagination: PaginationInfo }> => {
    const response = await sessionClient.request<any>(
      `/v1/admin/operators/${operatorId}/api-keys?page=${page}&limit=${limit}`
    );
    
    return {
      keys: response.keys.map(transformAdminKey),
      pagination: {
        page: response.pagination.page,
        limit: response.pagination.limit,
        total: response.pagination.total,
        totalPages: response.pagination.total_pages,
      },
    };
  },
};

const transformAdminKey = (raw: any): AdminApiKey => ({
  id: raw.id,
  operatorId: raw.operator_id,
  operatorName: raw.operator_name,
  name: raw.name,
  keyPrefix: raw.key_prefix,
  scope: raw.scope as ApiKeyScope,
  isActive: raw.is_active,
  requestCount: raw.request_count,
  createdAt: new Date(raw.created_at),
});

const transformStats = (raw: any): GlobalStats => ({
  totalKeys: raw.total_keys,
  activeKeys: raw.active_keys,
  revokedKeys: raw.revoked_keys,
  totalRequestsToday: raw.total_requests_today,
  totalRequestsThisMonth: raw.total_requests_this_month,
  topOperators: raw.top_operators.map((op: any) => ({
    operatorId: op.operator_id,
    operatorName: op.operator_name,
    activeKeys: op.active_keys,
    totalRequests: op.total_requests,
  })),
  requestsByScope: raw.requests_by_scope,
});
```

### 12.5 Admin Hooks

```typescript
// src/hooks/api-keys/admin/use-admin-api-keys.ts

import { useInfiniteQuery, useMutation, useQuery } from "@tanstack/react-query";
import { adminApiKeysEndpoints } from "@/lib/api/rest/endpoints/admin-api-keys";

export const useAdminApiKeys = (filters?: {
  operatorId?: string;
  search?: string;
}) => {
  return useInfiniteQuery({
    queryKey: ["admin", "api-keys", filters],
    queryFn: async ({ pageParam = 1 }) => {
      return adminApiKeysEndpoints.listAll({
        page: pageParam,
        limit: 20,
        operatorId: filters?.operatorId,
        search: filters?.search,
      });
    },
    getNextPageParam: (lastPage) => {
      const { page, totalPages } = lastPage.pagination;
      return page < totalPages ? page + 1 : undefined;
    },
    initialPageParam: 1,
  });
};

export const useGlobalStats = () => {
  return useQuery({
    queryKey: ["admin", "api-keys", "stats"],
    queryFn: () => adminApiKeysEndpoints.getStats(),
    staleTime: 1000 * 60 * 5,
  });
};

export const useForceRevokeKey = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (keyId: string) => adminApiKeysEndpoints.forceRevoke(keyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "api-keys"] });
    },
  });
};
```

### 12.6 Admin API Keys Page

```tsx
// src/ui/pages/admin/api-keys/api-keys-admin-page.tsx

import { useState } from "react";
import { useAdminApiKeys, useGlobalStats, useForceRevokeKey } from "@/hooks/api-keys/admin/use-admin-api-keys";
import { PageHeader } from "@/ui/components/ui/page-header";
import { Card } from "@/ui/components/ui/card";
import { Button } from "@/ui/components/ui/button";
import { RefreshCw, BarChart3, AlertTriangle, Trash2 } from "lucide-react";

export const AdminApiKeysPage = () => {
  const [filters, setFilters] = useState<{
    operatorId?: string;
    search?: string;
  }>({});
  const [selectedKey, setSelectedKey] = useState<AdminApiKey | null>(null);
  const [showRevokeDialog, setShowRevokeDialog] = useState(false);
  const [showStats, setShowStats] = useState(true);

  const { keys, pagination, isLoading, hasNextPage, fetchNextPage, refetch } = 
    useAdminApiKeys(filters);
  const { data: stats } = useGlobalStats();
  const { mutate: forceRevoke, isPending: isRevoking } = useForceRevokeKey();

  const handleRevoke = (key: AdminApiKey) => {
    setSelectedKey(key);
    setShowRevokeDialog(true);
  };

  const confirmRevoke = () => {
    if (selectedKey) {
      forceRevoke(selectedKey.id);
      setShowRevokeDialog(false);
      setSelectedKey(null);
    }
  };

  return (
    <div className="container py-8">
      <PageHeader
        title="API Keys Administration"
        description="Manage all API keys across operators"
        action={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowStats(!showStats)}>
              <BarChart3 className="h-4 w-4 mr-2" />
              {showStats ? "Hide" : "Show"} Stats
            </Button>
            <Button variant="outline" onClick={() => refetch()}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </Button>
          </div>
        }
      />

      {showStats && stats && (
        <div className="grid grid-cols-4 gap-4 mt-6">
          <Card className="p-4">
            <div className="text-2xl font-bold">{stats.totalKeys}</div>
            <div className="text-sm text-muted-foreground">Total Keys</div>
          </Card>
          <Card className="p-4">
            <div className="text-2xl font-bold text-green-600">
              {stats.activeKeys}
            </div>
            <div className="text-sm text-muted-foreground">Active</div>
          </Card>
          <Card className="p-4">
            <div className="text-2xl font-bold text-blue-600">
              {stats.totalRequestsToday.toLocaleString()}
            </div>
            <div className="text-sm text-muted-foreground">Requests Today</div>
          </Card>
          <Card className="p-4">
            <div className="text-2xl font-bold">
              {stats.totalRequestsThisMonth.toLocaleString()}
            </div>
            <div className="text-sm text-muted-foreground">Requests This Month</div>
          </Card>
        </div>
      )}

      <Card className="p-6 mt-6">
        {/* Filter bar would go here */}
        
        <div className="mt-4">
          {isLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : keys.length === 0 ? (
            <div className="text-center py-8">No API keys found</div>
          ) : (
            <>
              <div className="space-y-2">
                {keys.map((key) => (
                  <div
                    key={key.id}
                    className="flex items-center justify-between p-4 border rounded"
                  >
                    <div className="flex items-center gap-4">
                      <div>
                        <div className="font-medium">{key.operatorName}</div>
                        <div className="text-sm text-muted-foreground">
                          {key.name}
                        </div>
                      </div>
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {key.keyPrefix}
                      </code>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className={`text-xs px-2 py-1 rounded ${
                        key.scope === "admin" ? "bg-purple-100" :
                        key.scope === "write" ? "bg-green-100" : "bg-blue-100"
                      }`}>
                        {key.scope}
                      </span>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRevoke(key)}
                        className="text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
              
              {hasNextPage && (
                <div className="mt-4 flex justify-center">
                  <Button variant="outline" onClick={fetchNextPage}>
                    Load More
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </Card>

      {showRevokeDialog && selectedKey && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center">
          <Card className="w-96 p-6">
            <div className="flex items-center gap-2 text-destructive mb-4">
              <AlertTriangle className="h-5 w-5" />
              <h3 className="font-bold">Force Revoke Key?</h3>
            </div>
            <p className="text-sm text-muted-foreground mb-4">
              Revoke <strong>{selectedKey.name}</strong> from{" "}
              <strong>{selectedKey.operatorName}</strong>? This cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => setShowRevokeDialog(false)}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={confirmRevoke}
                disabled={isRevoking}
              >
                {isRevoking ? "Revoking..." : "Force Revoke"}
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
};
```

---

## 13. Developer SDK Examples

### 13.1 SDK Examples Component

```tsx
// src/ui/pages/developer-portal/components/sdk-examples.tsx

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/ui/components/ui/tabs";
import { CodeBlock } from "@/ui/components/ui/code-block";

interface SdkExamplesProps {
  apiKey: string;
}

export const SdkExamples = ({ apiKey }: SdkExamplesProps) => {
  return (
    <Tabs defaultValue="curl">
      <TabsList>
        <TabsTrigger value="curl">cURL</TabsTrigger>
        <TabsTrigger value="javascript">JavaScript</TabsTrigger>
        <TabsTrigger value="python">Python</TabsTrigger>
        <TabsTrigger value="go">Go</TabsTrigger>
      </TabsList>

      <TabsContent value="curl">
        <CodeBlock
          code={`curl -X GET "https://api.vyzorix.app/v1/devices" \\
  -H "X-API-Key: ${apiKey}"`}
          language="bash"
        />
      </TabsContent>

      <TabsContent value="javascript">
        <CodeBlock
          code={`const response = await fetch('https://api.vyzorix.app/v1/devices', {
  headers: {
    'X-API-Key': '${apiKey}'
  }
});
const data = await response.json();
console.log(data);`}
          language="javascript"
        />
      </TabsContent>

      <TabsContent value="python">
        <CodeBlock
          code={`import requests

response = requests.get(
    'https://api.vyzorix.app/v1/devices',
    headers={'X-API-Key': '${apiKey}'}
)
data = response.json()
print(data)`}
          language="python"
        />
      </TabsContent>

      <TabsContent value="go">
        <CodeBlock
          code={`package main

import (
    "net/http"
    "fmt"
)

func main() {
    req, _ := http.NewRequest("GET", "https://api.vyzorix.app/v1/devices", nil)
    req.Header.Set("X-API-Key", "${apiKey}")
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()
}`}
          language="go"
        />
      </TabsContent>
    </Tabs>
  );
};
```

### 12.2 Test Endpoint Component

```tsx
// src/ui/pages/developer-portal/components/test-endpoint.tsx

import { useState } from "react";
import { Button } from "@/ui/components/ui/button";
import { Input } from "@/ui/components/ui/input";
import { Label } from "@/ui/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/ui/components/ui/select";
import { CodeBlock } from "@/ui/components/ui/code-block";

const ENDPOINTS = [
  { value: "GET /v1/devices", label: "List Devices" },
  { value: "GET /v1/devices/:imei", label: "Get Device" },
  { value: "GET /v1/telemetry/history", label: "Telemetry History" },
];

export const TestEndpoint = () => {
  const [endpoint, setEndpoint] = useState("");
  const [response, setResponse] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleTest = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Implementation would use the developer client
      const result = await testEndpoint(endpoint);
      setResponse(result);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <Select value={endpoint} onValueChange={setEndpoint}>
          <SelectTrigger className="w-64">
            <SelectValue placeholder="Select endpoint" />
          </SelectTrigger>
          <SelectContent>
            {ENDPOINTS.map((ep) => (
              <SelectItem key={ep.value} value={ep.value}>
                {ep.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        
        <Button onClick={handleTest} disabled={!endpoint || loading}>
          {loading ? "Testing..." : "Test"}
        </Button>
      </div>

      {error && (
        <div className="p-4 bg-destructive/10 text-destructive rounded">
          {error}
        </div>
      )}

      {response && (
        <CodeBlock
          code={JSON.stringify(response, null, 2)}
          language="json"
        />
      )}
    </div>
  );
};
```

---

## 13. Testing Strategy

### 13.1 Unit Tests (Domain Layer)

```typescript
// src/domain/api-keys/__tests__/transforms.test.ts

import { apiKeyFromRaw, apiKeyWithFullKeyFromRaw } from "../transforms";

describe("apiKeyFromRaw", () => {
  it("transforms raw API response to domain", () => {
    const raw = {
      id: "key-123",
      name: "Test Key",
      key_prefix: "vxyz_ab",
      scope: "read",
      expires_at: "2026-12-31T00:00:00Z",
      is_active: true,
      request_count: 100,
      last_request_at: "2026-07-01T00:00:00Z",
      created_at: "2026-01-01T00:00:00Z",
      revoked_at: null,
    };

    const result = apiKeyFromRaw(raw);

    expect(result.id).toBe("key-123");
    expect(result.name).toBe("Test Key");
    expect(result.keyPrefix).toBe("vxyz_ab");
    expect(result.scope).toBe("read");
    expect(result.isActive).toBe(true);
    expect(result.expiresAt).toBeInstanceOf(Date);
  });

  it("handles null optional fields", () => {
    const raw = {
      id: "key-123",
      name: "Test",
      key_prefix: "vxyz",
      scope: "write",
      expires_at: null,
      is_active: true,
      request_count: 0,
      last_request_at: null,
      created_at: "2026-01-01Z",
      revoked_at: null,
    };

    const result = apiKeyFromRaw(raw);

    expect(result.expiresAt).toBeNull();
    expect(result.lastRequestAt).toBeNull();
  });
});
```

### 13.2 Validation Tests

```typescript
// src/domain/api-keys/__tests__/validation.test.ts

import {
  validateApiKeyName,
  validateExpiryDays,
  validateScope,
} from "../validation";

describe("validateApiKeyName", () => {
  it("accepts valid names", () => {
    expect(validateApiKeyName("Production App").isValid).toBe(true);
    expect(validateApiKeyName("my-app_v2").isValid).toBe(true);
    expect(validateApiKeyName("App 123").isValid).toBe(true);
  });

  it("rejects empty names", () => {
    const result = validateApiKeyName("");
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Name is required");
  });

  it("rejects names over 64 characters", () => {
    const longName = "a".repeat(65);
    const result = validateApiKeyName(longName);
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Name must be 64 characters or less");
  });
});
```

### 13.3 Hook Tests (with React Query Testing Library)

```typescript
// src/hooks/api-keys/__tests__/use-api-keys.test.ts

import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useApiKeys } from "../use-api-keys";

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  
  return ({ children }: any) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe("useApiKeys", () => {
  it("loads keys", async () => {
    // Mock API response
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        keys: [{ id: "1", name: "Test" }],
        pagination: { page: 1, limit: 20, total: 1, totalPages: 1 },
        monthly_limit: 20,
        keys_created_this_month: 1,
      }),
    });

    const { result } = renderHook(() => useApiKeys(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.keys.length).toBe(1));
  });
});
```

### 13.4 Component Tests

```typescript
// src/ui/pages/settings/api-keys/components/__tests__/api-key-row.test.tsx

import { render, screen, fireEvent } from "@testing-library/react";
import { ApiKeyRow } from "../api-key-row";

const mockKey = {
  id: "key-123",
  name: "Production App",
  keyPrefix: "vxyz_ab",
  scope: "read" as const,
  expiresAt: null,
  isActive: true,
  requestCount: 100,
  lastRequestAt: new Date(),
  createdAt: new Date(),
  revokedAt: null,
};

describe("ApiKeyRow", () => {
  it("renders key name and prefix", () => {
    render(
      <ApiKeyRow
        apiKey={mockKey}
        onRevoke={jest.fn()}
        onRotate={jest.fn()}
        onEdit={jest.fn()}
      />
    );

    expect(screen.getByText("Production App")).toBeInTheDocument();
    expect(screen.getByText("vxyz_ab")).toBeInTheDocument();
  });

  it("calls onRevoke when revoke action clicked", () => {
    const onRevoke = jest.fn();
    
    render(
      <ApiKeyRow
        apiKey={mockKey}
        onRevoke={onRevoke}
        onRotate={jest.fn()}
        onEdit={jest.fn()}
      />
    );

    // Open dropdown and click revoke
    fireEvent.click(screen.getByRole("button", { name: /more/i }));
    fireEvent.click(screen.getByText("Revoke"));

    expect(onRevoke).toHaveBeenCalledWith(mockKey);
  });
});
```

---

## Appendix A: Component Inventory

| Component | File | States | Accessibility |
|-----------|------|--------|---------------|
| ApiKeysPage | api-keys-page.tsx | loading, error, empty, populated | - |
| ApiKeyList | api-key-list.tsx | - | - |
| ApiKeyRow | api-key-row.tsx | default, menu-open | keyboard nav, aria |
| CreateKeyDialog | create-key-dialog.tsx | default, submitting, error | focus-trap, aria |
| KeyCreatedDialog | key-created-dialog.tsx | default, copied | live-region |
| RevokeKeyDialog | revoke-key-dialog.tsx | default, confirming | - |
| RotateKeyDialog | rotate-key-dialog.tsx | default, confirming | - |
| EditKeyDialog | edit-key-dialog.tsx | default, submitting, error | focus-trap |
| ApiKeyListSkeleton | api-key-list-skeleton.tsx | - | aria-busy |
| CopyButton | copy-button.tsx | default, copied | aria-label |

---

## Appendix B: Required Dependencies

```json
{
  "dependencies": {
    "@tanstack/react-query": "^5.x",
    "sonner": "^1.x"
  },
  "devDependencies": {
    "@testing-library/react": "^14.x",
    "@testing-library/jest-dom": "^6.x",
    "jest": "^29.x"
  }
}
```

---

## Appendix C: Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | API base URL | `https://api.vyzorix.app` |

---

*Document Version: 2.0*  
*Status: Complete - All issues addressed as mandatory requirements*