import { afterAll, afterEach, beforeAll } from 'vitest';
import { createVyzorMswServer } from './vyzor-msw-server';

const server = createVyzorMswServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

export { server as vyzorMswServer };
