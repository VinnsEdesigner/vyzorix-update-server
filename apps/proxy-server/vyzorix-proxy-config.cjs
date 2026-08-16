/**
 * vyzorix-proxy-config.cjs
 *
 * Configuration for the Vyzorix BFF proxy server.
 * All values are environment-driven with sensible defaults.
 */

'use strict';

const path = require('node:path');

function envStr(name, fallback) {
  const v = process.env[name];
  return v && v.length > 0 ? v : fallback;
}

function envInt(name, fallback) {
  const raw = process.env[name];
  if (!raw) return fallback;
  const n = parseInt(raw, 10);
  return Number.isNaN(n) ? fallback : n;
}

function envBool(name, fallback) {
  const raw = process.env[name];
  if (raw === undefined) return fallback;
  return raw === 'true' || raw === '1' || raw === 'yes';
}

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development';

// Resolve web app directory relative to this file.
const WEB_APP_DIR = envStr(
  'WEB_APP_DIR',
  path.resolve(__dirname, '..', 'VyzoriX_web'),
);

const config = {
  port: envInt('PROXY_PORT', 3001),
  apiTarget: envStr('API_TARGET', 'http://localhost:3000'),
  webAppDir: WEB_APP_DIR,
  webStaticDir: path.join(WEB_APP_DIR, 'dist', 'client'),
  mode: mode,
  verbose: envBool('PROXY_VERBOSE', true),
  signRequests: envBool('PROXY_SIGN', true),
  stripSigningKey: envBool('PROXY_STRIP_SIGNING_KEY', true),
  sessionCookieName: '_vyz_proxy_sid',
  sessionTtlMs: envInt('PROXY_SESSION_TTL_MS', 24 * 60 * 60 * 1000),
};

// Paths that should be proxied to the Go API (not served as static assets).
const API_PATH_PREFIXES = ['/v1/', '/api/', '/health', '/healthz', '/bin/'];

function isGraphQLPath(pathname) {
  // Org-scoped GraphQL: /:org/graphql, /:org/graphql/batch, /:org/graphql/ws
  return /^\/[^/]+\/graphql(\/(batch|ws))?$/.test(pathname);
}

function isApiPath(pathname) {
  if (API_PATH_PREFIXES.some(function (p) { return pathname.startsWith(p); })) return true;
  return isGraphQLPath(pathname);
}

// Auth endpoints that must NOT be signed (the Go API allows these unsigned).
const AUTH_ENDPOINT_PATTERNS = [
  '/v1/auth/login',
  '/v1/auth/refresh',
  '/v1/auth/csrf-token',
  '/v1/auth/register',
  '/v1/auth/mfa',
];

// Auth endpoints whose responses need interception (signing key, tokens, org ID).
// This is a superset of AUTH_ENDPOINT_PATTERNS — includes org select which IS
// signed but also needs response interception to capture organizationId.
const AUTH_INTERCEPT_PATTERNS = [
  '/v1/auth/login',
  '/v1/auth/refresh',
  '/v1/auth/csrf-token',
  '/v1/auth/register',
  '/v1/auth/mfa',
  '/v1/auth/organizations/select',
  '/v1/auth/logout',
];

function isAuthEndpoint(pathname) {
  return AUTH_ENDPOINT_PATTERNS.some(function (p) { return pathname.includes(p); });
}

function isAuthInterceptEndpoint(pathname) {
  return AUTH_INTERCEPT_PATTERNS.some(function (p) { return pathname.includes(p); });
}

function isStateChanging(method) {
  return ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method.toUpperCase());
}

module.exports = {
  config,
  isApiPath,
  isGraphQLPath,
  isAuthEndpoint,
  isAuthInterceptEndpoint,
  isStateChanging,
  API_PATH_PREFIXES,
  AUTH_ENDPOINT_PATTERNS,
  AUTH_INTERCEPT_PATTERNS,
};
