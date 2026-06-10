// @vyzorix/config/vitest - Test Setup and Utilities
// This file is automatically imported before each test file

import "@testing-library/jest-dom";

// Export test utilities
export const testUtils = {
  /**
   * Generate a unique ID for test isolation
   */
  generateId: () => `test-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,

  /**
   * Create a mock localStorage
   */
  mockLocalStorage: () => {
    const store: Record<string, string> = {};
    return {
      getItem: (key: string) => store[key] || null,
      setItem: (key: string, value: string) => { store[key] = value; },
      removeItem: (key: string) => { delete store[key]; },
      clear: () => { Object.keys(store).forEach(key => delete store[key]); },
      get length() { return Object.keys(store).length; },
      key: (index: number) => Object.keys(store)[index] || null,
    };
  },

  /**
   * Create mock fetch
   */
  mockFetch: (response: any, ok = true) => {
    return () =>
      Promise.resolve({
        ok,
        json: () => Promise.resolve(response),
        status: ok ? 200 : 400,
        headers: new Headers(),
      });
  },

  /**
   * Wait for a condition
   */
  waitFor: (condition: () => boolean, timeout = 1000) => {
    return new Promise<void>((resolve, reject) => {
      const startTime = Date.now();
      const check = () => {
        if (condition()) {
          resolve();
        } else if (Date.now() - startTime > timeout) {
          reject(new Error("Timeout waiting for condition"));
        } else {
          setTimeout(check, 10);
        }
      };
      check();
    });
  },
};

export default testUtils;