import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
  type ReactElement,
} from "react";

export const DEFAULT_SERVER_URL = "http://localhost:3000";
// No default device - must be selected from registered devices or empty
export const DEFAULT_DEVICE_ID = "";

const STORAGE_KEY = "vyzorix.config.v2";

// Settings that come from the server — do NOT persist to localStorage.
const SERVER_KEYS = ["autoReconnect", "strictHmac", "notificationsEnabled", "thresholds"] as const;

type ServerKey = (typeof SERVER_KEYS)[number];

export interface Thresholds {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

export interface Operator {
  name: string;
  role: "viewer" | "operator" | "super_admin";
  email: string;
}

export interface VyzorixSettings {
  serverUrl: string;
  deviceId: string;
  autoReconnect: boolean;
  requestTimeoutMs: number;
  logBufferLimit: number;
  signalHistoryLimit: number;
  strictHmac: boolean;
  dashboardToken: string;
  notificationsEnabled: boolean;
  operator: Operator;
  thresholds: Thresholds;
}

export const DEFAULT_SETTINGS: VyzorixSettings = {
  serverUrl: DEFAULT_SERVER_URL,
  deviceId: DEFAULT_DEVICE_ID,
  autoReconnect: true,
  requestTimeoutMs: 8000,
  logBufferLimit: 500,
  signalHistoryLimit: 240,
  strictHmac: false,
  dashboardToken: "",
  notificationsEnabled: true,
  operator: { name: "", role: "operator", email: "" },
  thresholds: {
    riskWarn: 50,
    riskCrit: 75,
    thermalWarn: 45,
    thermalCrit: 55,
    bufferWarn: 50,
    bufferCrit: 80,
  },
};

const loadInitial = (): VyzorixSettings => {
  if (typeof window === "undefined") return DEFAULT_SETTINGS;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULT_SETTINGS;
    const parsed = JSON.parse(raw);
    return {
      ...DEFAULT_SETTINGS,
      ...parsed,
      operator: { ...DEFAULT_SETTINGS.operator, ...(parsed.operator ?? {}) },
      // Server-backed settings fall back to defaults if not in localStorage
      thresholds: { ...DEFAULT_SETTINGS.thresholds, ...(parsed.thresholds ?? {}) },
    } as VyzorixSettings;
  } catch {
    return DEFAULT_SETTINGS;
  }
};

type Config = VyzorixSettings & {
  setServerUrl: (v: string) => void;
  setDeviceId: (v: string) => void;
  update: (patch: Partial<VyzorixSettings>) => void;
  reset: () => void;
};

const ConfigCtx = createContext<Config | null>(null);

export const VyzorixConfigProvider = ({ children }: { children: ReactNode }): ReactElement => {
  // Lazy init: read localStorage BEFORE first paint so consumers never see defaults
  // followed by a hydration swap (this was causing settings to "reset" visually
  // when navigating between pages).
  const [s, setS] = useState<VyzorixSettings>(loadInitial);

  // Only persist browser-local settings to localStorage.
  // Server-backed settings (thresholds, strictHmac, autoReconnect, notificationsEnabled)
  // are loaded from the server on mount and should not pollute localStorage.
  useEffect(() => {
    try {
      const local: Partial<VyzorixSettings> = {};
      for (const [k, v] of Object.entries(s)) {
        if (!SERVER_KEYS.includes(k as ServerKey)) {
          (local as Record<string, unknown>)[k] = v;
        }
      }
      localStorage.setItem(STORAGE_KEY, JSON.stringify(local));
    } catch {
      // ignore storage error
    }
  }, [s]);

  useEffect(() => {
    const syncOperator = (): void => {
      try {
        const raw = localStorage.getItem("vyz.auth.operator");
        if (!raw) return;
        const stored = JSON.parse(raw) as {
          id: string;
          email: string;
          name: string;
          role: string;
          createdAt: number;
        };
        setS((prev) => {
          if (
            prev.operator.email === stored.email &&
            prev.operator.name === stored.name &&
            prev.operator.role === stored.role
          ) {
            return prev;
          }
          return {
            ...prev,
            operator: {
              name: stored.name,
              role: stored.role as "viewer" | "operator" | "super_admin",
              email: stored.email,
            },
          };
        });
      } catch {
        // ignore parse/storage error
      }
    };
    window.addEventListener("vyz.operator.updated", syncOperator);
    return () => window.removeEventListener("vyz.operator.updated", syncOperator);
  }, []);

  const update = (patch: Partial<VyzorixSettings>): void => setS((prev) => ({ ...prev, ...patch }));
  const setServerUrl = (v: string): void => update({ serverUrl: v });
  const setDeviceId = (v: string): void => update({ deviceId: v });
  const reset = (): void => setS(DEFAULT_SETTINGS);

  return (
    <ConfigCtx.Provider value={{ ...s, setServerUrl, setDeviceId, update, reset }}>
      {children}
    </ConfigCtx.Provider>
  );
};

export const useVyzorixConfig = (): Config => {
  const ctx = useContext(ConfigCtx);
  if (!ctx) throw new Error("useVyzorixConfig must be used inside VyzorixConfigProvider");
  return ctx;
};

export const wsUrl = (serverUrl: string, path: string): string => {
  try {
    const u = new URL(path, serverUrl);
    u.protocol = u.protocol === "https:" ? "wss:" : "ws:";
    return u.toString();
  } catch {
    return "";
  }
};
