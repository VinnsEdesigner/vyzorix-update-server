import { describe, expect, it } from "vitest";

import { formatUptime, formatRelative, formatBytes, shortHash } from "./format";

describe("formatUptime", () => {
  it("returns — for undefined", () => {
    expect(formatUptime()).toBe("—");
  });

  it("returns — for negative values", () => {
    expect(formatUptime(-1)).toBe("—");
  });

  it("formats minutes only (no leading zero for hours)", () => {
    expect(formatUptime(30)).toBe("0m");
    expect(formatUptime(59)).toBe("0m");
    expect(formatUptime(60)).toBe("1m"); // Omits 0h
  });

  it("formats hours and minutes", () => {
    expect(formatUptime(3660)).toBe("1h 1m");
    expect(formatUptime(7200)).toBe("2h 0m");
  });

  it("formats days, hours, and minutes", () => {
    expect(formatUptime(86400)).toBe("1d 0h 0m");
    expect(formatUptime(90061)).toBe("1d 1h 1m");
  });
});

describe("formatRelative", () => {
  it("returns — for undefined", () => {
    expect(formatRelative()).toBe("—");
  });

  it("returns — for null", () => {
    expect(formatRelative(null)).toBe("—");
  });

  it("returns — for empty string", () => {
    expect(formatRelative("")).toBe("—");
  });

  it("returns 'just now' for recent timestamps", () => {
    expect(formatRelative(Date.now())).toBe("just now");
    expect(formatRelative(Date.now() - 1000)).toBe("just now");
  });

  it("returns seconds ago", () => {
    const past = Date.now() - 30_000; // 30 seconds ago
    expect(formatRelative(past)).toBe("30s ago");
  });

  it("returns minutes ago", () => {
    const past = Date.now() - 120_000; // 2 minutes ago
    expect(formatRelative(past)).toBe("2m ago");
  });

  it("returns hours ago", () => {
    const past = Date.now() - 3_600_000; // 1 hour ago
    expect(formatRelative(past)).toBe("1h ago");
  });

  it("handles string date input", () => {
    const past = new Date(Date.now() - 60_000).toISOString(); // 1 minute ago
    expect(formatRelative(past)).toBe("1m ago");
  });
});

describe("formatBytes", () => {
  it("returns — for undefined", () => {
    expect(formatBytes()).toBe("—");
  });

  it("returns — for null", () => {
    expect(formatBytes(null as unknown as number)).toBe("—");
  });

  it("formats bytes", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(512)).toBe("512 B");
    expect(formatBytes(1023)).toBe("1023 B");
  });

  it("formats kilobytes", () => {
    expect(formatBytes(1024)).toBe("1.0 KB");
    expect(formatBytes(1536)).toBe("1.5 KB");
  });

  it("formats megabytes", () => {
    expect(formatBytes(1024 * 1024)).toBe("1.0 MB");
    expect(formatBytes(1024 * 1024 * 2.5)).toBe("2.5 MB");
  });

  it("formats gigabytes", () => {
    expect(formatBytes(1024 * 1024 * 1024)).toBe("1.00 GB");
    expect(formatBytes(1024 * 1024 * 1024 * 1.5)).toBe("1.50 GB");
  });
});

describe("shortHash", () => {
  it("returns — for undefined", () => {
    expect(shortHash()).toBe("—");
  });

  it("returns — for empty string", () => {
    expect(shortHash("")).toBe("—");
  });

  it("returns short hash with defaults", () => {
    // Default head=6, tail=6
    // "abcdefghijklmnop" is 16 chars, so it truncates
    expect(shortHash("abcdefghijklmnop")).toBe("abcdef…klmnop");
  });

  it("respects head and tail parameters", () => {
    expect(shortHash("abcdefghijklmnop", 4, 4)).toBe("abcd…mnop");
  });

  it("returns original string if short enough", () => {
    expect(shortHash("abc")).toBe("abc");
    expect(shortHash("abcdefgh", 4, 4)).toBe("abcdefgh");
  });

  it("handles exactly head+tail+1 length", () => {
    expect(shortHash("abcdefghi", 4, 4)).toBe("abcdefghi");
  });
});
