/**
 * MSW handlers for the device logs REST endpoint.
 *
 * Mirrors the Go server contract for GET /v1/dashboard/device/:imei/logs
 * (application/logs.ListLogsResponse): { events: [{ id, type, timestamp,
 * data }], pagination: { nextCursor?, limit, hasMore } } with epoch-millis
 * timestamps.
 */
import { http, HttpResponse, delay } from 'msw';
import type { DeviceLogEventListResult } from '@vyzorix/api-client';

const now = Date.now();

const listLogsRaw: DeviceLogEventListResult = {
  events: [
    {
      id: 'log-1',
      type: 'connection',
      timestamp: now - 60_000,
      data: { ip: '10.0.0.1' },
    },
    {
      id: 'log-2',
      type: 'command',
      timestamp: now - 30_000,
      data: { command: 'reboot' },
    },
  ],
  pagination: { limit: 50, hasMore: false },
};

export function createLogsHandlers() {
  return [
    // GET /v1/dashboard/device/:imei/logs — list logs
    http.get('/v1/dashboard/device/:imei/logs', async () => {
      await delay(30);
      return HttpResponse.json(listLogsRaw);
    }),
  ];
}
