/**
 * Vyzorix Consolidated API Client
 *
 * High-level browser coordinator. This routing file delegates call signatures
 * directly to our 3 specialized micro-clients.
 *
 * - Credentials Client   : Standard login, registry form submissions, and password reset.
 * - SSO Client          : Direct Google & GitHub redirection routing.
 * - Verification Client : Continuous polling lookups and cancellation cycles.
 */

export * from "./clients/authClient";
export * from "./clients/ssoClient";
export * from "./clients/verificationClient";
