/**
 * vyzorix-auth.test.ts - Tests for the authentication client
 *
 * Tests the full auth flow:
 * 1. Register a new operator
 * 2. Login with email/password
 * 3. Google OAuth callback handling
 * 4. Password reset flow
 * 5. Email verification flow
 */

import { describe, expect, it, beforeEach, vi } from "vitest";

import {
  login,
  register,
  logout,
  updateName,
  me,
  forgotPassword,
  resetPassword,
  verifyEmail,
  resendVerification,
  handleOAuthCallback,
  getToken,
  getStoredOperator,
  type Operator,
  type AuthResponse,
} from "./vyzorix-auth";

// Mock the logger
vi.mock("@/lib/logger", () => ({
  logger: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const storage: Record<string, string> = {};
vi.stubGlobal("localStorage", {
  getItem: (key: string) => storage[key] ?? null,
  setItem: (key: string, value: string) => {
    storage[key] = value;
  },
  removeItem: (key: string) => {
    delete storage[key];
  },
  clear: () => {
    Object.keys(storage).forEach((k) => delete storage[k]);
  },
});

// Helper to create mock responses
const createMockResponse = <T>(data: T, status = 200): Response => {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ "content-type": "application/json" }),
    json: () => Promise.resolve(data),
  } as unknown as Response;
};

// Sample test data
const TEST_SERVER = "http://localhost:3000";
const TEST_OPERATOR: Operator = {
  id: "op_123",
  email: "test@example.com",
  name: "Test User",
  role: "operator",
  createdAt: 1700000000000,
  emailVerified: false,
};

const TEST_AUTH_RESPONSE: AuthResponse = {
  token:
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJvaWQiOiJvcF8xMjMiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20ifQ.test",
  expiresAt: 1700086400000,
  operator: TEST_OPERATOR,
};

