import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { useQueryClient } from '@tanstack/react-query';
import { QueryProvider } from '@/providers/query-provider';

function Consumer() {
  const client = useQueryClient();
  return <div data-testid="client-id">{client ? 'has-client' : 'no-client'}</div>;
}

describe('QueryProvider', () => {
  it('renders children', () => {
    const { getByText } = render(
      <QueryProvider>
        <span>child</span>
      </QueryProvider>,
    );
    expect(getByText('child')).toBeInTheDocument();
  });

  it('provides a QueryClient to children', () => {
    const { getByTestId } = render(
      <QueryProvider>
        <Consumer />
      </QueryProvider>,
    );
    expect(getByTestId('client-id').textContent).toBe('has-client');
  });

  it('uses the same client instance across re-renders', () => {
    const { rerender, getByTestId } = render(
      <QueryProvider>
        <Consumer />
      </QueryProvider>,
    );
    const firstId = getByTestId('client-id').textContent;
    rerender(
      <QueryProvider>
        <Consumer />
      </QueryProvider>,
    );
    expect(getByTestId('client-id').textContent).toBe(firstId);
  });
});
