import { useEffect, useRef, useState } from "react";

import { logger } from "@/lib/logger";
import type { TelemetryFrame } from "@/lib/vyzorix-api";
import { wsUrl } from "@/lib/vyzorix-config";

export type WsState = "connecting" | "connected" | "reconnecting" | "disconnected" | "idle";

export interface DeviceStreamState {
  state: WsState;
  lastTelemetry: TelemetryFrame | null;
  telemetryHistory: TelemetryFrame[];
  error?: string;
}

const HISTORY_LIMIT = 240;

// eslint-disable-next-line func-style
export function useDeviceStream(
  serverUrl: string,
  deviceId: string,
  enabled: boolean = true,
): DeviceStreamState {
  const [state, setState] = useState<WsState>("idle");
  const [lastTelemetry, setLast] = useState<TelemetryFrame | null>(null);
  const [history, setHistory] = useState<TelemetryFrame[]>([]);
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

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const connect = () => {
      const url = wsUrl(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/stream`);
      if (!url) return;
      setState(retryRef.current === 0 ? "connecting" : "reconnecting");
      logger.info("ws", `connect → ${url}`);
      let ws: WebSocket;
      try {
        ws = new WebSocket(url);
      } catch (e) {
        setError(String(e));
        setState("disconnected");
        logger.error("ws", `construct failed · ${String(e)}`);
        scheduleRetry();
        return;
      }
      wsRef.current = ws;

      ws.onopen = () => {
        retryRef.current = 0;
        setState("connected");
        setError(undefined);
        logger.info("ws", `open · device=${deviceId}`);
      };
      ws.onmessage = (ev) => {
        try {
          const frame = JSON.parse(ev.data);
          if (frame.type === "telemetry") {
            setLast(frame);
            setHistory((prev) => [...prev.slice(-(HISTORY_LIMIT - 1)), frame]);
            logger.debug(
              "ws",
              `telemetry risk=${frame.riskScore ?? "-"} buf=${frame.bufferLevel ?? "-"} temp=${frame.thermalTemp ?? "-"}`,
            );
          } else if (frame.type === "command") {
            logger.info(
              "ws",
              `command echo · ${frame.command} (${String(frame.dispatchId).slice(0, 8)})`,
            );
          } else if (frame.type === "ack") {
            logger.info(
              "ws",
              `ack · ${String(frame.dispatchId ?? "").slice(0, 8)} ${frame.status ?? ""}`,
            );
          } else {
            logger.debug("ws", `frame · ${String(ev.data).slice(0, 200)}`);
          }
        } catch {
          logger.warn("ws", `non-JSON frame · ${String(ev.data).slice(0, 120)}`);
        }
      };
      ws.onerror = () => {
        logger.error("ws", "socket error");
      };
      ws.onclose = (ev) => {
        logger.warn("ws", `closed code=${ev.code} reason=${ev.reason || "—"}`);
        setState("disconnected");
        scheduleRetry();
      };
    };

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
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

  return { state, lastTelemetry, telemetryHistory: history, error };
}
