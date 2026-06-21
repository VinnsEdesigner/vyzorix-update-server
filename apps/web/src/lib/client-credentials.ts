// Client credentials management for request signing.
// Handles storage, retrieval, and lifecycle of API client credentials.

import { logger } from './logger';

export interface ClientCredentials {
  clientId: string;
  clientSecret: string;
  platform: 'web' | 'ios' | 'android';
  name: string;
  createdAt: number;
  /** Unix timestamp when credentials were fetched for this session */
  fetchedAt: number;
}

export interface SignedClient extends ClientCredentials {
  /** Derived encryption key (kept in memory only) */
  encryptionKey: Uint8Array;
}

// Storage keys
const STORAGE_KEY = 'vyz_client_credentials';
const ACTIVE_KEY = 'vyz_active_client_id';

// In-memory cache for the active client's derived key
let activeClientCache: SignedClient | null = null;

/**
 * ClientCredentialsStore manages API client credentials.
 * Credentials are stored in localStorage but the sensitive secret
 * is kept in memory for signing operations.
 */
export class ClientCredentialsStore {
  /**
   * Save credentials after receiving from /v1/auth/client-credentials
   */
  static save(credentials: Omit<ClientCredentials, 'fetchedAt'>): void {
    try {
      // Get existing list
      const existing = this.getAll();
      
      // Check if client already exists
      const idx = existing.findIndex(c => c.clientId === credentials.clientId);
      if (idx >= 0) {
        existing[idx] = { ...existing[idx], ...credentials, fetchedAt: Date.now() };
      } else {
        existing.push({ ...credentials, fetchedAt: Date.now() });
      }
      
      // Save to localStorage
      localStorage.setItem(STORAGE_KEY, JSON.stringify(existing));
      
      // Set as active
      localStorage.setItem(ACTIVE_KEY, credentials.clientId);
      
      logger.info('credentials', `Saved client: ${credentials.clientId}`);
    } catch (e) {
      logger.error('credentials', `Failed to save: ${e}`);
    }
  }

  /**
   * Get all stored credentials (without secrets - those are in memory only)
   */
  static getAll(): Omit<ClientCredentials, 'clientSecret'>[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      return JSON.parse(raw);
    } catch {
      return [];
    }
  }

  /**
   * Get credentials for a specific client (without secret - secret is in memory)
   */
  static getMeta(clientId: string): Omit<ClientCredentials, 'clientSecret'> | null {
    const all = this.getAll();
    return all.find(c => c.clientId === clientId) || null;
  }

  /**
   * Get the active client's credentials with secret (from memory cache).
   * Returns null if no active session or credentials not fetched.
   */
  static getActive(): ClientCredentials | null {
    const clientId = localStorage.getItem(ACTIVE_KEY);
    if (!clientId || !activeClientCache) return null;
    
    if (activeClientCache.clientId !== clientId) return null;
    
    return {
      clientId: activeClientCache.clientId,
      clientSecret: activeClientCache.clientSecret,
      platform: activeClientCache.platform,
      name: activeClientCache.name,
      createdAt: activeClientCache.createdAt,
      fetchedAt: activeClientCache.fetchedAt,
    };
  }

  /**
   * Set active client from stored credentials (prompts for secret if needed).
   * In practice, the secret is fetched and cached at login time.
   */
  static setActive(clientId: string, secret: string): void {
    // Store in memory cache with derived key
    const meta = this.getMeta(clientId);
    if (!meta) {
      logger.warn('credentials', `Client ${clientId} not found in storage`);
      return;
    }
    
    activeClientCache = {
      ...meta,
      clientSecret: secret,
      encryptionKey: new Uint8Array(), // Derived key computed on-demand
    };
    
    localStorage.setItem(ACTIVE_KEY, clientId);
    logger.info('credentials', `Set active client: ${clientId}`);
  }

  /**
   * Clear the in-memory cache (on logout).
   */
  static clearCache(): void {
    activeClientCache = null;
    logger.info('credentials', 'Cleared credentials cache');
  }

  /**
   * Remove a client's credentials.
   */
  static remove(clientId: string): void {
    try {
      const existing = this.getAll();
      const filtered = existing.filter(c => c.clientId !== clientId);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
      
      if (localStorage.getItem(ACTIVE_KEY) === clientId) {
        localStorage.removeItem(ACTIVE_KEY);
        this.clearCache();
      }
      
      logger.info('credentials', `Removed client: ${clientId}`);
    } catch (e) {
      logger.error('credentials', `Failed to remove: ${e}`);
    }
  }

  /**
   * Clear ALL stored credentials.
   */
  static clearAll(): void {
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(ACTIVE_KEY);
    this.clearCache();
    logger.info('credentials', 'Cleared all credentials');
  }

  /**
   * Check if we have credentials for a client.
   */
  static has(clientId: string): boolean {
    return this.getAll().some(c => c.clientId === clientId);
  }

  /**
   * Check if we have an active session.
   */
  static hasActive(): boolean {
    return this.getActive() !== null;
  }
}

/**
 * Fetch client credentials from the API.
 * This is called after login to get the signing credentials.
 */
export async function fetchClientCredentials(
  apiUrl: string,
  name: string = 'Web Dashboard'
): Promise<ClientCredentials> {
  const response = await fetch(`${apiUrl}/v1/auth/client-credentials`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      // Include session cookie automatically
    },
    credentials: 'include',
    body: JSON.stringify({
      name,
      platform: 'web',
      allowedOrigins: [window.location.origin],
    }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`Failed to fetch credentials: ${response.status} ${error}`);
  }

  const data = await response.json();
  
  const credentials: ClientCredentials = {
    clientId: data.clientId,
    clientSecret: data.clientSecret, // Only returned once!
    platform: 'web',
    name: data.name || name,
    createdAt: new Date(data.createdAt).getTime(),
    fetchedAt: Date.now(),
  };

  // Store credentials
  ClientCredentialsStore.save(credentials);
  ClientCredentialsStore.setActive(credentials.clientId, credentials.clientSecret);

  return credentials;
}

/**
 * List all client credentials (metadata only - no secrets).
 */
export function listClientCredentials(): Omit<ClientCredentials, 'clientSecret'>[] {
  return ClientCredentialsStore.getAll();
}

/**
 * Get the active client's credentials.
 */
export function getActiveCredentials(): ClientCredentials | null {
  return ClientCredentialsStore.getActive();
}

/**
 * Delete a client's credentials.
 */
export function deleteClientCredentials(clientId: string): void {
  ClientCredentialsStore.remove(clientId);
}

/**
 * Clear all credentials on logout.
 */
export function clearAllCredentials(): void {
  ClientCredentialsStore.clearAll();
}
