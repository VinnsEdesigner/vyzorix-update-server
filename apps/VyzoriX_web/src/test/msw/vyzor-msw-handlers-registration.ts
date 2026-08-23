/**
 * MSW handlers for the registration REST endpoints.
 *
 * These mirror the Go server contract consumed by the generated SDK inbox
 * endpoints (`getInbox()`, /v1/device/inbox*). Field names + timestamp formats
 * match the wire DTOs: camelCase fields with numeric epoch-millisecond
 * timestamps and an offset pagination object ({ page, limit, total, total_pages }).
 *
 * NOTE: `/v1/devices*` (registered-device list/get/deregister) intentionally
 * live in the devices handlers file because the REST client reuses those exact
 * paths. Registration-specific tests override them per-test with the
 * registered-device raw shape (see use-registration.test.ts).
 */
import { http, HttpResponse, delay } from 'msw';

const now = () => Date.now();

const INBOX_ENTRY_RAW = {
  id: 'e1',
  imei: '123',
  deviceName: 'Dev',
  deviceClass: 'phone',
  model: 'X',
  manufacturer: 'Acme',
  osVersion: '14',
  appVersion: '1.0',
  fcmToken: 'tok',
  firebaseInstallId: 'fid',
  status: 'pending',
  acknowledgedAt: null,
  approvingAt: null,
  approvedAt: null,
  rejectedAt: null,
  notes: null,
  operatorId: null,
  createdAt: now(),
};

export function createRegistrationHandlers() {
  return [
    // GET /v1/device/inbox — list inbox entries
    http.get('/v1/device/inbox', async ({ request }) => {
      await delay(30);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');
      return HttpResponse.json({
        requests: [INBOX_ENTRY_RAW],
        pagination: { page, limit, total: 1, total_pages: 1 },
      });
    }),

    // POST /v1/device/inbox — create inbox request
    http.post('/v1/device/inbox', async ({ request }) => {
      await delay(40);
      const body = (await request.json()) as { imei?: string };
      return HttpResponse.json(
        {
          id: '1',
          imei: body.imei ?? '123',
          status: 'pending',
          createdAt: now(),
        },
        { status: 201 },
      );
    }),

    // POST /v1/device/confirm — confirm device
    http.post('/v1/device/confirm', async ({ request }) => {
      await delay(40);
      const body = (await request.json()) as { imei?: string };
      return HttpResponse.json({
        device_id: 'd1',
        imei: body.imei ?? '123',
        confirmed: true,
        online: true,
        registered_at: now(),
        server_time: now(),
      });
    }),

    // GET /v1/device/inbox/:imei — get a single inbox entry
    http.get('/v1/device/inbox/:imei', async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      return HttpResponse.json({ ...INBOX_ENTRY_RAW, imei });
    }),

    // POST /v1/device/inbox/:imei/ack — acknowledge inbox entry
    http.post('/v1/device/inbox/:imei/ack', async ({ request, params }) => {
      await delay(40);
      const imei = params.imei as string;
      const body = (await request.json()) as { action?: string };
      const status = body.action === 'reject' ? 'rejected' : 'approved';
      return HttpResponse.json({
        id: '1',
        imei,
        status,
        acknowledgedAt: now(),
        approvingAt: now(),
        approvedAt: body.action === 'reject' ? null : now(),
        rejectedAt: body.action === 'reject' ? now() : null,
        commandSecret: 'secret',
        fcmPushSent: true,
        notes: null,
      });
    }),

    // POST /v1/device/inbox/:imei/resend — resend inbox approval
    http.post('/v1/device/inbox/:imei/resend', async () => {
      await delay(30);
      return HttpResponse.json({ id: '1', imei: '123', status: 'approved', fcmPushSent: true });
    }),

  ];
}
