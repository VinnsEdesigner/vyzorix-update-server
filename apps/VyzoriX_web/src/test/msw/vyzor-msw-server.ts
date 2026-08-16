import { setupServer } from 'msw/node';
import { setupWorker } from 'msw/browser';
import { createVyzorHandlers } from './vyzor-msw-handlers-index';

export function createVyzorMswServer() {
  return setupServer(...createVyzorHandlers());
}

export function createVyzorMswWorker() {
  return setupWorker(...createVyzorHandlers());
}

export { createVyzorHandlers } from './vyzor-msw-handlers-index';
