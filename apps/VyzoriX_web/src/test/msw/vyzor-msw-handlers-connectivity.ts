/**
 * MSW handlers for connectivity REST endpoints.
 *
 * The connectivity monitor probes network health with a HEAD request to
 * /api/v1/health. These handlers serve that endpoint so the real
 * `checkConnectivity()` code path runs end-to-end during tests.
 */
import { http, HttpResponse } from 'msw';

const HealthEndpoint = '/api/v1/health';

export function createConnectivityHandlers() {
  return [
    http.head(HealthEndpoint, () => new HttpResponse(null, { status: 200 })),
  ];
}
