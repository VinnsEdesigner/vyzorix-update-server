export interface HeartbeatManager {
  start(): void;
  stop(): void;
  onMissedHeartbeat(handler: () => void): void;
  getLastPingTime(): Date | null;
  getRTT(): number | null;
}

export interface HeartbeatConfig {
  interval: number;
  timeout: number;
}

const DEFAULT_CONFIG: HeartbeatConfig = {
  interval: 30000,
  timeout: 10000,
};

export class HeartbeatManagerImpl implements HeartbeatManager {
  private config: HeartbeatConfig;
  private timer: ReturnType<typeof setInterval> | null = null;
  private timeoutTimer: ReturnType<typeof setTimeout> | null = null;
  private lastPingTime: Date | null = null;
  private lastPongTime: Date | null = null;
  private missedHandlers = new Set<() => void>();
  private sendPing: (timestamp: number) => void;

  constructor(sendPing: (timestamp: number) => void, config: Partial<HeartbeatConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.sendPing = sendPing;
  }

  start(): void {
    this.stop();
    this.lastPingTime = null;
    this.lastPongTime = null;

    this.timer = setInterval(() => {
      this.sendPingNow();
    }, this.config.interval);

    this.sendPingNow();
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    if (this.timeoutTimer) {
      clearTimeout(this.timeoutTimer);
      this.timeoutTimer = null;
    }
  }

  onPong(): void {
    this.lastPongTime = new Date();
    if (this.timeoutTimer) {
      clearTimeout(this.timeoutTimer);
      this.timeoutTimer = null;
    }
  }

  onMissedHeartbeat(handler: () => void): () => void {
    this.missedHandlers.add(handler);
    return () => this.missedHandlers.delete(handler);
  }

  getLastPingTime(): Date | null {
    return this.lastPingTime;
  }

  getRTT(): number | null {
    if (!this.lastPingTime || !this.lastPongTime) return null;
    return this.lastPongTime.getTime() - this.lastPingTime.getTime();
  }

  private sendPingNow(): void {
    this.lastPingTime = new Date();
    this.sendPing(Date.now());

    this.timeoutTimer = setTimeout(() => {
      this.missedHandlers.forEach((handler) => handler());
    }, this.config.timeout);
  }
}
