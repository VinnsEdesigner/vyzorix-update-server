/**
 * MSW handlers for logs REST endpoints.
 *
 * The API client's logs functions (logs.list, logs.get, logs.stats) use the
 * REAL restClient (axios) against:
 *   GET /v1/dashboard/device/:imei/logs        -> RawLogListResult
 *   GET /v1/dashboard/logs/detail/:id          -> RawLogEntry
 *   GET /v1/dashboard/device/:imei/logs/stats  -> RawLogStats
 *
 * These handlers return raw (snake_case) server shapes; the real domain mappers
 * in @vyzorix/api-client convert them to the entity types the hooks expose.
 */
import { http, HttpResponse, delay } from 'msw';

const now = Date.now();

// Raw (server) shapes mirror the domain Raw* interfaces in @vyzorix/api-client.
const listLogsRaw = {
  logs: [
    {
      id: 'log-1',
      deviceId: '123',
      eventType: 'connection',
      timestamp: Math.floor((now - 60_000) / 1000),
      data: { ip: '10.0.0.1' },
    },
    {
      id: 'log-2',
      deviceId: '123',
      eventType: 'command',
      timestamp: Math.floor((now - 30_000) / 1000),
      data: { command: 'reboot' },
    },
  ],
  pagination: { limit: 50, has_more: false, next_cursor: undefined },
};

const logDetailRaw = {
  id: 'log-detail-1',
  deviceId: '123',
  eventType: 'info',
  timestamp: Math.floor(now / 1000),
  data: { detail: true },
};

const logStatsRaw = {
  total: 2,
  by_type: { connection: 1, command: 1, telemetry: 0, error: 0, warning: 0 },
};

export function createLogsHandlers() {
  return [
    // GET /v1/dashboard/device/:imei/logs — list logs (RawLogListResult)
    http.get('/v1/dashboard/device/:imei/logs', async ({ params, request }) => {
      await delay(30);
      const imei = params.imei as string;
      const url = new URL(request.url);
      const limit = url.searchParams.get('limit');

      // `/stats` is a sibling path; route it via its own handler below by
      // checking the trailing segment. (MSW matches the more specific handler
      // first when both are registered, but guard anyway for safety.)
      const segments = url.pathname.split('/');
      if (segments[segments.length - 1] === 'stats') {
        return HttpResponse.json(logStatsRaw);
      }

      return HttpResponse.json({
        logs: listLogsRaw.logs.map((l) => ({ ...l, deviceId: imei })),
        pagination: {
          limit: limit ? Number(limit) : 50,
          has_more: false,
          next_cursor: undefined,
        },
      });
    }),

    // GET /v1/dashboard/device/:imei/logs/stats — log stats (RawLogStats)
    http.get('/v1/dashboard/device/:imei/logs/stats', async () => {
      await delay(30);
      return HttpResponse.json(logStatsRaw);
    }),

    // GET /v1/dashboard/logs/detail/:id — log detail (RawLogEntry)
    http.get('/v1/dashboard/logs/detail/:id', async ({ params }) => {
      await delay(30);
      const id = params.id as string;
      return HttpResponse.json({ ...logDetailRaw, id });
    }),
  ];
}
