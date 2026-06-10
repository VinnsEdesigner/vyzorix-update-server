// @vyzorix/config/auth - Authentication services
import { z } from "zod";

/**
 * Authentication Provider Configuration
 */
export interface AuthProviderConfig {
  /**
   * Provider name
   */
  name: "google" | "github" | "microsoft" | "email";
  
  /**
   * Client ID for OAuth
   */
  clientId?: string;
  
  /**
   * Redirect URI
   */
  redirectUri?: string;
  
  /**
   * Additional scopes
   */
  scopes?: string[];
}

/**
 * Authentication Client Configuration
 */
export interface AuthClientConfig {
  /**
   * Backend API base URL
   */
  backendUrl: string;
  
  /**
   * Frontend callback URL
   */
  callbackUrl: string;
  
  /**
   * Auth providers
   */
  providers?: AuthProviderConfig[];
  
  /**
   * Token storage
   */
  storage?: "localStorage" | "sessionStorage" | "memory";
  
  /**
   * Token key for storage
   */
  tokenKey?: string;
}

/**
 * User Session
 */
export interface UserSession {
  token: string;
  expiresAt: number;
  user: {
    id: string;
    email: string;
    name: string;
    role: string;
  };
}

/**
 * Authentication Error
 */
export class AuthError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = "AuthError";
  }
}

/**
 * Create Authentication Client
 * Replaces @lovable.dev/cloud-auth-js functionality
 */
export function createAuthClient(config: AuthClientConfig) {
  const { backendUrl, callbackUrl, providers = [], storage = "localStorage", tokenKey = "vyzorix_token" } = config;

  // Validate configuration
  const configSchema = z.object({
    backendUrl: z.string().url(),
    callbackUrl: z.string().url(),
    providers: z.array(
      z.object({
        name: z.enum(["google", "github", "microsoft", "email"]),
        clientId: z.string().optional(),
        redirectUri: z.string().url().optional(),
        scopes: z.array(z.string()).optional(),
      })
    ).optional(),
    storage: z.enum(["localStorage", "sessionStorage", "memory"]).default("localStorage"),
    tokenKey: z.string().default("vyzorix_token"),
  });

  const validatedConfig = configSchema.parse(config);

  // Storage management
  const storageAdapter = {
    get: () => {
      try {
        if (validatedConfig.storage === "localStorage") {
          return localStorage.getItem(validatedConfig.tokenKey);
        } else if (validatedConfig.storage === "sessionStorage") {
          return sessionStorage.getItem(validatedConfig.tokenKey);
        }
        return null;
      } catch {
        return null;
      }
    },
    set: (value: string) => {
      try {
        if (validatedConfig.storage === "localStorage") {
          localStorage.setItem(validatedConfig.tokenKey, value);
        } else if (validatedConfig.storage === "sessionStorage") {
          sessionStorage.setItem(validatedConfig.tokenKey, value);
        }
      } catch (e) {
        console.error("Failed to store token:", e);
      }
    },
    remove: () => {
      try {
        if (validatedConfig.storage === "localStorage") {
          localStorage.removeItem(validatedConfig.tokenKey);
        } else if (validatedConfig.storage === "sessionStorage") {
          sessionStorage.removeItem(validatedConfig.tokenKey);
        }
      } catch (e) {
        console.error("Failed to remove token:", e);
      }
    },
  };

  // Get current session
  const getSession = (): UserSession | null => {
    const token = storageAdapter.get();
    if (!token) return null;
    
    try {
      // In a real implementation, you would decode and validate the JWT here
      // For now, we'll return a basic session object
      return {
        token,
        expiresAt: Date.now() + 3600000, // 1 hour from now
        user: {
          id: "current-user",
          email: "user@example.com",
          name: "Current User",
          role: "operator",
        },
      };
    } catch (e) {
      storageAdapter.remove();
      return null;
    }
  };

  // Sign in with OAuth provider
  const signInWithOAuth = async (providerName: string, options: { state?: string } = {}) => {
    const provider = providers.find(p => p.name === providerName);
    if (!provider) {
      throw new AuthError(`Provider ${providerName} not configured`, "provider_not_found");
    }

    // Build OAuth URL
    const authUrl = new URL(`${validatedConfig.backendUrl}/v1/auth/${providerName}`);
    
    // Add state parameter for CSRF protection
    if (options.state) {
      authUrl.searchParams.set("state", options.state);
    }

    // Redirect to auth URL
    window.location.href = authUrl.toString();
  };

  // Handle OAuth callback
  const handleCallback = (searchParams: URLSearchParams): UserSession | null => {
    const token = searchParams.get("token");
    const error = searchParams.get("error");
    
    if (error) {
      throw new AuthError(error, "oauth_error");
    }
    
    if (!token) {
      throw new AuthError("No token in callback", "missing_token");
    }
    
    // Store the token
    storageAdapter.set(token);
    
    // Return session
    return getSession();
  };

  // Sign out
  const signOut = async () => {
    try {
      const session = getSession();
      if (session) {
        // Call backend logout endpoint
        await fetch(`${validatedConfig.backendUrl}/v1/auth/logout`, {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${session.token}`,
            "Content-Type": "application/json",
          },
        });
      }
    } catch (e) {
      console.error("Logout error:", e);
    } finally {
      // Always clear local storage
      storageAdapter.remove();
    }
  };

  return {
    getSession,
    signInWithOAuth,
    handleCallback,
    signOut,
    providers: validatedConfig.providers,
    config: validatedConfig,
  };
}