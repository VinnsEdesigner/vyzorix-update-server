

export type Platform = "web" | "ios" | "android";

export interface ClientCredential {
  clientId: string;
  platform: Platform;
  name: string;
  allowedOrigins?: string[];
  allowedPaths?: string[];
  rateLimit: number;
  active: boolean;
  createdAt: number;
}

export interface ClientCredentialWithSecret extends ClientCredential {
  clientSecret: string;
}

export interface CreateClientCredentialRequest {
  name: string;
  platform: Platform;
  allowedOrigins?: string[];
  allowedPaths?: string[];
  rateLimit?: number;
}

export interface UpdateClientCredentialRequest {
  name?: string;
  allowedOrigins?: string[];
  allowedPaths?: string[];
  rateLimit?: number;
  active?: boolean;
}

export interface ClientCredentialListResponse {
  clients: ClientCredential[];
}
