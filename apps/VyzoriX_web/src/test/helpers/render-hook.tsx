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

export function renderHookWithQueryClient<T, P>(
  hook: (props: P) => T,
  options?: RenderHookOptions<P> & { queryClient?: QueryClient },
) {
  const { queryClient: sharedClient, ...rest } = options ?? {};
  function WrapperWithClient({ children }: { children: ReactNode }) {
    const [client] = useState(() => sharedClient ?? createTestQueryClient());
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  }
  return renderHook(hook, { wrapper: WrapperWithClient, ...rest });
}

export { renderHook };
