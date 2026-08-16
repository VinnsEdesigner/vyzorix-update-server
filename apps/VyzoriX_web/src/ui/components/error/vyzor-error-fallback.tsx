import { useVyzorErrorRecovery } from '@/hooks/error';
import type { VyzorErrorClassification } from '@/lib/error';

interface VyzorErrorFallbackProps {
  classification: VyzorErrorClassification;
  message?: string;
  onRetry?: () => void;
  onReload?: () => void;
  onGoHome?: () => void;
  onDismiss?: () => void;
}

const CATEGORY_TITLES: Record<string, string> = {
  network: 'Connection Error',
  auth: 'Authentication Required',
  server: 'Server Error',
  render: 'Something Went Wrong',
  route: 'Page Not Found',
  unknown: 'Unexpected Error',
};

export function VyzorErrorFallback({
  classification,
  message,
  onRetry,
  onReload,
  onGoHome,
  onDismiss,
}: VyzorErrorFallbackProps) {
  const recovery = useVyzorErrorRecovery();
  const title = CATEGORY_TITLES[classification.category] ?? 'Error';
  const handleRetry = onRetry ?? recovery.retry;
  const handleReload = onReload ?? recovery.reload;
  const handleGoHome = onGoHome ?? recovery.goHome;
  const handleDismiss = onDismiss ?? recovery.dismiss;

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="flex min-h-[400px] flex-col items-center justify-center gap-4 p-8 text-center"
    >
      <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
      {message && <p className="text-muted-foreground max-w-md">{message}</p>}
      {classification.statusCode && (
        <p className="text-muted-foreground text-sm">
          Status: {classification.statusCode}
        </p>
      )}
      <div className="flex flex-wrap items-center justify-center gap-2">
        {classification.retryable && (
          <button
            type="button"
            onClick={handleRetry}
            className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-2 text-sm font-medium"
          >
            Retry
          </button>
        )}
        <button
          type="button"
          onClick={handleReload}
          className="bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-md px-4 py-2 text-sm font-medium"
        >
          Reload
        </button>
        <button
          type="button"
          onClick={handleGoHome}
          className="text-muted-foreground hover:text-foreground text-sm font-medium underline"
        >
          Go Home
        </button>
        {onDismiss && (
          <button
            type="button"
            onClick={handleDismiss}
            className="text-muted-foreground hover:text-foreground text-sm font-medium"
          >
            Dismiss
          </button>
        )}
      </div>
    </div>
  );
}
