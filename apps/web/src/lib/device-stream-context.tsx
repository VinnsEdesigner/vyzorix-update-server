import { createContext, useContext, type ReactNode, type ReactElement } from "react";

import { useDeviceStream, type DeviceStreamState } from "@/hooks/use-device-stream";

import { useVyzorixConfig } from "./vyzorix-config";

// Singleton wrapper: the WebSocket and log/history buffers are owned by the
// app layout, so every page reads from the SAME live connection. Without this,
// each page mounted its own useDeviceStream() and opened a second WS — the
// constant reconnect/race showed up in the UI as a flashing connection badge.
const Ctx = createContext<DeviceStreamState | null>(null);

export const DeviceStreamProvider = ({ children }: { children: ReactNode }): ReactElement => {
  const { serverUrl, deviceId: imei, autoReconnect } = useVyzorixConfig();
  const state = useDeviceStream(serverUrl, imei, autoReconnect);
  return <Ctx.Provider value={state}>{children}</Ctx.Provider>;
};

// eslint-disable-next-line react-refresh/only-export-components
export const useStream = (): DeviceStreamState => {
  const v = useContext(Ctx);
  if (!v) throw new Error("useStream must be used inside DeviceStreamProvider");
  return v;
};
