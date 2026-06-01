import { useEffect, useRef, useState } from "react";
import { wsUrl } from "@/lib/vyzorix-config";
import type { TelemetryFrame } from "@/lib/vyzorix-api";

export type WsState = "connecting" | "connected" | "reconnecting" | "disconnected" | "idle";

export interface StreamLogEntry {
  t: number;
  level: "info" | "warn" | "error";
  text: string;
}

export interface DeviceStreamState {
  state: WsState;
  lastTelemetry: TelemetryFrame | null;
  telemetryHistory: TelemetryFrame[];
  logs: StreamLogEntry[];
  error?: string;
}

const HISTORY_LIMIT = 240;
const LOG_LIMIT = 500;

export function useDeviceStream(serverUrl: string, deviceId: string, enabled: boolean = true): DeviceStreamState {
  const [state, setState] = useState<WsState>("idle");
  const [lastTelemetry, setLast] = useState<TelemetryFrame | null>(null);
  const [history, setHistory] = useState<TelemetryFrame[]>([]);
  const [logs, setLogs] = useState<StreamLogEntry[]>([]);
  const [error, setError] = useState<string | undefined>(undefined);
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef(0);
  const stopRef = useRef(false);

  useEffect(() => {
    if (!enabled || !serverUrl || !deviceId) {
      setState("idle");
      return;
    }
    stopRef.current = false;

    const log = (level: StreamLogEntry["level"], text: string) =>
      setLogs((prev) => [...prev.slice(-(LOG_LIMIT - 1)), { t: Date.now(), level, text }]);

    const connect = () => {
      const url = wsUrl(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/stream`);
      if (!url) return;
      setState(retryRef.current === 0 ? "connecting" : "reconnecting");
      log("info", `WS connect → ${url}`);
      let ws: WebSocket;
      try {
        ws = new WebSocket(url);
      } catch (e) {
        setError(String(e));
        setState("disconnected");
        scheduleRetry();
        return;
      }
      wsRef.current = ws;

      ws.onopen = () => {
        retryRef.current = 0;
        setState("connected");
        setError(undefined);
        log("info", `WS open · device=${deviceId}`);
      };
      ws.onmessage = (ev) => {
        try {
          const frame = JSON.parse(ev.data);
          if (frame.type === "telemetry") {
            setLast(frame);
            setHistory((prev) => [...prev.slice(-(HISTORY_LIMIT - 1)), frame]);
            log("info", `telemetry risk=${frame.riskScore ?? "-"} buf=${frame.bufferLevel ?? "-"} temp=${frame.thermalTemp ?? "-"}`);
          } else if (frame.type === "command") {
            log("info", `command echo · ${frame.command} (${frame.dispatchId})`);
          } else if (frame.type === "ack") {
            log("info", `ack · ${frame.dispatchId ?? ""} ${frame.status ?? ""}`);
          } else {
            log("info", `frame · ${ev.data.slice(0, 200)}`);
          }
        } catch {
          log("warn", `non-JSON frame · ${String(ev.data).slice(0, 120)}`);
        }
      };
      ws.onerror = () => {
        log("error", "WS error");
      };
      ws.onclose = (ev) => {
        log("warn", `WS closed code=${ev.code} reason=${ev.reason || "—"}`);
        setState("disconnected");
        scheduleRetry();
      };
    };

    const scheduleRetry = () => {
      if (stopRef.current) return;
      retryRef.current = Math.min(retryRef.current + 1, 6);
      const delay = Math.min(1000 * 2 ** retryRef.current, 15000);
      setTimeout(() => {
        if (!stopRef.current) connect();
      }, delay);
    };

    connect();

    return () => {
      stopRef.current = true;
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [serverUrl, deviceId, enabled]);

  return { state, lastTelemetry, telemetryHistory: history, logs, error };
}