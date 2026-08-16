export type VyzorErrorCategory =
  | 'network'
  | 'auth'
  | 'server'
  | 'render'
  | 'route'
  | 'unknown';

export interface VyzorErrorClassification {
  category: VyzorErrorCategory;
  recoverable: boolean;
  retryable: boolean;
  statusCode?: number;
}

export interface VyzorReportedError {
  id: string;
  error: Error | null;
  message: string;
  classification: VyzorErrorClassification;
  context?: string;
  timestamp: number;
  retryCount: number;
}
