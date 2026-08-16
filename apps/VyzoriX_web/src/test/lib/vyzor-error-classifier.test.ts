import { describe, it, expect } from 'vitest';
import {
  classifyVyzorError,
  classifyRenderError,
  classifyRouteError,
} from '@/lib/error';

describe('classifyVyzorError', () => {
  it('classifies 401 as auth', () => {
    const result = classifyVyzorError({ status: 401 });
    expect(result.category).toBe('auth');
    expect(result.recoverable).toBe(true);
    expect(result.retryable).toBe(false);
    expect(result.statusCode).toBe(401);
  });

  it('classifies 403 as auth', () => {
    const result = classifyVyzorError({ status: 403 });
    expect(result.category).toBe('auth');
    expect(result.statusCode).toBe(403);
  });

  it('classifies 404 as route', () => {
    const result = classifyVyzorError({ status: 404 });
    expect(result.category).toBe('route');
    expect(result.recoverable).toBe(false);
    expect(result.retryable).toBe(false);
    expect(result.statusCode).toBe(404);
  });

  it('classifies 500 as server', () => {
    const result = classifyVyzorError({ status: 500 });
    expect(result.category).toBe('server');
    expect(result.retryable).toBe(true);
    expect(result.statusCode).toBe(500);
  });

  it('classifies 503 as server', () => {
    const result = classifyVyzorError({ status: 503 });
    expect(result.category).toBe('server');
    expect(result.retryable).toBe(true);
    expect(result.statusCode).toBe(503);
  });

  it('classifies network error from message', () => {
    const result = classifyVyzorError(new Error('Network request failed'));
    expect(result.category).toBe('network');
    expect(result.recoverable).toBe(true);
    expect(result.retryable).toBe(true);
  });

  it('classifies timeout as network', () => {
    const result = classifyVyzorError(new Error('Request timeout'));
    expect(result.category).toBe('network');
    expect(result.retryable).toBe(true);
  });

  it('classifies abort as network', () => {
    const result = classifyVyzorError(new Error('Request aborted'));
    expect(result.category).toBe('network');
  });

  it('classifies ChunkLoadError as network', () => {
    const chunkError = new Error('Loading chunk failed');
    chunkError.name = 'ChunkLoadError';
    const result = classifyVyzorError(chunkError);
    expect(result.category).toBe('network');
    expect(result.retryable).toBe(true);
  });

  it('classifies 408 as network', () => {
    const result = classifyVyzorError({ status: 408 });
    expect(result.category).toBe('network');
  });

  it('classifies 429 as network', () => {
    const result = classifyVyzorError({ status: 429 });
    expect(result.category).toBe('network');
  });

  it('classifies 504 as network', () => {
    const result = classifyVyzorError({ status: 504 });
    expect(result.category).toBe('network');
  });

  it('classifies unknown errors', () => {
    const result = classifyVyzorError(new Error('Something weird'));
    expect(result.category).toBe('unknown');
    expect(result.retryable).toBe(true);
  });

  it('classifies string errors', () => {
    const result = classifyVyzorError('plain string error');
    expect(result.category).toBe('unknown');
  });

  it('classifies null as unknown', () => {
    const result = classifyVyzorError(null);
    expect(result.category).toBe('unknown');
  });

  it('includes statusCode when present', () => {
    const result = classifyVyzorError({ status: 502 });
    expect(result.statusCode).toBe(502);
    expect(result.category).toBe('server');
  });
});

describe('classifyRenderError', () => {
  it('returns render category', () => {
    const result = classifyRenderError(new Error('render boom'));
    expect(result.category).toBe('render');
    expect(result.recoverable).toBe(false);
    expect(result.retryable).toBe(false);
  });
});

describe('classifyRouteError', () => {
  it('classifies 404 as route', () => {
    const result = classifyRouteError({ status: 404 });
    expect(result.category).toBe('route');
    expect(result.statusCode).toBe(404);
  });

  it('classifies non-404 route errors as route', () => {
    const result = classifyRouteError(new Error('route error'));
    expect(result.category).toBe('route');
    expect(result.recoverable).toBe(false);
  });
});
