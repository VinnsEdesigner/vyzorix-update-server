import { AlertTriangle, RefreshCw, Copy, Check } from "lucide-react";
import React, { useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export interface GraphQLError {
  message: string;
  code?: string;
  path?: string[];
  extensions?: Record<string, unknown>;
}

export interface GraphQLErrorDisplayProps {
  error: GraphQLError | GraphQLError[] | null;
  title?: string;
  onRetry?: () => void;
  className?: string;
}

/**
 * Formats a GraphQL error for display.
 */
function formatError(error: GraphQLError): string {
  if (error.path && error.path.length > 0) {
    return `${error.message} (path: ${error.path.join(".")})`;
  }
  return error.message;
}

/**
 * Gets a user-friendly error title based on the error code.
 */
function getErrorTitle(code?: string): string {
  switch (code) {
    case "UNAUTHORIZED":
      return "Authentication Required";
    case "FORBIDDEN":
      return "Access Denied";
    case "NOT_FOUND":
      return "Not Found";
    case "VALIDATION_ERROR":
      return "Validation Error";
    case "RATE_LIMITED":
      return "Rate Limited";
    case "INTERNAL_ERROR":
      return "Server Error";
    case undefined:
    default:
      return "GraphQL Error";
  }
}

/**
 * Gets the severity/variant based on error code.
 */
function getErrorVariant(code?: string): "default" | "destructive" {
  switch (code) {
    case "UNAUTHORIZED":
    case "FORBIDDEN":
    case "RATE_LIMITED":
    case "INTERNAL_ERROR":
      return "destructive";
    case undefined:
    default:
      return "default";
  }
}

/**
 * GraphQLErrorDisplay displays GraphQL errors with retry functionality.
 */
export function GraphQLErrorDisplay({
  error,
  title,
  onRetry,
  className,
}: GraphQLErrorDisplayProps) {
  const [copied, setCopied] = useState(false);

  if (!error) return null;

  const errors = Array.isArray(error) ? error : [error];

  const handleCopy = async () => {
    const text = errors.map(formatError).join("\n");
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={className}>
      {errors.map((err, index) => {
        const code = err.code ?? (err.extensions?.code as string | undefined);
        const variant = getErrorVariant(code);
        const errorTitle = title ?? getErrorTitle(code);

        return (
          <Alert key={index} variant={variant} className="mb-2">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle className="flex items-center gap-2">
              {errorTitle}
              {code && <span className="text-xs font-normal text-muted-foreground">({code})</span>}
            </AlertTitle>
            <AlertDescription className="mt-2">
              <code className="text-xs bg-muted px-1 py-0.5 rounded">{err.message}</code>
              {err.path && err.path.length > 0 && (
                <p className="mt-1 text-xs text-muted-foreground">Field: {err.path.join(".")}</p>
              )}
            </AlertDescription>
          </Alert>
        );
      })}

      <div className="mt-3 flex gap-2">
        {onRetry && (
          <Button variant="outline" size="sm" onClick={onRetry}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        )}
        <Button variant="ghost" size="sm" onClick={handleCopy}>
          {copied ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
          {copied ? "Copied" : "Copy Error"}
        </Button>
      </div>
    </div>
  );
}

/**
 * GraphQL loading skeleton for use during data fetching.
 */
export function GraphQLLoading({ rows = 3, className }: { rows?: number; className?: string }) {
  return (
    <div className={className}>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="mb-3 animate-pulse rounded-md border border-border bg-muted/50 p-4">
          <div className="h-4 w-1/3 rounded bg-muted" />
          <div className="mt-2 h-3 w-2/3 rounded bg-muted" />
        </div>
      ))}
    </div>
  );
}

/**
 * GraphQLErrorBoundary provides error handling for GraphQL queries.
 */
interface GraphQLErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  onError?: (error: GraphQLError) => void;
}

interface GraphQLErrorBoundaryState {
  hasError: boolean;
  error: GraphQLError | null;
}

export class GraphQLErrorBoundary extends React.Component<
  GraphQLErrorBoundaryProps,
  GraphQLErrorBoundaryState
> {
  constructor(props: GraphQLErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): GraphQLErrorBoundaryState {
    return {
      hasError: true,
      error: {
        message: error.message ?? "An unexpected error occurred",
        code: "INTERNAL_ERROR",
      },
    };
  }

  override componentDidCatch(error: Error): void {
    this.props.onError?.({
      message: error.message ?? "An unexpected error occurred",
      code: "INTERNAL_ERROR",
    });
  }

  override render(): React.ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <GraphQLErrorDisplay
          error={this.state.error}
          onRetry={() => this.setState({ hasError: false, error: null })}
        />
      );
    }
    return this.props.children;
  }
}

export default GraphQLErrorDisplay;
