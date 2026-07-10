/**
 * WebSocket Client
 * 
 * WebSocket client for real-time device communication.
 * Based on REALTIME_WEBSOCKET_ARCHITECTURE.md.
 */

// ============================================================================
// Types
// ============================================================================

/**
 * WebSocket connection states
 */
export type WebSocketState = "connecting" | "connected" | "disconnected" | "error";

/**
 * Dashboard â Server message types
 */
export type DashboardMessageType = 
  | "AUTH" 
  | "SUBSCRIBE" 
  | "UNSUBSCRIBE" 
  | "COMMAND";

/**
 * Server â Dashboard message types
 */
export type ServerMessageType = 
  | "AUTH_ACK" 
  | "TELEMETRY" 
  | "EVENT" 
  | "COMMAND_ACK" 
  | "ERROR";

/**
 * Telemetry frame from device
 */
export interface WSTelemetryFrame {
  timestamp: number;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs?: number;
}

/**
 * Device event from server
 */
export interface WSDeviceEvent {
  type: string;
  deviceId: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

/**
 * Command to send to device
 */
export interface WSCommand {
  dispatchId: string;
  command: string;
  parameters?: Record<string, unknown>;
}

/**
 * WebSocket message
 */
export interface WSMessage<T = unknown> {
  type: string;
  payload: T;
  timestamp?: number;
}

/**
 * Client configuration
 */
export interface WSClientConfig {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

// ============================================================================
// Client
// ============================================================================

/**
 * WebSocket client for real-time communication
 */
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private config: Required<WSClientConfig>;
  private state: WebSocketState = "disconnected";
  private reconnectAttempts = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private subscriptions = new Set<string>();
  private messageHandlers = new Map<string, Set<(payload: unknown) => void>>();
  private stateHandlers = new Set<(state: WebSocketState) => void>();

  constructor(config: WSClientConfig) {
    this.config = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      heartbeatInterval: 30000,
      ...config,
    };
  }

  // ============================================================================
  // Connection Management
  // ============================================================================

  /**
   * Connect to WebSocket server
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.setState("connecting");

      try {
        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          this.setState("connected");
          this.reconnectAttempts = 0;
          this.startHeartbeat();
          this.resubscribeAll();
          resolve();
        };

        this.ws.onclose = () => {
          this.setState("disconnected");
          this.stopHeartbeat();
          this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
          this.setState("error");
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };
      } catch (error) {
        this.setState("error");
        reject(error);
      }
    });
  }

  /**
   * Disconnect from server
   */
  disconnect(): void {
    this.stopHeartbeat();
    this.reconnectAttempts = this.config.maxReconnectAttempts; // Prevent auto-reconnect
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.setState("disconnected");
  }

  /**
   * Authenticate with the server
   */
  authenticate(token: string): Promise<void> {
    return this.send("AUTH", { token });
  }

  // ============================================================================
  // Subscriptions
  // ============================================================================

  /**
   * Subscribe to device events
   */
  subscribe(deviceId: string): Promise<void> {
    this.subscriptions.add(deviceId);
    
    if (this.state === "connected") {
      return this.send("SUBSCRIBE", { deviceId });
    }
    
    return Promise.resolve();
  }

  /**
   * Unsubscribe from device events
   */
  unsubscribe(deviceId: string): Promise<void> {
    this.subscriptions.delete(deviceId);
    
    if (this.state === "connected") {
      return this.send("UNSUBSCRIBE", { deviceId });
    }
    
    return Promise.resolve();
  }

  /**
   * Subscribe to all events (dashboard-wide)
   */
  subscribeAll(): Promise<void> {
    return this.send("SUBSCRIBE", { deviceId: "*" });
  }

  // ============================================================================
  // Commands
  // ============================================================================

  /**
   * Send command to a device
   */
  sendCommand(command: WSCommand): Promise<void> {
    return this.send("COMMAND", command);
  }

  // ============================================================================
  // Event Handlers
  // ============================================================================

  /**
   * Add message handler for a specific message type
   */
  onMessage<T>(type: string, handler: (payload: T) => void): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    
    this.messageHandlers.get(type)!.add(handler as (payload: unknown) => void);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.get(type)?.delete(handler as (payload: unknown) => void);
    };
  }

  /**
   * Add state change handler
   */
  onStateChange(handler: (state: WebSocketState) => void): () => void {
    this.stateHandlers.add(handler);
    
    // Return unsubscribe function
    return () => {
      this.stateHandlers.delete(handler);
    };
  }

  /**
   * Convenience: Handle telemetry
   */
  onTelemetry(handler: (frame: WSTelemetryFrame) => void): () => void {
    return this.onMessage<WSTelemetryFrame>("TELEMETRY", handler);
  }

  /**
   * Convenience: Handle device events
   */
  onEvent(handler: (event: WSDeviceEvent) => void): () => void {
    return this.onMessage<WSDeviceEvent>("EVENT", handler);
  }

  /**
   * Convenience: Handle command acknowledgments
   */
  onCommandAck(handler: (ack: { dispatchId: string; status: string; result?: unknown }) => void): () => void {
    return this.onMessage("COMMAND_ACK", handler);
  }

  // ============================================================================
  // State Access
  // ============================================================================

  /**
   * Get current connection state
   */
  getState(): WebSocketState {
    return this.state;
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.state === "connected";
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private setState(state: WebSocketState): void {
    this.state = state;
    this.stateHandlers.forEach((handler) => handler(state));
  }

  private send(type: string, payload: unknown): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error("WebSocket not connected"));
        return;
      }

      const message: WSMessage = {
        type,
        payload,
        timestamp: Date.now(),
      };

      this.ws.send(JSON.stringify(message));
      resolve();
    });
  }

  private handleMessage(data: string): void {
    try {
      const message: WSMessage = JSON.parse(data);
      
      // Handle AUTH_ACK specially
      if (message.type === "AUTH_ACK") {
        const handlers = this.messageHandlers.get("AUTH_ACK");
        handlers?.forEach((handler) => handler(message.payload));
        return;
      }

      // Forward to type-specific handlers
      const handlers = this.messageHandlers.get(message.type);
      handlers?.forEach((handler) => handler(message.payload));
      
      // Also emit to wildcard handlers
      const wildcardHandlers = this.messageHandlers.get("*");
      wildcardHandlers?.forEach((handler) => handler(message));
    } catch (error) {
      console.error("Failed to parse WebSocket message:", error);
    }
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: "PING", timestamp: Date.now() }));
      }
    }, this.config.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.log("Max reconnect attempts reached");
      return;
    }

    this.reconnectAttempts++;
    console.log(`Attempting reconnect ${this.reconnectAttempts}/${this.config.maxReconnectAttempts}`);

    setTimeout(() => {
      this.connect().catch(() => {
        // Will trigger another reconnect attempt via onclose
      });
    }, this.config.reconnectInterval);
  }

  private resubscribeAll(): void {
    this.subscriptions.forEach((deviceId) => {
      this.send("SUBSCRIBE", { deviceId }).catch(console.error);
    });
  }
}

// ============================================================================
// Factory Function
// ============================================================================

/**
 * Create a WebSocket client with default config
 */
export function createWebSocketClient(config?: Partial<WSClientConfig>): WebSocketClient {
  const wsUrl = config?.url ?? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/ws`;
  
  return new WebSocketClient({
    url: wsUrl,
    ...config,
  });
}