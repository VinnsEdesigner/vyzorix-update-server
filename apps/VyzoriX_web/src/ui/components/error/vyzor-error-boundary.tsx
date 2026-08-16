import { Component, type ErrorInfo, type ReactNode } from 'react';
import { classifyRenderError } from '@/lib/error';
import { useVyzorErrorStore } from '@/stores/vyzor-error-store';
import { VyzorErrorFallback } from './vyzor-error-fallback';

interface VyzorErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface VyzorErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class VyzorErrorBoundary extends Component<
  VyzorErrorBoundaryProps,
  VyzorErrorBoundaryState
> {
  override state: VyzorErrorBoundaryState = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): VyzorErrorBoundaryState {
    return { hasError: true, error };
  }

  override componentDidCatch(error: Error, _info: ErrorInfo): void {
    const classification = classifyRenderError(error);
    useVyzorErrorStore.getState().report(error, classification, 'render');
  }

  handleDismiss = (): void => {
    this.setState({ hasError: false, error: null });
  };

  override render(): ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <VyzorErrorFallback
          classification={classifyRenderError(this.state.error)}
          message={this.state.error?.message}
          onDismiss={this.handleDismiss}
        />
      );
    }
    return this.props.children;
  }
}
