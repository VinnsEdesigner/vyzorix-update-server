import {
  telemetryFromRaw,
  eventFromRaw,
  commandAckFromRaw,
  authResponseFromRaw,
  type WSTelemetry,
  type WSEvent,
  type WSCommandAck,
  type RawWSTelemetry,
  type RawWSEvent,
  type RawWSCommandAck,
  type RawWSAuthResponse,
} from "@/domain/realtime";
import { ConnectionStateMachineImpl, stateFromReadyState } from "./websocket-connection";
import { HeartbeatManagerImpl } from "./websocket-heartbeat";
import { ReconnectManagerImpl } from "./websocket-reconnect";
import { createWSMessage, parseWSMessage, type WSMessage } from "./websocket-messages";
import { getWebSocketConfig, type WebSocketConfig } from "../config";

export interface WSClientConfig {
  url?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  heartbeatTimeout?: number;
}

const DEFAULT_CONFIG: Required<WSClientConfig> = {
  url: "",
  reconnectInterval: 3000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000,
  heartbeatTimeout: 10000,
};

export interface WebSocketClient {
  connect(): Promise<void>;
  disconnect(): void;
  subscribe(deviceId: string): Promise<void>;
  unsubscribe(deviceId: string): Promise<void>;
  subscribeAll(): Promise<void>;
  sendCommand(command: { dispatchId: string; command: string; parameters?: Record<string, unknown> }): Promise<void>;
  onTelemetry(handler: (telemetry: WSTelemetry) => void): () => void;
  onEvent(handler: (event: WSEvent) => void): () => void;
  onCommandAck(handler: (ack: WSCommandAck) => void): () => void;
  onAuthAck(handler: (response: { success: boolean; error?: string }) => void): () => void;
  onStateChange(handler: (state: string) => void): () => void;
  getState(): string;
  isConnected(): boolean;
}

export class WebSocketClientImpl implements WebSocketClient {
  private ws: WebSocket | null = null;
  private config: Required<WSClientConfig>;
  private stateMachine: ConnectionStateMachineImpl;
  private heartbeat: HeartbeatManagerImpl;
  private reconnect: ReconnectManagerImpl;

  private subscriptions = new Set<string>();
  private telemetryHandlers = new Set<(t: WSTelemetry) => void>();
  private eventHandlers = new Set<(e: WSEvent) => void>();
  private commandAckHandlers = new Set<(a: WSCommandAck) => void>();
  private authAckHandlers = new Set<(r: { success: boolean; error?: string }) => void>();
  private stateHandlers = new Set<(s: string) => void>();

  constructor(config: WSClientConfig = {}) {
    
    const envConfig = getWebSocketConfig();
    const mergedConfig: Required<WSClientConfig> = {
      url: config.url ?? envConfig.url,
      reconnectInterval: config.reconnectInterval ?? envConfig.reconnectInterval,
      maxReconnectAttempts: config.maxReconnectAttempts ?? envConfig.maxReconnectAttempts,
      heartbeatInterval: config.heartbeatInterval ?? envConfig.heartbeatInterval,
      heartbeatTimeout: config.heartbeatTimeout ?? envConfig.heartbeatTimeout,
    };
    this.config = mergedConfig;
    this.stateMachine = new ConnectionStateMachineImpl();

    this.heartbeat = new HeartbeatManagerImpl(
      (timestamp) => this.sendRaw({ type: "PING", payload: { timestamp } }),
      { interval: this.config.heartbeatInterval, timeout: this.config.heartbeatTimeout }
    );

    this.reconnect = new ReconnectManagerImpl(
      () => this.connectInternal(),
      { 
        initialDelay: this.config.reconnectInterval, 
        maxAttempts: this.config.maxReconnectAttempts,
        maxDelay: envConfig.reconnectMaxDelay,
        multiplier: envConfig.reconnectMultiplier,
      }
    );

    this.stateMachine.onStateChange((state) => {
      this.stateHandlers.forEach((h) => h(state));
    });
  }

