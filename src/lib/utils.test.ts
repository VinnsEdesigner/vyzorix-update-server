import { describe, expect, it } from "vitest";
import { cn } from "./utils";

describe("cn (class name merger)", () => {
  it("merges class names", () => {
    const result = cn("foo", "bar");
    expect(result).toBe("foo bar");
  });

  it("handles empty inputs", () => {
    const result = cn();
    expect(result).toBe("");
  });

  it("merges conflicting tailwind classes", () => {
    const result = cn("px-2 px-4");
    // tailwind-merge should handle duplicate conflicts
    expect(result).toBe("px-4");
  });

  it("handles conditional classes", () => {
    const isActive = true;
    const result = cn("base-class", isActive && "active-class", !isActive && "inactive-class");
    expect(result).toBe("base-class active-class");
  });

  it("handles clsx classValue objects", () => {
    const result = cn({ "foo": true, "bar": false });
    expect(result).toBe("foo");
  });

  it("handles mixed inputs", () => {
    const result = cn("foo", { bar: true, baz: false }, "qux");
    expect(result).toBe("foo bar qux");
  });
});