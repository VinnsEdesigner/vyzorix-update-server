

import type {
  ClientCredential,
  ClientCredentialWithSecret,
  Platform,
} from "./clientcredentials-entity";


export interface RawClientCredentialItem {
  id: string;
  platform: string;
  name: string;
  allowedOrigins?: string[];
  allowedPaths?: string[];
  rateLimit: number;
  isActive: boolean;
  createdAt: number;
}


export interface RawClientCredentialCreated {
  clientId: string;
  clientSecret: string;
  platform: string;
  name: string;
  createdAt: number;
}


export interface RawClientCredentialResponse {
  client: RawClientCredentialItem;
}


export interface RawClientCredentialListResponse {
  clients: RawClientCredentialItem[];
}


export function clientCredentialFromRaw(raw: RawClientCredentialItem): ClientCredential {
  return {
    clientId: raw.id,
    platform: raw.platform as Platform,
    name: raw.name,
    allowedOrigins: raw.allowedOrigins,
    allowedPaths: raw.allowedPaths,
    rateLimit: raw.rateLimit,
    active: raw.isActive,
    createdAt: raw.createdAt,
  };
}


export function clientCredentialWithSecretFromRaw(
  raw: RawClientCredentialCreated,
  rateLimit: number = 100
): ClientCredentialWithSecret {
  return {
    clientId: raw.clientId,
    clientSecret: raw.clientSecret,
    platform: raw.platform as Platform,
    name: raw.name,
    rateLimit: rateLimit,
    active: true,
    createdAt: raw.createdAt,
  };
}


export function clientCredentialListFromRaw(raw: RawClientCredentialListResponse): {
  clients: ClientCredential[];
} {
  return {
    clients: raw.clients.map(clientCredentialFromRaw),
  };
}
