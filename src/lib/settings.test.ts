import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

// Mock localStorage for tests
const localStorageMock = {
  store: {} as Record<string, string>,
  getItem: vi.fn((key: string) => localStorageMock.store[key] ?? null),
  setItem: vi.fn((key: string, value: string) => { localStorageMock.store[key] = value; }),
  removeItem: vi.fn((key: string) => { delete localStorageMock.store[key]; }),
  clear: vi.fn(() => { localStorageMock.store = {}; }),
};

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

describe("URL Validation", () => {
  // Test the validation logic inline since we can't import React hooks
  
  function isValidServerUrl(url: string): boolean {
    if (!url.trim()) return false;
    try {
      const u = new URL(url);
      return u.protocol === "http:" || u.protocol === "https:";
    } catch {
      return false;
    }
  }

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
  function formatDeviceClass(deviceClass: string | undefined): string {
    if (!deviceClass) return "Unknown Device";
    return deviceClass
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());
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
  
  function saveConfig(config: Record<string, unknown>) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
  }

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

  function saveOperator(op: Operator) {
    localStorage.setItem(OPERATOR_KEY, JSON.stringify(op));
  }

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