  async connect(): Promise<void> {
    return this.connectInternal();
  }

  disconnect(): void {
    this.reconnect.reset();
    this.heartbeat.stop();
    this.stateMachine.transition("CLOSED");

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  async subscribe(deviceId: string): Promise<void> {
    this.subscriptions.add(deviceId);
    if (this.stateMachine.state === "OPEN") {
      await this.send({ type: "SUBSCRIBE", payload: { deviceId } });
    }
  }

  async unsubscribe(deviceId: string): Promise<void> {
    this.subscriptions.delete(deviceId);
    if (this.stateMachine.state === "OPEN") {
      await this.send({ type: "UNSUBSCRIBE", payload: { deviceId } });
    }
  }

  async subscribeAll(): Promise<void> {
    await this.subscribe("*");
  }

  async sendCommand(command: { dispatchId: string; command: string; parameters?: Record<string, unknown> }): Promise<void> {
    await this.send({ type: "COMMAND", payload: command });
  }

  onTelemetry(handler: (telemetry: WSTelemetry) => void): () => void {
    this.telemetryHandlers.add(handler);
    return () => this.telemetryHandlers.delete(handler);
  }

  onEvent(handler: (event: WSEvent) => void): () => void {
    this.eventHandlers.add(handler);
    return () => this.eventHandlers.delete(handler);
  }

  onCommandAck(handler: (ack: WSCommandAck) => void): () => void {
    this.commandAckHandlers.add(handler);
    return () => this.commandAckHandlers.delete(handler);
  }

  onAuthAck(handler: (response: { success: boolean; error?: string }) => void): () => void {
    this.authAckHandlers.add(handler);
    return () => this.authAckHandlers.delete(handler);
  }

  onStateChange(handler: (state: string) => void): () => void {
    this.stateHandlers.add(handler);
    return () => this.stateHandlers.delete(handler);
  }

  getState(): string {
    return this.stateMachine.state;
  }

  isConnected(): boolean {
    return this.stateMachine.state === "OPEN";
  }

  private async connectInternal(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.stateMachine.transition("CONNECTING");
    this.reconnect.reset();

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          this.stateMachine.transition("OPEN");
          this.heartbeat.start();
          this.resubscribeAll();
          resolve();
        };

        this.ws.onclose = () => {
          this.heartbeat.stop();
          this.stateMachine.transition("CLOSED");
          this.reconnect.scheduleReconnect();
        };

        this.ws.onerror = (error) => {
          this.stateMachine.transition("CLOSED");
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };
      } catch (error) {
        this.stateMachine.transition("CLOSED");
        reject(error);
      }
    });
  }

  private handleMessage(data: string): void {
    const message = parseWSMessage(data);
    if (!message) return;

    switch (message.type) {
      case "PONG":
        this.heartbeat.onPong();
        break;
      case "TELEMETRY":
        this.telemetryHandlers.forEach((h) => h(telemetryFromRaw(message.payload as RawWSTelemetry)));
        break;
      case "EVENT":
        this.eventHandlers.forEach((h) => h(eventFromRaw(message.payload as RawWSEvent)));
        break;
      case "COMMAND_ACK":
        this.commandAckHandlers.forEach((h) => h(commandAckFromRaw(message.payload as RawWSCommandAck)));
        break;
      case "AUTH_ACK":
        this.authAckHandlers.forEach((h) => h(authResponseFromRaw(message.payload as RawWSAuthResponse)));
        break;
    }
  }

  private send<T>(message: WSMessage<T>): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error("WebSocket not connected"));
        return;
      }
      this.ws.send(JSON.stringify(message));
      resolve();
    });
  }

  private sendRaw(message: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  private async resubscribeAll(): Promise<void> {
    for (const deviceId of this.subscriptions) {
      await this.send({ type: "SUBSCRIBE", payload: { deviceId } }).catch(() => {});
    }
  }
}

export function createWebSocketClient(config: WSClientConfig): WebSocketClient {
  return new WebSocketClientImpl(config);
}
