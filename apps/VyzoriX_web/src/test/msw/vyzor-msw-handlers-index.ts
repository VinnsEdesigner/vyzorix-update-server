import type { HttpHandler } from 'msw';
import { createAuthHandlers } from './vyzor-msw-handlers-auth';
import { createDevicesHandlers } from './vyzor-msw-handlers-devices';
import { createUpdatesHandlers } from './vyzor-msw-handlers-updates';
import { createSettingsHandlers } from './vyzor-msw-handlers-settings';
import { createCommandsHandlers } from './vyzor-msw-handlers-commands';
import { createDiagnosticsHandlers } from './vyzor-msw-handlers-diagnostics';
import { createLogsHandlers } from './vyzor-msw-handlers-logs';
import { createRegistrationHandlers } from './vyzor-msw-handlers-registration';
import { createGraphQLHandlers } from './vyzor-msw-handlers-graphql';
import { createConnectivityHandlers } from './vyzor-msw-handlers-connectivity';
import { createApiKeysHandlers } from './vyzor-msw-handlers-apikeys';
import { createAdminApiKeysHandlers } from './vyzor-msw-handlers-admin-apikeys';

export function createVyzorHandlers(): HttpHandler[] {
  return [
    ...createAuthHandlers(),
    ...createDevicesHandlers(),
    ...createUpdatesHandlers(),
    ...createSettingsHandlers(),
    ...createCommandsHandlers(),
    ...createDiagnosticsHandlers(),
    ...createLogsHandlers(),
    ...createRegistrationHandlers(),
    ...createGraphQLHandlers(),
    ...createConnectivityHandlers(),
    ...createApiKeysHandlers(),
    ...createAdminApiKeysHandlers(),
  ];
}
