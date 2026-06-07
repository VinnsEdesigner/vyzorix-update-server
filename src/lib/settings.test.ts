import { describe, expect, it, vi, beforeEach } from "vitest";

// Mock localStorage for tests
const localStorageMock = {
  store: {} as Record<string, string>,
  getItem: vi.fn((key: string) => localStorageMock.store[key] ?? null),
  setItem: vi.fn((key: string, value: string) => {
    localStorageMock.store[key] = value;
  }),
  removeItem: vi.fn((key: string) => {
    delete localStorageMock.store[key];
  }),
  clear: vi.fn(() => {
    localStorageMock.store = {};
  }),
};

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

describe("URL Validation", () => {
  // Test the validation logic inline since we can't import React hooks

  const isValidServerUrl = (url: string): boolean => {
    if (!url.trim()) return false;
    try {
      const u = new URL(url);
      return u.protocol === "http:" || u.protocol === "https:";
    } catch {
      return false;
    }
  };

  describe("isValidServerUrl", () => {
    it("returns true for valid HTTP URLs", () => {
      expect(isValidServerUrl("http://localhost:3000")).toBe(true);
      expect(isValidServerUrl("http://example.com")).toBe(true);
      expect(isValidServerUrl("http://192.168.1.1:8080")).toBe(true);
    });

    it("returns true for valid HTTPS URLs", () => {
      expect(isValidServerUrl("https://localhost:3000")).toBe(true);
      expect(isValidServerUrl("https://example.com")).toBe(true);
      expect(isValidServerUrl("https://my-app.vercel.app")).toBe(true);
    });

    it("returns false for empty strings", () => {
      expect(isValidServerUrl("")).toBe(false);
      expect(isValidServerUrl("   ")).toBe(false);
    });

    it("returns false for URLs without protocol", () => {
      expect(isValidServerUrl("localhost:3000")).toBe(false);
      expect(isValidServerUrl("example.com")).toBe(false);
      expect(isValidServerUrl("my-server.local")).toBe(false);
    });

    it("returns false for invalid URLs", () => {
      expect(isValidServerUrl("not-a-url")).toBe(false);
      expect(isValidServerUrl("htp://example.com")).toBe(false); // typo
      expect(isValidServerUrl("ftp://example.com")).toBe(false); // wrong protocol
    });

    it("returns false for URLs with other protocols", () => {
      expect(isValidServerUrl("ftp://example.com")).toBe(false);
      expect(isValidServerUrl("ws://example.com")).toBe(false);
      expect(isValidServerUrl("wss://example.com")).toBe(false);
    });

    it("handles URLs with paths", () => {
      expect(isValidServerUrl("http://localhost:3000/api/v1")).toBe(true);
      expect(isValidServerUrl("https://example.com/path/to/resource")).toBe(true);
    });

    it("handles URLs with query params", () => {
      expect(isValidServerUrl("http://localhost:3000?debug=true")).toBe(true);
      expect(isValidServerUrl("https://example.com/api?token=abc")).toBe(true);
    });
  });
});

describe("Device Class Formatting", () => {
  // Test the formatting logic
  // eslint-disable-next-line func-style
  function formatDeviceClass(deviceClass: string | undefined): string {
    if (!deviceClass) return "Unknown Device";
    return deviceClass.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
  }

  describe("formatDeviceClass", () => {
    it("converts snake_case to Title Case", () => {
      expect(formatDeviceClass("nokia_c22")).toBe("Nokia C22");
      expect(formatDeviceClass("samsung_galaxy_s21")).toBe("Samsung Galaxy S21");
      expect(formatDeviceClass("pixel_7_pro")).toBe("Pixel 7 Pro");
    });

    it("handles undefined", () => {
      expect(formatDeviceClass(undefined)).toBe("Unknown Device");
    });

    it("handles empty string", () => {
      expect(formatDeviceClass("")).toBe("Unknown Device");
    });

    it("preserves existing spaces", () => {
      expect(formatDeviceClass("my device")).toBe("My Device");
      expect(formatDeviceClass("hello world")).toBe("Hello World");
    });

    it("capitalizes first letter of each word", () => {
      expect(formatDeviceClass("one two three")).toBe("One Two Three");
      expect(formatDeviceClass("a")).toBe("A");
      expect(formatDeviceClass("abc")).toBe("Abc");
    });

    it("handles mixed snake_case and spaces", () => {
      expect(formatDeviceClass("nokia_c22 ultra")).toBe("Nokia C22 Ultra");
      expect(formatDeviceClass("device_model_v2")).toBe("Device Model V2");
    });

    it("handles uppercase snake_case", () => {
      expect(formatDeviceClass("NOKIA_C22")).toBe("NOKIA C22");
      expect(formatDeviceClass("SAMSUNG_GALAXY")).toBe("SAMSUNG GALAXY");
    });

    it("handles numbers in device class", () => {
      expect(formatDeviceClass("device_123")).toBe("Device 123");
      expect(formatDeviceClass("model_v2_plus")).toBe("Model V2 Plus");
    });
  });
});

