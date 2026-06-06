import { describe, expect, it } from "vitest";
import { cn, escapeHTML, stripHTML } from "./utils";

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
    const result = cn({ foo: true, bar: false });
    expect(result).toBe("foo");
  });

  it("handles mixed inputs", () => {
    const result = cn("foo", { bar: true, baz: false }, "qux");
    expect(result).toBe("foo bar qux");
  });
});

describe("escapeHTML (XSS prevention)", () => {
  it("escapes ampersand", () => {
    expect(escapeHTML("foo & bar")).toBe("foo &amp; bar");
  });

  it("escapes less than", () => {
    expect(escapeHTML("<script>")).toBe("&lt;script&gt;");
  });

  it("escapes greater than", () => {
    expect(escapeHTML("a > b")).toBe("a &gt; b");
  });

  it("escapes double quotes", () => {
    expect(escapeHTML('say "hello"')).toBe("say &quot;hello&quot;");
  });

  it("escapes single quotes", () => {
    expect(escapeHTML("it's")).toBe("it&#39;s");
  });

  it("escapes XSS payload", () => {
    const payload = '<script>alert("XSS")</script>';
    expect(escapeHTML(payload)).toBe("&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;");
  });

  it("handles empty string", () => {
    expect(escapeHTML("")).toBe("");
  });

  it("handles string with no special chars", () => {
    expect(escapeHTML("hello world")).toBe("hello world");
  });
});

describe("stripHTML (XSS prevention)", () => {
  it("strips HTML tags", () => {
    expect(stripHTML("<b>bold</b>")).toBe("bold");
  });

  it("strips script tags", () => {
    expect(stripHTML('<script>alert("XSS")</script>')).toBe('alert("XSS")');
  });

  it("handles nested tags", () => {
    expect(stripHTML("<div><span>text</span></div>")).toBe("text");
  });

  it("handles self-closing tags", () => {
    expect(stripHTML("line 1<br/>line 2")).toBe("line 1line 2");
  });

  it("handles empty string", () => {
    expect(stripHTML("")).toBe("");
  });

  it("handles string with no tags", () => {
    expect(stripHTML("hello world")).toBe("hello world");
  });
});
