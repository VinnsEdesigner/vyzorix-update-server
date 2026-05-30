import { createContext, useContext, useEffect, useState, type ReactNode } from "react";

export const DEFAULT_SERVER_URL = "http://localhost:8080";
export const DEFAULT_DEVICE_ID = "nokia-c22-primary";

const SERVER_KEY = "vyzorix.serverUrl";
const DEVICE_KEY = "vyzorix.deviceId";

type Config = {
  serverUrl: string;
  deviceId: string;
  setServerUrl: (v: string) => void;
  setDeviceId: (v: string) => void;
};

const ConfigCtx = createContext<Config | null>(null);

export function VyzorixConfigProvider({ children }: { children: ReactNode }) {
  const [serverUrl, setServerUrlState] = useState(DEFAULT_SERVER_URL);
  const [deviceId, setDeviceIdState] = useState(DEFAULT_DEVICE_ID);

  useEffect(() => {
    try {
      const s = localStorage.getItem(SERVER_KEY);
      const d = localStorage.getItem(DEVICE_KEY);
      if (s) setServerUrlState(s);
      if (d) setDeviceIdState(d);
    } catch {}
  }, []);

  const setServerUrl = (v: string) => {
    setServerUrlState(v);
    try { localStorage.setItem(SERVER_KEY, v); } catch {}
  };
  const setDeviceId = (v: string) => {
    setDeviceIdState(v);
    try { localStorage.setItem(DEVICE_KEY, v); } catch {}
  };

  return (
    <ConfigCtx.Provider value={{ serverUrl, deviceId, setServerUrl, setDeviceId }}>
      {children}
    </ConfigCtx.Provider>
  );
}

export function useVyzorixConfig() {
  const ctx = useContext(ConfigCtx);
  if (!ctx) throw new Error("useVyzorixConfig must be used inside VyzorixConfigProvider");
  return ctx;
}

export function wsUrl(serverUrl: string, path: string) {
  try {
    const u = new URL(path, serverUrl);
    u.protocol = u.protocol === "https:" ? "wss:" : "ws:";
    return u.toString();
  } catch {
    return "";
  }
}