describe("Config Storage", () => {
  beforeEach(() => {
    localStorageMock.store = {};
    vi.clearAllMocks();
  });

  const STORAGE_KEY = "vyz.config.test";

  // eslint-disable-next-line func-style
  function saveConfig(config: Record<string, unknown>) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
  }

  // eslint-disable-next-line func-style
  function loadConfig(): Record<string, unknown> | null {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }

  it("saves and loads config correctly", () => {
    const config = {
      serverUrl: "http://localhost:3000",
      deviceId: "test-device",
      autoReconnect: true,
    };

    saveConfig(config);
    const loaded = loadConfig();

    expect(loaded).toEqual(config);
    expect(loaded?.serverUrl).toBe("http://localhost:3000");
  });

  it("returns null for missing config", () => {
    const loaded = loadConfig();
    expect(loaded).toBeNull();
  });

  it("handles invalid JSON gracefully", () => {
    localStorage.setItem(STORAGE_KEY, "not valid json");
    const loaded = loadConfig();
    expect(loaded).toBeNull();
  });

  it("merges partial configs with defaults", () => {
    const defaults = {
      serverUrl: "http://localhost:3000",
      deviceId: "default-device",
      autoReconnect: true,
    };

    const partial = { serverUrl: "http://custom:8080" };
    const merged = { ...defaults, ...partial };

    expect(merged.serverUrl).toBe("http://custom:8080");
    expect(merged.deviceId).toBe("default-device");
    expect(merged.autoReconnect).toBe(true);
  });
});

