// @vyzorix/config/env - Environment Management
import { z } from "zod";

/**
 * Environment Variable Schema
 */
export const envSchema = z.object({
  // Core variables
  NODE_ENV: z.enum(["development", "production", "test"]).default("development"),
  PORT: z.coerce.number().default(3000),
  
  // API Configuration
  VITE_API_BASE_URL: z.string().url().default("http://localhost:3000"),
  VITE_GOOGLE_CLIENT_ID: z.string().optional(),
  VITE_GOOGLE_CLIENT_SECRET: z.string().optional(),
  
  // Authentication
  VITE_AUTH_REDIRECT_URI: z.string().url().default("http://localhost:5173/auth/callback"),
  VITE_JWT_SECRET: z.string().min(32).optional(),
  
  // Feature Flags
  VITE_ENABLE_ANALYTICS: z.coerce.boolean().default(false),
  VITE_ENABLE_ERROR_REPORTING: z.coerce.boolean().default(false),
  VITE_ENABLE_EXPERIMENTAL_FEATURES: z.coerce.boolean().default(false),
  
  // UI Configuration
  VITE_DEFAULT_THEME: z.enum(["light", "dark", "system"]).default("system"),
  VITE_PRIMARY_COLOR: z.string().default("oklch(0.645 0.246 16.439)"), // Vyzorix brand color
  
  // Build Configuration
  VITE_BUILD_DATE: z.string().default(new Date().toISOString()),
  VITE_COMMIT_HASH: z.string().optional(),
  VITE_VERSION: z.string().default("1.0.0"),
});

/**
 * Environment Variables Type
 */
export type EnvVariables = z.infer<typeof envSchema>;

/**
 * Load and validate environment variables
 */
export function loadEnv(overrides: Partial<EnvVariables> = {}): EnvVariables {
  // Get environment variables from process.env
  const processEnv = {
    NODE_ENV: process.env.NODE_ENV,
    PORT: process.env.PORT,
    VITE_API_BASE_URL: process.env.VITE_API_BASE_URL,
    VITE_GOOGLE_CLIENT_ID: process.env.VITE_GOOGLE_CLIENT_ID,
    VITE_GOOGLE_CLIENT_SECRET: process.env.VITE_GOOGLE_CLIENT_SECRET,
    VITE_AUTH_REDIRECT_URI: process.env.VITE_AUTH_REDIRECT_URI,
    VITE_JWT_SECRET: process.env.VITE_JWT_SECRET,
    VITE_ENABLE_ANALYTICS: process.env.VITE_ENABLE_ANALYTICS,
    VITE_ENABLE_ERROR_REPORTING: process.env.VITE_ENABLE_ERROR_REPORTING,
    VITE_ENABLE_EXPERIMENTAL_FEATURES: process.env.VITE_ENABLE_EXPERIMENTAL_FEATURES,
    VITE_DEFAULT_THEME: process.env.VITE_DEFAULT_THEME,
    VITE_PRIMARY_COLOR: process.env.VITE_PRIMARY_COLOR,
    VITE_BUILD_DATE: process.env.VITE_BUILD_DATE,
    VITE_COMMIT_HASH: process.env.VITE_COMMIT_HASH,
    VITE_VERSION: process.env.VITE_VERSION,
  };

  // Merge with overrides
  const mergedEnv = { ...processEnv, ...overrides };

  // Validate and parse
  const parsedEnv = envSchema.parse(mergedEnv);

  // Add helper methods
  const env = {
    ...parsedEnv,
    
    /**
     * Check if running in development
     */
    isDev: () => parsedEnv.NODE_ENV === "development",
    
    /**
     * Check if running in production
     */
    isProd: () => parsedEnv.NODE_ENV === "production",
    
    /**
     * Check if running in test
     */
    isTest: () => parsedEnv.NODE_ENV === "test",
    
    /**
     * Get API URL
     */
    getApiUrl: (path: string = "") => {
      return new URL(path, parsedEnv.VITE_API_BASE_URL).toString();
    },
    
    /**
     * Get Google OAuth URL
     */
    getGoogleOAuthUrl: () => {
      if (!parsedEnv.VITE_GOOGLE_CLIENT_ID) {
        throw new Error("Google Client ID not configured");
      }
      
      const authUrl = new URL("https://accounts.google.com/o/oauth2/v2/auth");
      authUrl.searchParams.set("client_id", parsedEnv.VITE_GOOGLE_CLIENT_ID);
      authUrl.searchParams.set("redirect_uri", parsedEnv.VITE_AUTH_REDIRECT_URI);
      authUrl.searchParams.set("response_type", "code");
      authUrl.searchParams.set("scope", "email profile");
      authUrl.searchParams.set("access_type", "offline");
      authUrl.searchParams.set("prompt", "consent");
      
      return authUrl.toString();
    },
    
    /**
     * Check if feature is enabled
     */
    isFeatureEnabled: (feature: "analytics" | "errorReporting" | "experimental") => {
      switch (feature) {
        case "analytics": return parsedEnv.VITE_ENABLE_ANALYTICS;
        case "errorReporting": return parsedEnv.VITE_ENABLE_ERROR_REPORTING;
        case "experimental": return parsedEnv.VITE_ENABLE_EXPERIMENTAL_FEATURES;
        default: return false;
      }
    },
  };

  return env;
}

/**
 * Get environment variable with type safety
 */
export function getEnv<T extends keyof EnvVariables>(key: T): EnvVariables[T] {
  const env = loadEnv();
  return env[key];
}