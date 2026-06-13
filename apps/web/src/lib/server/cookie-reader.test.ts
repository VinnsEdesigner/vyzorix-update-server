/**
 * cookie-reader.test.ts - Tests for server-side cookie parsing and auth state prefetching
 *
 * Tests:
 * 1. parseCookies - cookie header parsing
 * 2. getPrefetchedAuthState - SSR auth state fetching
 * 3. hasSessionCookie - quick session check
 */

import { describe, expect, it, vi, beforeEach } from "vitest";

import { parseCookies, getPrefetchedAuthState, hasSessionCookie } from "./cookie-reader";

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch as unknown as typeof global.fetch;

// Mock environment
const originalEnv = process.env;
beforeEach(() => {
  vi.clearAllMocks();
  process.env = { ...originalEnv, API_URL: "http://localhost:3000" };
});

describe("parseCookies", () => {
  it("should parse single cookie", () => {
    const result = parseCookies("vyz_session=abc123");
    expect(result).toEqual({ vyz_session: "abc123" });
  });

  it("should parse multiple cookies", () => {
    const result = parseCookies("session=abc; token=xyz; auth=123");
    expect(result).toEqual({ session: "abc", token: "xyz", auth: "123" });
  });

  it("should handle empty string", () => {
    const result = parseCookies("");
    expect(result).toEqual({});
  });

  it("should handle null", () => {
    const result = parseCookies(null);
    expect(result).toEqual({});
  });

  it("should decode URI components", () => {
    const result = parseCookies("token=abc%3D123");
    expect(result).toEqual({ token: "abc=123" });
  });

  it("should handle cookies with spaces", () => {
    const result = parseCookies("session=value; another=123");
    expect(result).toEqual({ session: "value", another: "123" });
  });

  it("should handle vyz_session cookie (Library compatibility)", () => {
    const result = parseCookies("vyz_session=encrypted-operator-id");
    expect(result).toEqual({ vyz_session: "encrypted-operator-id" });
  });

  it("should handle vyzozix_session cookie (old Library name)", () => {
    const result = parseCookies("vyzorix_session=old-encrypted-id");
    expect(result).toEqual({ vyzorix_session: "old-encrypted-id" });
  });

  it("should return empty for malformed cookie strings", () => {
    const result = parseCookies("no-equals-sign");
    expect(result).toEqual({});
  });

  it("should handle trailing semicolon", () => {
    const result = parseCookies("token=abc;");
    expect(result).toEqual({ token: "abc" });
  });
});

describe("hasSessionCookie", () => {
  it("should return true when vyz_session cookie exists", () => {
    const mockRequest = {
      headers: new Headers({ cookie: "vyz_session=abc123" }),
    } as unknown as Request;

    expect(hasSessionCookie(mockRequest)).toBe(true);
  });

  it("should return true when vyzozix_session cookie exists", () => {
    const mockRequest = {
      headers: new Headers({ cookie: "vyzorix_session=abc123" }),
    } as unknown as Request;

    expect(hasSessionCookie(mockRequest)).toBe(true);
  });

  it("should return false when no cookies", () => {
    const mockRequest = {
      headers: new Headers(),
    } as unknown as Request;

    expect(hasSessionCookie(mockRequest)).toBe(false);
  });

  it("should return false when no session cookie", () => {
    const mockRequest = {
      headers: new Headers({ cookie: "other=value" }),
    } as unknown as Request;

    expect(hasSessionCookie(mockRequest)).toBe(false);
  });

  it("should prefer vyz_session over vyzorix_session", () => {
    const mockRequest = {
      headers: new Headers({
        cookie: "vyz_session=new; vyzorix_session=old",
      }),
    } as unknown as Request;

    expect(hasSessionCookie(mockRequest)).toBe(true);
  });
});

describe("getPrefetchedAuthState", () => {
  const mockRequest = (cookie: string | null) =>
    ({
      headers: new Headers(cookie ? { cookie } : {}),
    }) as unknown as Request;

  const mockOperator = {
    id: "op_123",
    email: "test@example.com",
    name: "Test User",
    role: "operator" as const,
    createdAt: 1700000000000,
    emailVerified: true,
  };

  it("should return unauthenticated when no cookie", async () => {
    const result = await getPrefetchedAuthState(mockRequest(null));

    expect(result.isAuthenticated).toBe(false);
    expect(result.operator).toBeNull();
  });

  it("should return unauthenticated when cookie is empty", async () => {
    const result = await getPrefetchedAuthState(mockRequest(""));

    expect(result.isAuthenticated).toBe(false);
    expect(result.operator).toBeNull();
  });

  it("should return authenticated when session is valid", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      json: () => Promise.resolve(mockOperator),
    } as unknown as Response);

    const result = await getPrefetchedAuthState(mockRequest("vyz_session=valid-encrypted-id"));

    expect(result.isAuthenticated).toBe(true);
    expect(result.operator).toEqual(mockOperator);
  });

  it("should return unauthenticated on 401 response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
    } as unknown as Response);

    const result = await getPrefetchedAuthState(mockRequest("vyz_session=invalid-session"));

    expect(result.isAuthenticated).toBe(false);
    expect(result.operator).toBeNull();
  });

  it("should return unauthenticated on network error", async () => {
    mockFetch.mockRejectedValueOnce(new Error("Network error"));

    const result = await getPrefetchedAuthState(mockRequest("vyz_session=some-session"));

    expect(result.isAuthenticated).toBe(false);
    expect(result.operator).toBeNull();
  });

  it("should use vyzorix_session as fallback", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      json: () => Promise.resolve(mockOperator),
    } as unknown as Response);

    const result = await getPrefetchedAuthState(mockRequest("vyzorix_session=old-session"));

    expect(result.isAuthenticated).toBe(true);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/v1/auth/me"),
      expect.objectContaining({
        headers: expect.objectContaining({
          Cookie: expect.stringContaining("vyz_session=old-session"),
        }),
      }),
    );
  });

  it("should prefer vyz_session over vyzorix_session", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      json: () => Promise.resolve(mockOperator),
    } as unknown as Response);

    await getPrefetchedAuthState(
      mockRequest("vyz_session=new-session; vyzorix_session=old-session"),
    );

    // Should use vyz_session (new session)
    expect(mockFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        headers: expect.objectContaining({
          Cookie: expect.stringContaining("vyz_session=new-session"),
        }),
      }),
    );
  });

  it("should use custom API_URL when set", async () => {
    process.env.API_URL = "https://api.example.com";
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      json: () => Promise.resolve(mockOperator),
    } as unknown as Response);

    await getPrefetchedAuthState(mockRequest("vyz_session=abc"));

    expect(mockFetch).toHaveBeenCalledWith("https://api.example.com/v1/auth/me", expect.anything());

    process.env.API_URL = "http://localhost:3000";
  });

  it("should include Accept header", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "content-type": "application/json" }),
      json: () => Promise.resolve(mockOperator),
    } as unknown as Response);

    await getPrefetchedAuthState(mockRequest("vyz_session=test"));

    expect(mockFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: "application/json",
        }),
      }),
    );
  });
});
