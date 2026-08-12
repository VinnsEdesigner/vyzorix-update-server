import {
  telemetryFromRaw,
  wsEventFromRaw,
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
import { ConnectionStateMachineImpl } from "./websocket-connection";
import { HeartbeatManagerImpl } from "./websocket-heartbeat";
import { ReconnectManagerImpl } from "./websocket-reconnect";
import { parseWSMessage, type WSMessage } from "./websocket-messages";
import { getWebSocketConfig } from "../config";
import { signWebSocketConnect, generateNonce, generateTimestamp } from "../crypto";

export interface WSDeviceCredentials {
  deviceId: string;
  secret: string;
}

export interface WSClientConfig {
  url?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  heartbeatTimeout?: number;
  deviceCredentials?: WSDeviceCredentials;
}

export interface SubscriptionResult {
  success: boolean;
  deviceId: string;
  error?: string;
}

export interface WebSocketClient {
  connect(): Promise<void>;
  disconnect(): void;
  subscribe(deviceId: string): Promise<SubscriptionResult>;
  unsubscribe(deviceId: string): Promise<SubscriptionResult>;
  subscribeAll(): Promise<SubscriptionResult>;
  sendCommand(command: { dispatchId: string; command: string; parameters?: Record<string, unknown> }): Promise<void>;
  onTelemetry(handler: (telemetry: WSTelemetry) => void): () => void;
  onEvent(handler: (event: WSEvent) => void): () => void;
  onCommandAck(handler: (ack: WSCommandAck) => void): () => void;
  onAuthAck(handler: (response: { success: boolean; error?: string }) => void): () => void;
  onSubscribeAck(handler: (result: SubscriptionResult) => void): () => void;
  onUnsubscribeAck(handler: (result: SubscriptionResult) => void): () => void;
  onStateChange(handler: (state: string) => void): () => void;
  getState(): string;
  isConnected(): boolean;
  getSubscriptions(): string[];
  setCredentials(credentials: WSDeviceCredentials): void;
  clearCredentials(): void;
}

export class WebSocketClientImpl implements WebSocketClient {
  private ws: WebSocket | null = null;
  private config: Required<Omit<WSClientConfig, 'deviceCredentials'>>;
  private credentials: WSDeviceCredentials | null = null;
  private stateMachine: ConnectionStateMachineImpl;
  private heartbeat: HeartbeatManagerImpl;
  private reconnect: ReconnectManagerImpl;

  private subscriptions = new Set<string>();
  private pendingSubscriptions = new Map<string, { resolve: (result: SubscriptionResult) => void; timeout: ReturnType<typeof setTimeout> }>();
  private pendingUnsubscriptions = new Map<string, { resolve: (result: SubscriptionResult) => void; timeout: ReturnType<typeof setTimeout> }>();

  private telemetryHandlers = new Set<(t: WSTelemetry) => void>();
  private eventHandlers = new Set<(e: WSEvent) => void>();
  private commandAckHandlers = new Set<(a: WSCommandAck) => void>();
  private authAckHandlers = new Set<(r: { success: boolean; error?: string }) => void>();
  private subscribeAckHandlers = new Set<(r: SubscriptionResult) => void>();
  private unsubscribeAckHandlers = new Set<(r: SubscriptionResult) => void>();
  private stateHandlers = new Set<(s: string) => void>();

  private static readonly SUBSCRIBE_TIMEOUT_MS = 5000;

  constructor(config: WSClientConfig = {}) {
    
    const envConfig = getWebSocketConfig();
    this.config = {
      url: config.url ?? envConfig.url,
      reconnectInterval: config.reconnectInterval ?? envConfig.reconnectInterval,
      maxReconnectAttempts: config.maxReconnectAttempts ?? envConfig.maxReconnectAttempts,
      heartbeatInterval: config.heartbeatInterval ?? envConfig.heartbeatInterval,
      heartbeatTimeout: config.heartbeatTimeout ?? envConfig.heartbeatTimeout,
    };
    this.credentials = config.deviceCredentials ?? null;
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

  setCredentials(credentials: WSDeviceCredentials): void {
    this.credentials = credentials;
  }

  clearCredentials(): void {
    this.credentials = null;
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

  async subscribe(deviceId: string): Promise<SubscriptionResult> {
    const result = await this.sendSubscription(deviceId, "SUBSCRIBE");
    if (result.success) {
      this.subscriptions.add(deviceId);
    }
    return result;
  }

  async unsubscribe(deviceId: string): Promise<SubscriptionResult> {
    const result = await this.sendSubscription(deviceId, "UNSUBSCRIBE");
    if (result.success) {
      this.subscriptions.delete(deviceId);
    }
    return result;
  }

  async subscribeAll(): Promise<SubscriptionResult> {
    return this.subscribe("*");
  }

  getSubscriptions(): string[] {
    return Array.from(this.subscriptions);
  }

  private async sendSubscription(deviceId: string, type: "SUBSCRIBE" | "UNSUBSCRIBE"): Promise<SubscriptionResult> {
    return new Promise<SubscriptionResult>((resolve) => {
      if (this.stateMachine.state !== "OPEN") {
        const result = { success: false, deviceId, error: "WebSocket not connected" };
        resolve(result);
        return;
      }

      const pendingMap = type === "SUBSCRIBE" ? this.pendingSubscriptions : this.pendingUnsubscriptions;
      
      const timeout = setTimeout(() => {
        pendingMap.delete(deviceId);
        resolve({ success: false, deviceId, error: `Subscription timeout after ${WebSocketClientImpl.SUBSCRIBE_TIMEOUT_MS}ms` });
      }, WebSocketClientImpl.SUBSCRIBE_TIMEOUT_MS);

      pendingMap.set(deviceId, { resolve, timeout });

      try {
        this.ws?.send(JSON.stringify({ type, payload: { deviceId }, timestamp: Date.now() }));
      } catch (error) {
        clearTimeout(timeout);
        pendingMap.delete(deviceId);
        resolve({ success: false, deviceId, error: String(error) });
      }
    });
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

  onSubscribeAck(handler: (result: SubscriptionResult) => void): () => void {
    this.subscribeAckHandlers.add(handler);
    return () => this.subscribeAckHandlers.delete(handler);
  }

  onUnsubscribeAck(handler: (result: SubscriptionResult) => void): () => void {
    this.unsubscribeAckHandlers.add(handler);
    return () => this.unsubscribeAckHandlers.delete(handler);
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

  private buildAuthenticatedUrl(): string {
    let url = this.config.url;
    
    if (this.credentials) {
      const timestamp = generateTimestamp();
      const nonce = generateNonce();
      const signature = signWebSocketConnect(
        this.credentials.deviceId,
        timestamp.toString(),
        nonce,
        this.credentials.secret
      );
      
      const separator = url.includes('?') ? '&' : '?';
      url = `${url}${separator}hmac_timestamp=${encodeURIComponent(timestamp)}&hmac_nonce=${encodeURIComponent(nonce)}&hmac_signature=${encodeURIComponent(signature)}&device_id=${encodeURIComponent(this.credentials.deviceId)}`;
    }
    
    return url;
  }

  private async connectInternal(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.stateMachine.transition("CONNECTING");
    this.reconnect.reset();

    return new Promise((resolve, reject) => {
      try {
        const authenticatedUrl = this.buildAuthenticatedUrl();
        this.ws = new WebSocket(authenticatedUrl);

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
        this.eventHandlers.forEach((h) => h(wsEventFromRaw(message.payload as RawWSEvent)));
        break;
      case "COMMAND_ACK":
        this.commandAckHandlers.forEach((h) => h(commandAckFromRaw(message.payload as RawWSCommandAck)));
        break;
      case "AUTH_ACK":
        this.authAckHandlers.forEach((h) => h(authResponseFromRaw(message.payload as RawWSAuthResponse)));
        break;
      case "SUBSCRIBE_ACK":
        this.handleSubscribeAck(message.payload as { success: boolean; deviceId: string; error?: string });
        break;
      case "UNSUBSCRIBE_ACK":
        this.handleUnsubscribeAck(message.payload as { success: boolean; deviceId: string; error?: string });
        break;
      default:
        // Unknown message type - ignore for forward compatibility
        break;
    }
  }

  private handleSubscribeAck(payload: { success: boolean; deviceId: string; error?: string }): void {
    const pending = this.pendingSubscriptions.get(payload.deviceId);
    if (pending) {
      clearTimeout(pending.timeout);
      this.pendingSubscriptions.delete(payload.deviceId);
      pending.resolve(payload);
    }

    if (payload.success) {
      this.subscriptions.add(payload.deviceId);
    }

    this.subscribeAckHandlers.forEach((h) => h(payload));
  }

  private handleUnsubscribeAck(payload: { success: boolean; deviceId: string; error?: string }): void {
    const pending = this.pendingUnsubscriptions.get(payload.deviceId);
    if (pending) {
      clearTimeout(pending.timeout);
      this.pendingUnsubscriptions.delete(payload.deviceId);
      pending.resolve(payload);
    }

    if (payload.success) {
      this.subscriptions.delete(payload.deviceId);
    }

    this.unsubscribeAckHandlers.forEach((h) => h(payload));
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
      try {
        await this.send({ type: "SUBSCRIBE", payload: { deviceId } });
      } catch {
        // Ignore errors during resubscription - will retry on next connect
      }
    }
  }
}

export function createWebSocketClient(config: WSClientConfig): WebSocketClient {
  return new WebSocketClientImpl(config);
}
