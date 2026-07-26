// Signed API client that implements request signing and response encryption.
// Per PRD: All requests must be signed with HMAC-SHA512 and encrypted with AES-256-GCM.

import {
  ClientCredentials,
  ClientCredentialsStore,
  fetchClientCredentials,
  getActiveCredentials,
} from "./client-credentials";
import { hmacSha512Hex, aes256GcmEncrypt, aes256GcmDecryptCombined } from "./crypto";
import { logger } from "./logger";

export interface SignedRequestOptions {
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: object | string;
  client?: ClientCredentials;
  /** Skip signing for exempt endpoints (health, auth, etc.) */
  skipSigning?: boolean;
}

export interface SignedRequestHeaders {
  "X-Client-ID": string;
  "X-Timestamp": string;
  "X-Signature": string;
  "X-Encrypted-Body": string;
  "Content-Type": string;
}

/**
 * Create a signed and encrypted request.
 * Returns headers needed for the request.
 */
const createSignedRequest = async (
  options: SignedRequestOptions,
): Promise<{ headers: SignedRequestHeaders; body?: string }> => {
  const { method, path, body, client } = options;

  // Use provided client or get active one
  const creds = client ?? getActiveCredentials();
  if (!creds) {
    throw new Error("No client credentials available. Please login first.");
  }

  const timestamp = Math.floor(Date.now() / 1000).toString();

  // Prepare body for signing
  let bodyString: string;

  if (body) {
    bodyString = typeof body === "string" ? body : JSON.stringify(body);
  } else {
    bodyString = "";
  }

  // Step 1: Encrypt body with AES-256-GCM
  // The PRD specifies: body_enc = AES-256-GCM(clientSecret, body)
  const { nonce, ciphertext } = await aes256GcmEncrypt(creds.clientSecret, bodyString || "");

  // Combine nonce and ciphertext for X-Encrypted-Body
  // Format: base64(nonce || ciphertext)
  const encryptedBody = `${nonce}.${ciphertext}`;

  // Step 2: Create signature string
  // Per PRD: string_to_sign = "METHOD\nPATH\nTIMESTAMP\nSHA512(body)"
  // But since we encrypt the body first, we sign the encrypted body hash
  const encoder = new TextEncoder();
  const bodyBytes = encoder.encode(bodyString || "");
  const bodyHashBuffer = await crypto.subtle.digest("SHA-512", bodyBytes);
  const bodyHashHex = Array.from(new Uint8Array(bodyHashBuffer))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");

  const stringToSign = `${method}\n${path}\n${timestamp}\n${bodyHashHex}`;

  // Step 3: Compute HMAC-SHA512 signature
  // Per PRD: signature = HMAC-SHA512(clientSecret, string_to_sign)
  const signatureHex = await hmacSha512Hex(creds.clientSecret, stringToSign);

  // Step 4: Format signature
  // Per PRD: "t={timestamp},v1={hex_signature}"
  const signature = `t=${timestamp},v1=${signatureHex}`;

  const headers: SignedRequestHeaders = {
    "X-Client-ID": creds.clientId,
    "X-Timestamp": timestamp,
    "X-Signature": signature,
    "X-Encrypted-Body": encryptedBody,
    "Content-Type": "application/json",
  };

  return { headers, body: undefined };
};

/**
 * Decrypt a response body.
 */
