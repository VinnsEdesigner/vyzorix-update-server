import { restClient, getOrganizationContext } from "../_shared/rest-client";

const PATHS = {
  connections: "/v1/connections",
  connectionMetrics: "/v1/connections/metrics",
} as const;

export interface ConnectionStatus {
  imei: string;
  connected: boolean;
  connectionId?: string;
  connectedAt?: number;
  lastActivityAt?: number;
  ipAddress?: string;
  userAgent?: string;
}

export interface ConnectionMetrics {
  totalConnections: number;
  activeConnections: number;
  disconnectedToday: number;
  averageSessionDuration: number;
  peakConnections: number;
  peakTime?: number;
}

export const connections = {
  async getAllStatus(organizationId?: string): Promise<{
    connections: ConnectionStatus[];
    total: number;
  }> {
    return restClient.get<{
      connections: ConnectionStatus[];
      total: number;
    }>(PATHS.connections, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async getMetrics(organizationId?: string): Promise<ConnectionMetrics> {
    return restClient.get<ConnectionMetrics>(PATHS.connectionMetrics, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },
};