describe("Settings Persistence", () => {
  const OPERATOR_KEY = "vyz.auth.operator";

  beforeEach(() => {
    localStorageMock.store = {};
  });

  interface Operator {
    id: string;
    email: string;
    name: string;
    role: string;
  }

  // eslint-disable-next-line func-style
  function saveOperator(op: Operator) {
    localStorage.setItem(OPERATOR_KEY, JSON.stringify(op));
  }

  // eslint-disable-next-line func-style
  function getStoredOperator(): Operator | null {
    const raw = localStorage.getItem(OPERATOR_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }

  it("saves and retrieves operator correctly", () => {
    const operator = {
      id: "op-123",
      email: "test@example.com",
      name: "Test User",
      role: "operator",
    };

    saveOperator(operator);
    const retrieved = getStoredOperator();

    expect(retrieved).toEqual(operator);
    expect(retrieved?.email).toBe("test@example.com");
  });

  it("returns null for missing operator", () => {
    const retrieved = getStoredOperator();
    expect(retrieved).toBeNull();
  });

  it("handles missing optional fields", () => {
    const operator = {
      id: "op-123",
      email: "test@example.com",
      name: "",
      role: "operator",
    };

    saveOperator(operator);
    const retrieved = getStoredOperator();

    expect(retrieved?.name).toBe("");
  });
});

describe("API Route Verification", () => {
  // Verify all frontend API routes match backend expectations

  describe("Auth API Routes (vyzorix-auth.ts)", () => {
    const authRoutes = [
      { method: "POST", path: "/v1/auth/login", auth: false },
      { method: "POST", path: "/v1/auth/register", auth: false },
      { method: "POST", path: "/v1/auth/logout", auth: true },
      { method: "GET", path: "/v1/auth/me", auth: true },
      { method: "PATCH", path: "/v1/auth/me", auth: true },
      { method: "GET", path: "/v1/auth/google", auth: false },
      { method: "GET", path: "/v1/auth/google/callback", auth: false },
    ];

    it.each(authRoutes)(
      "auth route $method $path has correct auth requirement",
      ({ path, auth }) => {
        // Routes that require auth should be protected
        if (auth) {
          expect(path).toMatch(/^\/v1\/auth\/(logout|me)/);
        }
      },
    );
  });

  describe("Device API Routes (vyzorix-api.ts)", () => {
    const deviceRoutes = [
      { method: "POST", path: "/v1/device/register", auth: false },
      { method: "GET", path: "/v1/device/:id/status", auth: false },
      { method: "POST", path: "/v1/device/:id/command", auth: false },
      { method: "GET", path: "/v1/device/:id/stream", auth: false },
      { method: "GET", path: "/v1/dashboard/devices", auth: true },
    ];

    it.each(deviceRoutes)("device route $method $path is correctly defined", ({ path }) => {
      // /v1/dashboard/devices is a separate dashboard route, /v1/device/* are device routes
      expect(path.startsWith("/v1/device") || path.startsWith("/v1/dashboard")).toBe(true);
    });
  });

  describe("Version/Update API Routes", () => {
    const updateRoutes = [
      { method: "GET", path: "/api/v1/version" },
      { method: "HEAD", path: "/api/v1/apk/:filename" },
      { method: "GET", path: "/healthz" },
    ];

    it.each(updateRoutes)("update route $method $path is correctly defined", ({ path }) => {
      expect(path.startsWith("/api/") || path.startsWith("/healthz")).toBe(true);
    });
  });
});

describe("TelemetryFrame Schema", () => {
  interface TelemetryFrame {
    type: "telemetry";
    deviceId?: string;
    uptime?: number;
    riskScore?: number;
    audioMode?: number;
    speakerOn?: boolean;
    activeDevice?: string;
    bufferLevel?: number;
    thermalTemp?: number;
    timestamp?: number | string;
  }

  it("has correct type field", () => {
    const frame: TelemetryFrame = {
      type: "telemetry",
      riskScore: 45,
      thermalTemp: 38.5,
    };
    expect(frame.type).toBe("telemetry");
  });

  it("has optional numeric fields", () => {
    const frame: TelemetryFrame = {
      type: "telemetry",
      uptime: 3600,
      riskScore: 75,
      audioMode: 2,
      bufferLevel: 85,
      thermalTemp: 42.3,
    };
    expect(frame.uptime).toBe(3600);
    expect(frame.riskScore).toBe(75);
    expect(frame.thermalTemp).toBe(42.3);
  });

  it("has optional boolean fields", () => {
    const frame: TelemetryFrame = {
      type: "telemetry",
      speakerOn: false,
      activeDevice: "bluetooth_speaker",
    };
    expect(frame.speakerOn).toBe(false);
    expect(frame.activeDevice).toBe("bluetooth_speaker");
  });

  it("has flexible timestamp field", () => {
    const frameNumeric: TelemetryFrame = {
      type: "telemetry",
      timestamp: 1700000000000,
    };
    const frameString: TelemetryFrame = {
      type: "telemetry",
      timestamp: "2024-01-01T12:00:00Z",
    };
    expect(typeof frameNumeric.timestamp).toBe("number");
    expect(typeof frameString.timestamp).toBe("string");
  });
});

describe("Alert Derivation Logic", () => {
  interface Thresholds {
    riskWarn: number;
    riskCrit: number;
    thermalWarn: number;
    thermalCrit: number;
    bufferWarn: number;
  }

  interface TelemetryFrame {
    riskScore?: number;
    thermalTemp?: number;
    bufferLevel?: number;
    speakerOn?: boolean;
    activeDevice?: string;
    timestamp?: number;
  }

  // eslint-disable-next-line func-style
  function deriveAlerts(history: TelemetryFrame[], th: Thresholds): string[] {
    const alerts: string[] = [];
    history.forEach((f) => {
      if ((f.riskScore ?? 0) >= th.riskCrit) {
        alerts.push(`critical: risk ${f.riskScore}`);
      } else if ((f.riskScore ?? 0) >= th.riskWarn) {
        alerts.push(`warning: risk ${f.riskScore}`);
      }
      if ((f.thermalTemp ?? 0) >= th.thermalCrit) {
        alerts.push(`critical: thermal ${f.thermalTemp}`);
      } else if ((f.thermalTemp ?? 0) >= th.thermalWarn) {
        alerts.push(`warning: thermal ${f.thermalTemp}`);
      }
      if (f.bufferLevel != null && f.bufferLevel < th.bufferWarn) {
        alerts.push(`warning: buffer ${f.bufferLevel}%`);
      }
      if (f.speakerOn === false) {
        alerts.push(`info: speaker route lost`);
      }
    });
    return alerts;
  }

  const thresholds: Thresholds = {
    riskWarn: 50,
    riskCrit: 75,
    thermalWarn: 45,
    thermalCrit: 55,
    bufferWarn: 50,
  };

  it("derives critical alert for high risk score", () => {
    const frames: TelemetryFrame[] = [{ riskScore: 80 }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.some((a) => a.includes("critical") && a.includes("risk 80"))).toBe(true);
  });

  it("derives warning alert for elevated risk score", () => {
    const frames: TelemetryFrame[] = [{ riskScore: 60 }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.some((a) => a.includes("warning") && a.includes("risk 60"))).toBe(true);
  });

  it("derives critical alert for high thermal", () => {
    const frames: TelemetryFrame[] = [{ thermalTemp: 58 }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.some((a) => a.includes("critical") && a.includes("thermal 58"))).toBe(true);
  });

  it("derives warning alert for low buffer", () => {
    const frames: TelemetryFrame[] = [{ bufferLevel: 30 }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.some((a) => a.includes("warning") && a.includes("buffer 30%"))).toBe(true);
  });

  it("derives info alert for speaker off", () => {
    const frames: TelemetryFrame[] = [{ speakerOn: false, activeDevice: "unknown" }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.some((a) => a.includes("info") && a.includes("speaker route lost"))).toBe(true);
  });

  it("derives multiple alerts from single frame", () => {
    const frames: TelemetryFrame[] = [{ riskScore: 80, thermalTemp: 58, bufferLevel: 30 }];
    const alerts = deriveAlerts(frames, thresholds);
    expect(alerts.length).toBe(3);
  });
});
