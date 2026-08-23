import { http, HttpResponse, delay } from 'msw';
import {
  buildVersion,
  buildSyncState,
  buildPush,
  buildPushRequest,
  resetFixtureCounter,
} from '../fixtures/vyzor-test-fixtures';
import type {
  UpdateVersionResponse,
  UpdatePushHistoryEntry,
  UpdateVersionListResult,
  UpdatePushHistoryListResult,
  UpdatePushDetailResult,
  UpdatePushResult,
  UpdateCancelPushResult,
  UpdateSyncResponse,
  UpdateStatusResult,
  UpdateChangelogResult,
} from '@vyzorix/api-client';

const API_BASE = '/v1/updates';

function toWireVersion(v: ReturnType<typeof buildVersion>): UpdateVersionResponse {
  return {
    version: v.version,
    apkFilename: v.apkFilename,
    apkSize: v.apkSize,
    sha256: v.sha256,
    releaseType: v.releaseType,
    releaseNotes: v.releaseNotes,
    releasedAt: v.releaseDate.getTime(),
    isLatest: v.isLatest,
  };
}

function toWireHistoryEntry(p: ReturnType<typeof buildPush>): UpdatePushHistoryEntry {
  return {
    id: p.id,
    version: p.version,
    installType: p.installType,
    status: p.status,
    initiatedBy: p.initiatedBy,
    initiatedAt: p.initiatedAt.getTime(),
    scheduledAt: p.scheduledAt?.getTime(),
    completedAt: p.completedAt?.getTime(),
    cancelledAt: p.cancelledAt?.getTime(),
    deviceCount: p.devices.total,
    devices: {
      pending: p.devices.pending,
      sent: p.devices.sent,
      acknowledged: p.devices.acknowledged,
      failed: p.devices.failed,
    },
  };
}

export function createUpdatesHandlers() {
  resetFixtureCounter();
  const version1 = buildVersion({ version: 'v1.2.0', isLatest: true });
  const version2 = buildVersion({ version: 'v1.1.0', isLatest: false, id: 'version-test-2' });
  const version3 = buildVersion({ version: 'v1.0.0', isLatest: false, id: 'version-test-3' });
  const versions = [version1, version2, version3];

  const push1 = buildPush({ id: 'push-test-1', status: 'completed' });
  const push2 = buildPush({ id: 'push-test-2', status: 'pending' });

  return [
    http.get(`${API_BASE}/status`, async () => {
      await delay(50);
      const sync = buildSyncState();
      const result: UpdateStatusResult = {
        sync: {
          status: sync.status,
          lastSyncAt: sync.lastSyncAt?.getTime(),
          nextSyncAt: sync.nextSyncAt?.getTime(),
          versionsFound: sync.versionsFound,
          error: sync.error,
        },
        latest: toWireVersion(version1),
      };
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/versions`, async ({ request }) => {
      await delay(50);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: UpdateVersionListResult = {
        versions: versions.map(toWireVersion),
        pagination: {
          page,
          limit,
          total: versions.length,
          total_pages: 1,
        },
      };
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/changelog`, async () => {
      await delay(30);
      const result: UpdateChangelogResult = {
        changelog: versions.map((v) => ({
          version: v.version,
          date: v.releaseDate.toISOString(),
          type: v.releaseType,
          notes: v.releaseNotes ?? '',
        })),
      };
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/history`, async ({ request }) => {
      await delay(50);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: UpdatePushHistoryListResult = {
        pushes: [push1, push2].map(toWireHistoryEntry),
        pagination: {
          page,
          limit,
          total: 2,
          total_pages: 1,
        },
      };
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/history/:pushId`, async ({ params }) => {
      await delay(30);
      const pushId = params.pushId as string;
      const found = [push1, push2].find((p) => p.id === pushId);
      if (!found) {
        return HttpResponse.json(
          { error: 'push not found' },
          { status: 404 },
        );
      }
      const result: UpdatePushDetailResult = {
        id: found.id,
        version: found.version,
        installType: found.installType,
        status: found.status,
        initiatedBy: found.initiatedBy,
        initiatedAt: found.initiatedAt.getTime(),
        scheduledAt: found.scheduledAt?.getTime(),
        completedAt: found.completedAt?.getTime(),
        cancelledAt: found.cancelledAt?.getTime(),
        devices: [],
      };
      return HttpResponse.json(result);
    }),

    http.post(`${API_BASE}/push`, async ({ request }) => {
      await delay(100);
      const body = (await request.json()) as ReturnType<typeof buildPushRequest>;
      const result: UpdatePushResult = {
        pushId: 'push-test-new',
        version: body.version,
        deviceIds: body.deviceIds,
        installType: body.installType,
        scheduledAt: body.scheduledAt,
        status: 'pending',
        initiatedBy: 'operator-test-1',
        initiatedAt: Date.now(),
        devices: {
          total: body.deviceIds?.length ?? 0,
          pending: body.deviceIds?.length ?? 0,
          sent: 0,
          acknowledged: 0,
          failed: 0,
        },
      };
      return HttpResponse.json(result, { status: 201 });
    }),

    http.post(`${API_BASE}/history/:pushId/cancel`, async ({ params }) => {
      await delay(80);
      const result: UpdateCancelPushResult = {
        id: params.pushId as string,
        status: 'cancelled',
        cancelledAt: Date.now(),
        cancelledBy: 'operator-test-1',
      };
      return HttpResponse.json(result);
    }),

    http.post(`${API_BASE}/sync`, async () => {
      await delay(200);
      const result: UpdateSyncResponse = {
        status: 'started',
        startedAt: Date.now(),
        versionsFound: 3,
      };
      return HttpResponse.json(result);
    }),
  ];
}