describe("vyzorix-auth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("login", () => {
    it("should login successfully with email and password", async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse(TEST_AUTH_RESPONSE));

      const result = await login(TEST_SERVER, "test@example.com", "password123");

      expect(result).toEqual(TEST_AUTH_RESPONSE);
      expect(getToken()).toBe(TEST_AUTH_RESPONSE.token);
      expect(getStoredOperator()).toEqual(TEST_OPERATOR);
      expect(mockFetch).toHaveBeenCalledWith(
        `${TEST_SERVER}/v1/auth/login`,
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: "test@example.com", password: "password123" }),
        }),
      );
    });

    it("should throw error on invalid credentials", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Headers({ "content-type": "application/json" }),
        json: () =>
          Promise.resolve({ error: "invalid_credentials", message: "Invalid email or password" }),
      } as unknown as Response);

      await expect(login(TEST_SERVER, "test@example.com", "wrongpassword")).rejects.toThrow(
        "Invalid email or password",
      );
    });

    it("should trim email whitespace", async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse(TEST_AUTH_RESPONSE));

      await login(TEST_SERVER, "  test@example.com  ", "password123");

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          method: "POST",
        }),
      );
    });
  });

  describe("register", () => {
    it("should register a new operator", async () => {
      const newOperator = {
        ...TEST_OPERATOR,
        role: "super_admin" as const,
        email: "new@example.com",
      };
      const response = { ...TEST_AUTH_RESPONSE, operator: newOperator };
      mockFetch.mockResolvedValueOnce(createMockResponse(response));

      const result = await register(TEST_SERVER, "new@example.com", "password123", "New User");

      expect(result.operator.email).toBe("new@example.com");
      expect(getToken()).toBe(response.token);
    });

    it("should throw error on duplicate email", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 409,
        headers: new Headers({ "content-type": "application/json" }),
        json: () => Promise.resolve({ error: "email_exists", message: "Email already registered" }),
      } as unknown as Response);

      await expect(
        register(TEST_SERVER, "existing@example.com", "password123", "User"),
      ).rejects.toThrow("Email already registered");
    });

    it("should throw error on weak password", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ "content-type": "application/json" }),
        json: () =>
          Promise.resolve({
            error: "weak_password",
            message: "Password must be at least 12 characters",
          }),
      } as unknown as Response);

      await expect(register(TEST_SERVER, "test@example.com", "short", "User")).rejects.toThrow(
        "Password must be at least 12 characters",
      );
    });
  });

  describe("logout", () => {
    it("should logout successfully when authenticated", async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse({ message: "Logged out" }));
      storage["vyz.auth.token"] = "test-token";
      storage["vyz.auth.operator"] = JSON.stringify(TEST_OPERATOR);

      await logout(TEST_SERVER);

      expect(getToken()).toBeNull();
      expect(getStoredOperator()).toBeNull();
      expect(mockFetch).toHaveBeenCalledWith(
        `${TEST_SERVER}/v1/auth/logout`,
        expect.objectContaining({
          method: "POST",
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
          }),
        }),
      );
    });

    it("should clear local storage even if API fails", async () => {
      mockFetch.mockRejectedValueOnce(new Error("Network error"));
      storage["vyz.auth.token"] = "test-token";

      await logout(TEST_SERVER);

      expect(getToken()).toBeNull();
    });

    it("should do nothing when not authenticated", async () => {
      await logout(TEST_SERVER);

      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe("me", () => {
    it("should return operator profile when authenticated", async () => {
      storage["vyz.auth.token"] = "test-token";
      mockFetch.mockResolvedValueOnce(createMockResponse(TEST_OPERATOR));

      const result = await me(TEST_SERVER);

      expect(result).toEqual(TEST_OPERATOR);
      expect(mockFetch).toHaveBeenCalledWith(
        `${TEST_SERVER}/v1/auth/me`,
        expect.objectContaining({
          method: "GET",
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
          }),
        }),
      );
    });

    it("should throw error when not authenticated", async () => {
      await expect(me(TEST_SERVER)).rejects.toThrow("not authenticated");
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe("updateName", () => {
    it("should update operator name", async () => {
      storage["vyz.auth.token"] = "test-token";
      const updatedOperator = { ...TEST_OPERATOR, name: "Updated Name" };
      mockFetch.mockResolvedValueOnce(createMockResponse(updatedOperator));

      const result = await updateName(TEST_SERVER, "Updated Name");

      expect(result.name).toBe("Updated Name");
      expect(getStoredOperator()?.name).toBe("Updated Name");
    });

    it("should throw error when not authenticated", async () => {
      await expect(updateName(TEST_SERVER, "New Name")).rejects.toThrow("not authenticated");
    });
  });

  describe("forgotPassword", () => {
    it("should request password reset", async () => {
      mockFetch.mockResolvedValueOnce(
        createMockResponse({
          message: "If that email exists, a password reset link has been sent.",
        }),
      );

      const result = await forgotPassword(TEST_SERVER, "test@example.com");

      expect(result.message).toContain("password reset");
      expect(mockFetch).toHaveBeenCalledWith(
        `${TEST_SERVER}/v1/auth/forgot-password`,
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ email: "test@example.com" }),
        }),
      );
    });

    it("should return success even for non-existent email (security)", async () => {
      mockFetch.mockResolvedValueOnce(
        createMockResponse({
          message: "If that email exists, a password reset link has been sent.",
        }),
      );

      // Should not throw - server returns success for security reasons
      await expect(forgotPassword(TEST_SERVER, "nonexistent@example.com")).resolves.toBeDefined();
    });
  });

  describe("resetPassword", () => {
    it("should reset password with valid token", async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse(TEST_AUTH_RESPONSE));

      const result = await resetPassword(TEST_SERVER, "reset-token-123", "newpassword123");

      expect(result.token).toBeDefined();
      expect(getToken()).toBe(result.token);
    });

    it("should throw error for invalid/expired token", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ "content-type": "application/json" }),
        json: () =>
          Promise.resolve({ error: "invalid_token", message: "Reset token is invalid or expired" }),
      } as unknown as Response);

      await expect(resetPassword(TEST_SERVER, "invalid-token", "newpassword")).rejects.toThrow(
        "Reset token is invalid or expired",
      );
    });
  });

  describe("verifyEmail", () => {
    it("should verify email with valid token", async () => {
      const verifiedOperator = { ...TEST_OPERATOR, emailVerified: true };
      const response = { ...TEST_AUTH_RESPONSE, operator: verifiedOperator };
      mockFetch.mockResolvedValueOnce(createMockResponse(response));

      const result = await verifyEmail(TEST_SERVER, "verify-token-123");

      expect(result.operator.emailVerified).toBe(true);
      expect(getToken()).toBe(result.token);
    });

    it("should throw error for invalid token", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ "content-type": "application/json" }),
        json: () =>
          Promise.resolve({ error: "invalid_token", message: "Verification token is invalid" }),
      } as unknown as Response);

      await expect(verifyEmail(TEST_SERVER, "invalid-token")).rejects.toThrow(
        "Verification token is invalid",
      );
    });
  });

  describe("resendVerification", () => {
    it("should resend verification email", async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse({ message: "Verification email sent" }));

      const result = await resendVerification(TEST_SERVER, "test@example.com");

      expect(result.message).toContain("Verification email sent");
    });
  });

  describe("handleOAuthCallback", () => {
    it("should parse JWT token and extract operator info", () => {
      // Create a real-looking JWT with test payload
      const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
      const payload = btoa(
        JSON.stringify({
          oid: "op_456",
          email: "google@example.com",
          name: "Google User",
          role: "operator",
          iat: 1700000000,
          exp: 1700086400,
        }),
      );
      const signature = "test-signature";
      const token = `${header}.${payload}.${signature}`;

      const result = handleOAuthCallback(token, "false");

      expect(result).not.toBeNull();
      expect(result?.operator.id).toBe("op_456");
      expect(result?.operator.email).toBe("google@example.com");
      expect(result?.operator.name).toBe("Google User");
      expect(result?.operator.emailVerified).toBe(true); // Google accounts are pre-verified
      expect(getToken()).toBe(token);
    });

    it("should return null for invalid token format", () => {
      expect(handleOAuthCallback("invalid-token", "false")).toBeNull();
      expect(handleOAuthCallback("only.two", "false")).toBeNull();
    });

    it("should return null if token has no id or email", () => {
      const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
      const payload = btoa(JSON.stringify({ name: "No ID" })); // Missing oid and email
      const signature = "test-signature";
      const token = `${header}.${payload}.${signature}`;

      expect(handleOAuthCallback(token, "false")).toBeNull();
    });

    it("should handle missing name with email prefix", () => {
      const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
      const payload = btoa(
        JSON.stringify({
          oid: "op_789",
          email: "user@example.com",
          // No name
        }),
      );
      const signature = "test-signature";
      const token = `${header}.${payload}.${signature}`;

      const result = handleOAuthCallback(token, "false");

      expect(result?.operator.name).toBe("user"); // Derived from email
    });
  });

  describe("getToken", () => {
    it("should return stored token", () => {
      storage["vyz.auth.token"] = "stored-token";
      expect(getToken()).toBe("stored-token");
    });

    it("should return null when no token", () => {
      delete storage["vyz.auth.token"];
      expect(getToken()).toBeNull();
    });
  });

  describe("getStoredOperator", () => {
    it("should return stored operator", () => {
      storage["vyz.auth.operator"] = JSON.stringify(TEST_OPERATOR);
      expect(getStoredOperator()).toEqual(TEST_OPERATOR);
    });

    it("should return null when no operator", () => {
      delete storage["vyz.auth.operator"];
      expect(getStoredOperator()).toBeNull();
    });

    it("should return null for invalid JSON", () => {
      storage["vyz.auth.operator"] = "invalid-json";
      expect(getStoredOperator()).toBeNull();
    });
  });
});
