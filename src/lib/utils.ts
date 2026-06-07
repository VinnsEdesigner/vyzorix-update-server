import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const cn = (...inputs: ClassValue[]): string => {
  return twMerge(clsx(inputs));
};

// XSS Prevention: HTML escape utility
// Use this when rendering user-controlled content that might contain HTML
const HTML_ESCAPE_MAP: Record<string, string> = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#39;",
};

/**
 * Escapes HTML special characters to prevent XSS attacks.
 * Use this for any user-controlled content that needs to be displayed safely.
 */
export const escapeHTML = (str: string): string => {
  return str.replace(/[&<>"']/g, (char) => HTML_ESCAPE_MAP[char] ?? char);
};

/**
 * Validates that a string doesn't contain HTML tags.
 * Returns the string if clean, otherwise sanitized version.
 */
export const stripHTML = (str: string): string => {
  return str.replace(/<[^>]*>/g, "");
};
