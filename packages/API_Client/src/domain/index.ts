// Re-export shared utilities
export * from "./_shared";

// Bounded contexts. Each context owns its types and mappers with names unique
// per context (resolved at source, not via barrel aliases), so every module is
// re-exported with `export *` without collisions.
export * from "./admin";
export * from "./apikey";
export * from "./auth";
export * from "./clientcredentials";
export * from "./email";
export * from "./oauth";
export * from "./organization";
export * from "./session";
export * from "./diagnostics";
export * from "./events";
export * from "./realtime";
export * from "./commands";
export * from "./device";
export * from "./invitation";
export * from "./logs";
export * from "./metrics";
export * from "./registration";
export * from "./settings";
export * from "./telemetry";
export * from "./updates";
