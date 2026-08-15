import { useState, type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, type RenderHookOptions } from '@testing-library/react';

export function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
      mutations: { retry: false },
    },
  });
}

function Wrapper({ children }: { children: ReactNode }) {
  const [client] = useState(() => createTestQueryClient());
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

export function renderHookWithQueryClient<T, P>(hook: (props: P) => T, options?: RenderHookOptions<P>) {
  return renderHook(hook, { wrapper: Wrapper, ...options });
}

export { renderHook };
