import type { VyzorErrorCategory, VyzorErrorClassification } from './vyzor-error-types';

function getStatusFromError(error: unknown): number | undefined {
  if (error == null || typeof error !== 'object') return undefined;
  const status = (error as { status?: unknown }).status;
  return typeof status === 'number' ? status : undefined;
}

function getMessageFromError(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  if (error == null) return 'Unknown error';
  return String(error);
}

const STATUS_TO_CATEGORY: Record<number, VyzorErrorCategory> = {
  400: 'server',
  401: 'auth',
  403: 'auth',
  404: 'route',
  408: 'network',
  429: 'network',
  500: 'server',
  502: 'server',
  503: 'server',
  504: 'network',
};

export function classifyVyzorError(error: unknown): VyzorErrorClassification {
  if (error instanceof Error && error.name === 'ChunkLoadError') {
    return { category: 'network', recoverable: true, retryable: true };
  }

  const statusCode = getStatusFromError(error);
  const message = getMessageFromError(error);

  if (typeof navigator !== 'undefined' && navigator.onLine === false) {
    return { category: 'network', recoverable: true, retryable: true, statusCode };
  }

  if (/network|fetch|timeout|abort|econnrefused|err_network/i.test(message)) {
    return { category: 'network', recoverable: true, retryable: true, statusCode };
  }

  if (statusCode !== undefined) {
    const category = STATUS_TO_CATEGORY[statusCode] ?? 'server';
    const recoverable = category === 'auth' || category === 'network';
    const retryable = category === 'network' || category === 'server';
    return { category, recoverable, retryable, statusCode };
  }

  return { category: 'unknown', recoverable: true, retryable: true, statusCode };
}

export function classifyRenderError(_error: unknown): VyzorErrorClassification {
  return {
    category: 'render',
    recoverable: false,
    retryable: false,
  };
}

export function classifyRouteError(error: unknown): VyzorErrorClassification {
  const statusCode = getStatusFromError(error);
  if (statusCode === 404) {
    return { category: 'route', recoverable: false, retryable: false, statusCode };
  }
  return { category: 'route', recoverable: false, retryable: false, statusCode };
}
