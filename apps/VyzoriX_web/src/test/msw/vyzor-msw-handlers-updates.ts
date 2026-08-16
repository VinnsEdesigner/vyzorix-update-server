import { http, HttpResponse, delay } from 'msw';
import {
  buildVersion,
  buildSyncState,
  buildPush,
  buildPushRequest,
  resetFixtureCounter,
} from '../fixtures/vyzor-test-fixtures';
import type {
  RawVersion,
  RawSyncState,
  RawUpdatePush,
  RawVersionListResult,
  RawUpdateHistoryResult,
} from '@vyzorix/api-client';

const API_BASE = '/v1/updates';

function toRawVersion(v: ReturnType<typeof buildVersion>): RawVersion {
  return {
    ...v,
    releaseDate: v.releaseDate.toISOString(),
    createdAt: v.createdAt.toISOString(),
    updatedAt: v.updatedAt.toISOString(),
  } as unknown as RawVersion;
}

function toRawPush(p: ReturnType<typeof buildPush>): RawUpdatePush {
  return {
    ...p,
    initiatedAt: p.initiatedAt.toISOString(),
    scheduledAt: p.scheduledAt?.toISOString(),
    completedAt: p.completedAt?.toISOString(),
    cancelledAt: p.cancelledAt?.toISOString(),
  } as unknown as RawUpdatePush;
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
      const syncState: RawSyncState = {
        ...buildSyncState(),
        lastSyncAt: buildSyncState().lastSyncAt?.toISOString(),
      } as unknown as RawSyncState;
      return HttpResponse.json({
        sync: syncState,
        latest: toRawVersion(version1),
      });
    }),

    http.get(`${API_BASE}/versions`, async ({ request }) => {
      await delay(50);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: RawVersionListResult = {
        versions: versions.map(toRawVersion),
        pagination: {
          page,
          limit,
          total: versions.length,
          totalPages: 1,
        },
      } as unknown as RawVersionListResult;
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/changelog`, async () => {
      await delay(30);
      return HttpResponse.json({
        changelog: versions.map((v) => ({
          version: v.version,
          date: v.releaseDate.toISOString(),
          type: v.releaseType,
          notes: v.releaseNotes ?? '',
        })),
      });
    }),

    http.get(`${API_BASE}/history`, async ({ request }) => {
      await delay(50);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: RawUpdateHistoryResult = {
        pushes: [push1, push2].map(toRawPush),
        pagination: {
          page,
          limit,
          total: 2,
          totalPages: 1,
        },
      } as unknown as RawUpdateHistoryResult;
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
      return HttpResponse.json(toRawPush(found));
    }),

    http.post(`${API_BASE}/push`, async ({ request }) => {
      await delay(100);
      const body = (await request.json()) as ReturnType<typeof buildPushRequest>;
      const newPush = buildPush({
        version: body.version,
        installType: body.installType,
        status: 'pending',
        devices: {
          total: body.deviceIds.length,
          pending: body.deviceIds.length,
          sent: 0,
          acknowledged: 0,
          failed: 0,
        },
      });
      return HttpResponse.json(toRawPush(newPush), { status: 201 });
    }),

    http.post(`${API_BASE}/history/:pushId/cancel`, async ({ params }) => {
      await delay(80);
      const pushId = params.pushId as string;
      const cancelled = buildPush({
        id: pushId,
        status: 'cancelled',
        cancelledAt: new Date(),
        cancelledBy: 'operator-test-1',
      });
      return HttpResponse.json(toRawPush(cancelled));
    }),

    http.post(`${API_BASE}/sync`, async () => {
      await delay(200);
      return HttpResponse.json({
        status: 'started',
        startedAt: new Date().toISOString(),
        versionsFound: 3,
      });
    }),
  ];
}