const decryptResponse = async (
  ciphertextB64: string,
  client: ClientCredentials,
): Promise<string> => {
  // Response format: base64(nonce || ciphertext)
  // Header: X-Encryption-Nonce contains the nonce
  try {
    return await aes256GcmDecryptCombined(client.clientSecret, ciphertextB64);
  } catch (e) {
    logger.error("signed-api", `Decryption failed: ${e}`);
    throw new Error("Failed to decrypt response");
  }
};
export class SignedApiClient {
  private readonly apiUrl: string;
  private client: ClientCredentials | null;

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl.replace(/\/+$/, "");
    this.client = getActiveCredentials();
  }

  /**
   * Set the client credentials to use for signing.
   */
  setClient(client: ClientCredentials): void {
    this.client = client;
  }

  /**
   * Check if we have valid credentials.
   */
  hasCredentials(): boolean {
    return this.client !== null;
  }

  /**
   * Fetch new client credentials after login.
   */
  async fetchCredentials(name?: string): Promise<ClientCredentials> {
    const creds = await fetchClientCredentials(this.apiUrl, name);
    this.client = creds;
    return creds;
  }

  /**
   * Determine if a path requires signing.
   * Matches the server's IsSigningRequiredPath logic.
   */
  private requiresSigning(path: string): boolean {
    // Exempt only truly public endpoints
    if (
      path === "/health/live" ||
      path === "/health/ready" ||
      path === "/healthz" ||
      path === "/health"
    ) {
      return false;
    }
    if (path.startsWith("/assets/") || path === "/favicon.ico") {
      return false;
    }
    if (path.startsWith("/v1/auth/")) {
      return false;
    }
    if (path === "/v1/device/register") {
      return false;
    }
    if (path === "/v1/device/status") {
      return false;
    }
    if (path === "/api/v1/version" || path === "/api/v1/changelog") {
      return false;
    }
    return true;
  }

  /**
   * Make a signed request.
   */
  async request<T = unknown>(
    options: SignedRequestOptions,
  ): Promise<{ data: T; encrypted: boolean }> {
    const { method, path, body, skipSigning } = options;

    if (!this.client) {
      throw new Error("No client credentials. Call fetchCredentials() first.");
    }

    const url = `${this.apiUrl}${path}`;

    // Check if signing is required for this path
    const needsSigning = !skipSigning && this.requiresSigning(path);

    let headers: Record<string, string>;
    let requestBody: string | undefined;

    if (needsSigning) {
      // Create signed + encrypted request
      const signed = await createSignedRequest({
        method,
        path,
        body,
        client: this.client,
      });
      headers = signed.headers as unknown as Record<string, string>;
      // For encrypted requests, we don't send body separately - it's in X-Encrypted-Body
      requestBody = undefined;
    } else {
      // Unsigned request (for exempt endpoints)
      headers = {
        "Content-Type": "application/json",
      };
      if (body) {
        requestBody = typeof body === "string" ? body : JSON.stringify(body);
      } else {
        requestBody = undefined;
      }
    }

    logger.debug("signed-api", `${method} ${path}`, { needsSigning });

    const response = await fetch(url, {
      method,
      headers,
      credentials: "include",
      body: requestBody,
    });

    // Check for encryption header in response
    const encryptionHeader = response.headers.get("X-Content-Encryption");
    const isEncrypted = encryptionHeader === "AES-256-GCM";

    // Get response body
    const responseText = await response.text();

    if (!response.ok) {
      // Try to parse error, even if encrypted
      let errorMessage = `HTTP ${response.status}`;
      try {
        if (isEncrypted && this.client) {
          const decrypted = await decryptResponse(responseText, this.client);
          const errorData = JSON.parse(decrypted);
          errorMessage = errorData.message ?? errorMessage;
        } else {
          const errorData = JSON.parse(responseText);
          errorMessage = errorData.message ?? errorMessage;
        }
      } catch {
        // Use status text
      }
      throw new Error(errorMessage);
    }

    // Decrypt response if encrypted
    if (isEncrypted && this.client) {
      const decrypted = await decryptResponse(responseText, this.client);
      logger.debug("signed-api", `Decrypted response for ${path}`);
      return { data: JSON.parse(decrypted), encrypted: true };
    }

    // Plain JSON response
    try {
      return { data: JSON.parse(responseText), encrypted: false };
    } catch {
      return { data: responseText as unknown as T, encrypted: false };
    }
  }

  /**
   * GET request.
   */
  async get<T = unknown>(path: string, skipSigning?: boolean): Promise<T> {
    const result = await this.request<T>({ method: "GET", path, skipSigning });
    return result.data;
  }

  /**
   * POST request.
   */
  async post<T = unknown>(path: string, body?: object, skipSigning?: boolean): Promise<T> {
    const result = await this.request<T>({ method: "POST", path, body, skipSigning });
    return result.data;
  }

  /**
   * PATCH request.
   */
  async patch<T = unknown>(path: string, body?: object, skipSigning?: boolean): Promise<T> {
    const result = await this.request<T>({ method: "PATCH", path, body, skipSigning });
    return result.data;
  }

  /**
   * DELETE request.
   */
  async delete<T = unknown>(path: string, skipSigning?: boolean): Promise<T> {
    const result = await this.request<T>({ method: "DELETE", path, skipSigning });
    return result.data;
  }
}

// Default instance
let defaultClient: SignedApiClient | null = null;

/**
 * Get or create the default SignedApiClient.
 */
export const getSignedApiClient = (apiUrl: string): SignedApiClient => {
  if (defaultClient?.["apiUrl"] !== apiUrl) {
    defaultClient = new SignedApiClient(apiUrl);
  }
  return defaultClient;
};

/**
 * Clear the default client (on logout).
 */
export const clearSignedApiClient = (): void => {
  defaultClient = null;
  ClientCredentialsStore.clearCache();
};
