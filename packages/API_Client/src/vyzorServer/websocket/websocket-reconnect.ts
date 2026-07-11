export interface ReconnectConfig {
  initialDelay: number;
  maxDelay: number;
  maxAttempts: number;
  multiplier: number;
}

const DEFAULT_CONFIG: ReconnectConfig = {
  initialDelay: 1000,
  maxDelay: 30000,
  maxAttempts: 5,
  multiplier: 2,
};

export interface ReconnectManager {
  scheduleReconnect(): void;
  reset(): void;
  onReconnecting(handler: (attempt: number) => void): () => void;
  getAttempts(): number;
}

export class ReconnectManagerImpl implements ReconnectManager {
  private config: ReconnectConfig;
  private attempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectHandlers = new Set<(attempt: number) => void>();
  private shouldReconnect = true;
  private connectFn: () => Promise<void>;

  constructor(connectFn: () => Promise<void>, config: Partial<ReconnectConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.connectFn = connectFn;
  }

  scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    if (!this.shouldReconnect || this.attempts >= this.config.maxAttempts) {
      return;
    }

    this.attempts++;
    const delay = this.calculateDelay();

    this.reconnectHandlers.forEach((handler) => handler(this.attempts));

    this.reconnectTimer = setTimeout(async () => {
      try {
        await this.connectFn();
      } catch {
        // Will trigger another schedule via disconnect
      }
    }, delay);
  }

  reset(): void {
    this.attempts = 0;
    this.shouldReconnect = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  stopReconnecting(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  onReconnecting(handler: (attempt: number) => void): () => void {
    this.reconnectHandlers.add(handler);
    return () => this.reconnectHandlers.delete(handler);
  }

  getAttempts(): number {
    return this.attempts;
  }

  private calculateDelay(): number {
    const exponentialDelay = this.config.initialDelay * Math.pow(this.config.multiplier, this.attempts - 1);
    return Math.min(exponentialDelay, this.config.maxDelay);
  }
}
