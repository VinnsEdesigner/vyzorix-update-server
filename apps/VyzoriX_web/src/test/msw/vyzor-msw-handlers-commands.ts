import { http, HttpResponse, delay } from 'msw';
import type {
  CommandHistoryEntry,
  CommandHistoryResult,
  CommandResponse,
} from '@vyzorix/api-client';

const IMEI = '123456789012345';
const now = Date.now();

function makeRawCommandListItem(
  overrides: Partial<CommandHistoryEntry> = {},
): CommandHistoryEntry {
  return {
    id: 'cmd-1',
    dispatchId: 'disp-1',
    deviceId: IMEI,
    command: 'FORCE_SPEAKER',
    status: 'pending',
    createdAt: now,
    ...overrides,
  };
}

function makeRawCommand(overrides: Partial<CommandResponse> = {}): CommandResponse {
  return {
    id: 'cmd-1',
    dispatchId: 'disp-1',
    deviceId: IMEI,
    command: 'FORCE_SPEAKER',
    status: 'pending',
    serverTime: now,
    ...overrides,
  };
}

const historyCommands = [
  makeRawCommandListItem({ id: 'cmd-1', dispatchId: 'disp-1', status: 'pending' }),
  makeRawCommandListItem({ id: 'cmd-2', dispatchId: 'disp-2', status: 'completed' }),
];

export function createCommandsHandlers() {
  return [
    // GET /v1/dashboard/device/:imei/commands — command history
    http.get('/v1/dashboard/device/:imei/commands', async ({ request }) => {
      await delay(30);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: CommandHistoryResult = {
        commands: historyCommands,
        pagination: {
          page,
          limit,
          total: historyCommands.length,
          total_pages: 1,
        },
      };
      return HttpResponse.json(result);
    }),

    // GET /v1/device/:imei/commands/pending — pending commands
    http.get('/v1/device/:imei/commands/pending', async () => {
      await delay(20);
      const pending = [
        makeRawCommand({ id: 'cmd-1', dispatchId: 'disp-1', status: 'pending' }),
      ];
      return HttpResponse.json({ commands: pending });
    }),

    // GET /v1/command/:dispatchId/status — command status
    http.get('/v1/command/:dispatchId/status', async ({ params }) => {
      await delay(20);
      const dispatchId = params.dispatchId as string;
      return HttpResponse.json(
        makeRawCommand({ dispatchId, status: 'completed' }),
      );
    }),

    // POST /v1/device/:imei/command — send command
    http.post('/v1/device/:imei/command', async () => {
      await delay(40);
      return HttpResponse.json({
        dispatchId: 'disp-1',
        command_id: 'cmd-1',
        device_online: true,
        status: 'pending',
        serverTime: now,
      });
    }),

    // DELETE /v1/command/:dispatchId — cancel command
    http.delete('/v1/command/:dispatchId', async ({ params }) => {
      await delay(30);
      return HttpResponse.json({
        cancelled: true,
        dispatchId: params.dispatchId as string,
        serverTime: Date.now(),
      });
    }),

    // POST /v1/command/:dispatchId/retry — retry command
    http.post('/v1/command/:dispatchId/retry', async ({ params }) => {
      await delay(40);
      return HttpResponse.json({
        command_id: 'cmd-1',
        dispatchId: params.dispatchId as string,
        retried: true,
        serverTime: Date.now(),
      });
    }),
  ];
}

export { makeRawCommand, makeRawCommandListItem };
