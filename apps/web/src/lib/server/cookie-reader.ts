/**
 * cookie-reader.ts - Server-side cookie parsing for SSR hydration
 *
 * This module handles reading authentication cookies from HTTP requests
 * and fetching the authenticated operator data from the Go API.
 *
 * Based on Library's server.ts implementation pattern
 */

export interface Operator {
  id: string;
  email: string;
  name: string;
  role: "viewer" | "operator" | "super_admin";
  createdAt?: number;
  emailVerified?: boolean;
  thresholds?: Record<string, number>;
}

export interface PrefetchedAuthState {
  isAuthenticated: boolean;
  operator: Operator | null;
}

/**
 * Parse cookies from Cookie header string
 *
 * Matches Library's server.ts cookie parsing pattern:
 *   const reqCookies = req.headers.cookie || '';
 *   if (reqCookies.includes('vyzorix_session=')) { ... }
 */
const parseCookies = (cookieHeader: string | null): Record<string, string> => {
  if (!cookieHeader) return {};

  return cookieHeader.split(";").reduce<Record<string, string>>((acc, cookie) => {
    const trimmed = cookie.trim();
    const eqIndex = trimmed.indexOf("=");
    if (eqIndex > 0) {
      const key = trimmed.substring(0, eqIndex).trim();
      const value = decodeURIComponent(trimmed.substring(eqIndex + 1));
      acc[key] = value;
    }
    return acc;
  }, {});
};

export { parseCookies };

/**
 * Server-side cookie reader for SSR hydration
 *
 * This runs on the SERVER (Node.js) during SSR:
 * 1. Extracts session cookie from request headers
 * 2. Calls Go API to validate session and get operator data
 * 3. Returns auth state to be injected into HTML
 *
 * Mirrors Library's server.ts flow:
 *   if (reqCookies.includes('vyzorix_session=')) {
 *     const meResponse = await fetch(`${GO_BACKEND}/api/auth/me`, {
 *       headers: { 'Cookie': reqCookies }
 *     });
 *     if (meResponse.ok) {
 *       const report = await meResponse.json();
 *       prefetchedState.view = 'success';
 *       prefetchedState.successReport = report;
 *     }
 *   }
 */
const getPrefetchedAuthState = async (request: Request): Promise<PrefetchedAuthState> => {
  // 1. Get cookie header from request
  const cookieHeader = request.headers.get("cookie");
  if (!cookieHeader) {
    return { isAuthenticated: false, operator: null };
  }

  // 2. Parse cookies (same logic as Library: server.ts)
  const cookies = parseCookies(cookieHeader);

  // 3. Check for session cookie
  // Supports both old 'vyzorix_session' (Library) and new 'vyz_session' (vyzorix-update-server)
  const sessionCookie = cookies["vyz_session"] ?? cookies["vyzorix_session"];
  if (!sessionCookie) {
    return { isAuthenticated: false, operator: null };
  }

  // 4. Call Go API to get operator data
  // The Go API reads the cookie from the request and validates the session
  // During migration, we'll switch from JWT validation to cookie-based validation
  const apiUrl = process.env.API_URL ?? "http://localhost:3000";

  try {
    // Send the cookie to the Go API for validation
    const response = await fetch(`${apiUrl}/v1/auth/me`, {
      headers: {
        // Forward the session cookie for validation
        Cookie: `vyz_session=${sessionCookie}`,
        Accept: "application/json",
      },
    });

    // If not authenticated (401), return unauthenticated state
    // The Go API returns 401 when the JWT/session is invalid or expired
    if (!response.ok) {
      if (response.status === 401) {
        console.warn("[SSR Cookie Reader] Session invalid or expired");
        return { isAuthenticated: false, operator: null };
      }
      // For other errors, log and continue as unauthenticated
      console.error(`[SSR Cookie Reader] API error: ${response.status}`);
      return { isAuthenticated: false, operator: null };
    }

    // 5. Parse operator data from response
    const operator = (await response.json()) as Operator;

    return {
      isAuthenticated: true,
      operator,
    };
  } catch (error) {
    console.error("[SSR Cookie Reader] Failed to fetch operator:", error);
    return { isAuthenticated: false, operator: null };
  }
};

export { getPrefetchedAuthState };

/**
 * Check if request has a valid session cookie
 * Useful for quick checks without API call
 */
const hasSessionCookie = (request: Request): boolean => {
  const cookieHeader = request.headers.get("cookie");
  if (!cookieHeader) return false;

  const cookies = parseCookies(cookieHeader);
  return Boolean(cookies["vyz_session"] ?? cookies["vyzorix_session"]);
};

export { hasSessionCookie };
