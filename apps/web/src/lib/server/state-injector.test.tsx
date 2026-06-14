/**
 * state-injector.test.tsx - Tests for SSR state injection utilities
 *
 * Tests:
 * 1. generateStateScript - script tag generation
 * 2. injectStateIntoHtml - state injection into HTML
 * 3. injectStateAfterRoot - alternate injection method
 * 4. getHydratedState - client-side state reading
 * 5. getFullHydratedState - full state retrieval
 */

import { describe, expect, it, beforeEach, afterEach } from "vitest";

import {
  generateStateScript,
  injectStateIntoHtml,
  injectStateAfterRoot,
  getHydratedState,
  getFullHydratedState,
  type HydratedState,
} from "./state-injector";

describe("generateStateScript", () => {
  const mockState: HydratedState = {
    isAuthenticated: true,
    operator: {
      id: "op_123",
      email: "test@example.com",
      name: "Test User",
      role: "operator",
    },
  };

  it("should generate valid script tag", () => {
    const script = generateStateScript(mockState);

    expect(script).toContain('<script id="__vyzorix-prefetched-state__"');
    expect(script).toContain('type="application/json"');
  });

  it("should contain window.__VYZORIX_PREFETCHED_STATE__ assignment", () => {
    const script = generateStateScript(mockState);

    expect(script).toContain("window.__VYZORIX_PREFETCHED_STATE__");
  });

  it("should JSON stringify the state", () => {
    const script = generateStateScript(mockState);

    expect(script).toContain('"isAuthenticated":true');
    expect(script).toContain('"email":"test@example.com"');
  });

  it("should handle unauthenticated state", () => {
    const unauthState: HydratedState = {
      isAuthenticated: false,
      operator: null,
    };
    const script = generateStateScript(unauthState);

    expect(script).toContain('"isAuthenticated":false');
    expect(script).toContain('"operator":null');
  });

  it("should handle operator with all fields", () => {
    const fullState: HydratedState = {
      isAuthenticated: true,
      operator: {
        id: "op_456",
        email: "admin@example.com",
        name: "Admin User",
        role: "super_admin",
      },
    };
    const script = generateStateScript(fullState);

    expect(script).toContain('"role":"super_admin"');
  });

  it("should escape special characters in JSON", () => {
    const stateWithSpecialChars: HydratedState = {
      isAuthenticated: true,
      operator: {
        id: "op_789",
        email: 'test"special@example.com',
        name: "Test <script>alert('xss')</script>",
        role: "operator",
      },
    };
    const script = generateStateScript(stateWithSpecialChars);

    // Should be valid JSON (quotes escaped)
    expect(script).toContain('"email":"test\\"special@example.com"');
  });
});

describe("injectStateIntoHtml", () => {
  const mockState: HydratedState = {
    isAuthenticated: true,
    operator: {
      id: "op_123",
      email: "test@example.com",
      name: "Test User",
      role: "operator",
    },
  };

  it("should replace app-state placeholder with script", () => {
    const html = '<html><body><div id="app"></div><!--app-state--></body></html>';
    const result = injectStateIntoHtml(html, mockState);

    // The placeholder should be replaced (not contained)
    expect(result).not.toContain("<!--app-state--><!--app-state-->");
    expect(result).toContain('<script id="__vyzorix-prefetched-state__"');
  });

  it("should inject state after replacement", () => {
    const html = "<html><body><!--app-state--></body></html>";
    const result = injectStateIntoHtml(html, mockState);

    expect(result).toContain("window.__VYZORIX_PREFETCHED_STATE__");
  });

  it("should handle missing placeholder gracefully", () => {
    const html = "<html><body><div>No placeholder</div></body></html>";
    const result = injectStateIntoHtml(html, mockState);

    // Should return original HTML without modification (no placeholder to replace)
    expect(result).toBe(html);
  });

  it("should return empty string for empty HTML (no placeholder)", () => {
    const result = injectStateIntoHtml("", mockState);
    // No placeholder to replace, so returns original empty string
    expect(result).toBe("");
  });

  it("should handle HTML with multiple placeholders", () => {
    const html = "<div><!--app-state--></div><!--app-state-->";
    const result = injectStateIntoHtml(html, mockState);

    // Only first occurrence replaced
    expect(result).toContain("__vyzorix-prefetched-state__");
  });

  it("should preserve rest of HTML", () => {
    const html = "<html><head><title>Test</title></head><!--app-state--><body></body></html>";
    const result = injectStateIntoHtml(html, mockState);

    expect(result).toContain("<title>Test</title>");
    expect(result).toContain("<html>");
    expect(result).toContain("</html>");
  });
});

describe("injectStateAfterRoot", () => {
  const mockState: HydratedState = {
    isAuthenticated: false,
    operator: null,
  };

  it("should inject after div#root", () => {
    const html = '<div id="root"><!--@tanstack/start-entry--></div>';
    const result = injectStateAfterRoot(html, mockState);

    expect(result).toContain('<div id="root">');
    expect(result).toContain("window.__VYZORIX_PREFETCHED_STATE__");
  });

  it("should inject after div#app", () => {
    const html = '<div id="app"><!--@tanstack/start-entry--></div>';
    const result = injectStateAfterRoot(html, mockState);

    expect(result).toContain('<div id="app">');
    expect(result).toContain("window.__VYZORIX_PREFETCHED_STATE__");
  });

  it("should handle both id patterns", () => {
    const html = '<div id="app"><!--app--></div>';
    const result = injectStateAfterRoot(html, mockState);

    // Should inject after app div
    expect(result).not.toContain('<div id="app"><script');
    expect(result).toContain('<div id="app">');
    expect(result).toContain("__VYZORIX_PREFETCHED_STATE__");
  });

  it("should return original if no root div", () => {
    const html = "<div>No root</div>";
    const result = injectStateAfterRoot(html, mockState);

    expect(result).toBe(html);
  });
});

