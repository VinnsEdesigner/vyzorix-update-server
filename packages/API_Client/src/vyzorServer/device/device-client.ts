

import { signRequest, signCommand, signWebSocketConnect, generateNonce, generateTimestamp, generateTimestampMs, type CommandFrame } from '../crypto';
import { restClient } from '../rest/_shared/rest-client';

export interface DeviceCredentials {
  deviceId: string;
  secret: string;
}

export class DeviceClient {
  private credentials: DeviceCredentials | null = null;
  private organizationId: string | null = null;

  constructor(credentials?: DeviceCredentials) {
    if (credentials) {
      this.setCredentials(credentials);
    }
  }

  setCredentials(credentials: DeviceCredentials): void {
    this.credentials = credentials;
  }

  setOrganization(orgId: string): void {
    this.organizationId = orgId;
  }

  clear(): void {
    this.credentials = null;
    this.organizationId = null;
  }

  private getHeaders(method: string, path: string, body?: unknown): Record<string, string> {
    if (!this.credentials) {
      throw new Error('Device credentials not set');
    }

    const bodyStr = body ? JSON.stringify(body) : '';
    const headers = signRequest(method, path, bodyStr, this.credentials.deviceId, this.credentials.secret);

    if (this.organizationId) {
      headers['X-Organization-ID'] = this.organizationId;
    }

    return headers;
  }

    async registerDevice(imei: string, name: string): Promise<{ commandSecret: string }> {
    return restClient.post<{ commandSecret: string }>('/v1/device/register', {
      imei,
      name,
    });
  }

    async confirmDevice(imei: string, commandSecret: string): Promise<{ success: boolean; deviceId: string }> {
    return restClient.post<{ success: boolean; deviceId: string }>('/v1/device/confirm', {
      imei,
      commandSecret,
    });
  }

    async getStatus(imei: string): Promise<{ online: boolean; lastSeen?: number }> {
    return restClient.get<{ online: boolean; lastSeen?: number }>(`/v1/device/${imei}/status`);
  }

    async sendCommand(command: string, args?: Record<string, unknown>): Promise<{ dispatchId: string }> {
    if (!this.credentials) {
      throw new Error('Device credentials not set');
    }

    const dispatchId = `disp_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    const timestamp = generateTimestampMs();

    const frame: CommandFrame = {
      dispatchId,
      command,
      args: args ? JSON.stringify(args) : '{}',
      timestamp,
    };

    const { nonce, signature } = signCommand(frame, this.credentials.deviceId, this.credentials.secret);

    return restClient.post<{ dispatchId: string }>(
      `/v1/device/${this.credentials.deviceId}/command`,
      { command, args },
      {
        headers: {
          ...this.getHeaders('POST', `/v1/device/${this.credentials.deviceId}/command`, { command, args }),
          'X-Nonce': nonce,
          'X-Command-Signature': signature,
          'X-Dispatch-Id': dispatchId,
          'X-Timestamp-Ms': timestamp.toString(),
        },
      }
    );
  }

    async submitTelemetry(metrics: Record<string, number>): Promise<{ received: boolean }> {
    if (!this.credentials) {
      throw new Error('Device credentials not set');
    }

    const path = `/v1/device/${this.credentials.deviceId}/telemetry`;
    const body = JSON.stringify({ metrics });

    return restClient.post<{ received: boolean }>(
      path,
      { metrics },
      {
        headers: this.getHeaders('POST', path, { metrics }),
      }
    );
  }

    async updateFCMToken(token: string): Promise<{ success: boolean }> {
    if (!this.credentials) {
      throw new Error('Device credentials not set');
    }

    const path = `/v1/device/${this.credentials.deviceId}/fcm-token`;
    const body = { fcmToken: token };

    return restClient.post<{ success: boolean }>(
      path,
      body,
      {
        headers: this.getHeaders('POST', path, body),
      }
    );
  }

    async getPendingCommands(): Promise<Array<{ dispatchId: string; command: string; args?: string }>> {
    if (!this.credentials) {
      throw new Error('Device credentials not set');
    }

    const path = `/v1/device/${this.credentials.deviceId}/commands/pending`;

    return restClient.get<Array<{ dispatchId: string; command: string; args?: string }>>(
      path,
      {
        headers: this.getHeaders('GET', path),
      }
    );
  }
}

let deviceClient: DeviceClient | null = null;

export function getDeviceClient(): DeviceClient {
  if (!deviceClient) {
    deviceClient = new DeviceClient();
  }
  return deviceClient;
}

export function initDeviceClient(credentials: DeviceCredentials, organizationId?: string): DeviceClient {
  deviceClient = new DeviceClient(credentials);
  if (organizationId) {
    deviceClient.setOrganization(organizationId);
  }
  return deviceClient;
}
