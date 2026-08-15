import { vi } from 'vitest';

export type MockFn = ReturnType<typeof vi.fn>;

export function createMockEndpoint(methods: Record<string, MockFn>) {
  return methods;
}

export function mockFn<T extends (...args: never[]) => unknown>(impl?: T): ReturnType<typeof vi.fn> {
  return vi.fn(impl as never);
}
