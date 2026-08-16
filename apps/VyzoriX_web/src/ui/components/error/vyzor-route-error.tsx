import { type ReactNode } from 'react';
import { classifyRouteError } from '@/lib/error';
import { VyzorErrorFallback } from './vyzor-error-fallback';

interface VyzorRouteErrorProps {
  error: unknown;
}

export function VyzorRouteError({ error }: VyzorRouteErrorProps): ReactNode {
  const classification = classifyRouteError(error);
  const message =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
        ? error
        : classification.category === 'route'
          ? 'The page you are looking for does not exist.'
          : 'An error occurred while loading this page.';

  return (
    <VyzorErrorFallback classification={classification} message={message} />
  );
}
