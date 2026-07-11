import type { WebSocketState } from "@/domain/realtime";

export type ConnectionState = "CLOSED" | "CONNECTING" | "OPEN" | "CLOSING" | "RECONNECTING";

export interface ConnectionStateMachine {
  state: ConnectionState;
  transition(newState: ConnectionState): void;
  onStateChange(handler: (state: ConnectionState) => void): () => void;
}

export class ConnectionStateMachineImpl implements ConnectionStateMachine {
  state: ConnectionState = "CLOSED";
  private handlers = new Set<(state: ConnectionState) => void>();

  transition(newState: ConnectionState): void {
    if (this.state === newState) return;
    this.state = newState;
    this.handlers.forEach((handler) => handler(newState));
  }

  onStateChange(handler: (state: ConnectionState) => void): () => void {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }
}

export function stateFromReadyState(readyState: number): ConnectionState {
  switch (readyState) {
    case WebSocket.CONNECTING:
      return "CONNECTING";
    case WebSocket.OPEN:
      return "OPEN";
    case WebSocket.CLOSING:
      return "CLOSING";
    case WebSocket.CLOSED:
    default:
      return "CLOSED";
  }
}