describe("getHydratedState", () => {
  const mockState: HydratedState = {
    isAuthenticated: true,
    operator: {
      id: "op_123",
      email: "test@example.com",
      name: "Test User",
      role: "operator",
    },
  };

  // Helper to set window.__VYZORIX_PREFETCHED_STATE__
  const setWindowState = (state: HydratedState | undefined) => {
    Object.defineProperty(global, "window", {
      value: {
        ...(global.window as object),
        __VYZORIX_PREFETCHED_STATE__: state,
      },
      writable: true,
      configurable: true,
    });
  };

  beforeEach(() => {
    // Reset to undefined state before each test
    setWindowState(undefined);
  });

  afterEach(() => {
    // Reset window
    setWindowState(undefined);
  });

  it("should return hydrated value when state exists", () => {
    setWindowState(mockState);

    expect(getHydratedState("isAuthenticated", false)).toBe(true);
  });

  it("should return default value when state is missing", () => {
    expect(getHydratedState("isAuthenticated", false)).toBe(false);
  });

  it("should return default for missing key in state", () => {
    setWindowState(mockState);
    expect(getHydratedState("nonexistent", "default")).toBe("default");
  });

  it("should return default when key value is undefined", () => {
    setWindowState({ isAuthenticated: true, operator: null });
    expect(getHydratedState("operator", null)).toBe(null);
  });

  it("should work with different types", () => {
    setWindowState(mockState);
    expect(getHydratedState("isAuthenticated", false)).toBe(true);
    expect(getHydratedState("operator", null)).toEqual(mockState.operator);
  });

  it("should handle SSR (no window)", () => {
    // Save original window descriptor
    const descriptor = Object.getOwnPropertyDescriptor(global, "window");
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const originalWindow = global.window as unknown as Record<string, unknown>;

    // Create a new window-like object without __VYZORIX_PREFETCHED_STATE__
    const newWindow: Record<string, unknown> = {};
    for (const key of Object.keys(originalWindow)) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      newWindow[key] = (originalWindow as any)[key];
    }
    Object.defineProperty(global, "window", {
      value: newWindow,
      writable: true,
      configurable: true,
    });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    delete (global as any).__VYZORIX_PREFETCHED_STATE__;

    expect(getHydratedState("isAuthenticated", false)).toBe(false);

    // Restore original window
    Object.defineProperty(global, "window", descriptor!);
  });
});

describe("getFullHydratedState", () => {
  const mockState: HydratedState = {
    isAuthenticated: true,
    operator: {
      id: "op_123",
      email: "test@example.com",
      name: "Test User",
      role: "operator",
    },
  };

  // Helper to set window.__VYZORIX_PREFETCHED_STATE__
  const setWindowState = (state: HydratedState | undefined) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const existingWindow = global.window as unknown as Record<string, unknown>;
    const newWindow: Record<string, unknown> = {};
    for (const key of Object.keys(existingWindow)) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      newWindow[key] = (existingWindow as any)[key];
    }
    newWindow.__VYZORIX_PREFETCHED_STATE__ = state;
    Object.defineProperty(global, "window", {
      value: newWindow,
      writable: true,
      configurable: true,
    });
  };

  beforeEach(() => {
    // Reset to undefined state before each test
    setWindowState(undefined);
  });

  afterEach(() => {
    // Reset window
    setWindowState(undefined);
  });

  it("should return full state when available", () => {
    setWindowState(mockState);
    const result = getFullHydratedState();
    expect(result).toEqual(mockState);
  });

  it("should return null when no state", () => {
    const result = getFullHydratedState();
    expect(result).toBeNull();
  });

  it("should return null when state is undefined", () => {
    const result = getFullHydratedState();
    expect(result).toBeNull();
  });

  it("should handle SSR (no window)", () => {
    // Save original window descriptor
    const descriptor = Object.getOwnPropertyDescriptor(global, "window");
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const originalWindow = global.window as unknown as Record<string, unknown>;

    // Create a new window-like object without __VYZORIX_PREFETCHED_STATE__
    const newWindow: Record<string, unknown> = {};
    for (const key of Object.keys(originalWindow)) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      newWindow[key] = (originalWindow as any)[key];
    }
    Object.defineProperty(global, "window", {
      value: newWindow,
      writable: true,
      configurable: true,
    });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    delete (global as any).__VYZORIX_PREFETCHED_STATE__;

    const result = getFullHydratedState();
    expect(result).toBeNull();

    // Restore original window
    Object.defineProperty(global, "window", descriptor!);
  });

  it("should return state (not null) for empty object", () => {
    // Empty object is truthy in JS, so it returns the object not null
    setWindowState({} as HydratedState);
    const result = getFullHydratedState();
    expect(result).toEqual({});
  });
});
