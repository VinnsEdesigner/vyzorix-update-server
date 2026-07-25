import { restClient } from "../_shared/rest-client";

const PATHS = {
  health: "/health",
  healthz: "/healthz",
} as const;

export interface HealthResponse {
  status: "ok";
  timestamp: string;
  checks?: {
    database?: string;
  };
}

export interface DetailedHealthResponse {
  ok: boolean;
  database: string;
  dbOk: boolean;
  serverTime: number;
  connectedDevices: number;
  version: string;
  dbError?: string;
}

export const health = {
  async check(): Promise<HealthResponse> {
    return restClient.get<HealthResponse>(PATHS.health);
  },

  async checkDetailed(): Promise<DetailedHealthResponse> {
    return restClient.get<DetailedHealthResponse>(PATHS.healthz);
  },
};
