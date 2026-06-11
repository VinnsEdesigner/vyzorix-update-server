# @vyzorix/config

> Comprehensive configuration package for Vyzorix ecosystem projects

**@vyzorix/config** provides a unified configuration system for Vyzorix projects,.

## Features

###  Core Configuration
- **Vite Configuration** (`defineViteConfig`) - Complete Vite setup with TanStack Start, React, Tailwind CSS, and SSR support
- **API Client** (`createApiClient`) - Type-safe API client with authentication, error handling, and automatic token management
- **Auth Client** (`createAuthClient`) - OAuth and JWT authentication flow management
- **Environment Management** (`loadEnv`) - Zod-validated environment variables with type safety

###  Code Quality
- **ESLint Configuration** - Comprehensive ESLint setup for TypeScript, React, and imports
- **Prettier Configuration** - Standardized code formatting rules
- **Vitest Configuration** - Test setup with coverage and utilities

###  Styling
- **Tailwind CSS Preset** - Vyzorix brand colors, design system, and theme configuration
- **Dark Mode Support** - Built-in dark mode theming

###  Development Tools
- **Git Hooks** - Husky integration with lint-staged for pre-commit validation
- **GitHub Actions Templates** - CI/CD workflows for testing, linting, and deployment
- **VSCode Settings** - Recommended extensions and editor configuration
- **Docker Compose** - Local development environment setup

###  Deployment
- **Multi-target Support** - Cloudflare Workers, Node.js, Static hosting, Docker
- **CI/CD Templates** - GitHub Actions workflows for automated deployments

## Installation

```bash
# Using pnpm (recommended)
pnpm add -D @vyzorix/config

# Using npm
npm install -D @vyzorix/config

# Using yarn
yarn add -D @vyzorix/config
```

## Quick Start

### Interactive Setup

```bash
# Interactive CLI - will ask you questions
npx @vyzorix/config init

# Use a preset
npx @vyzorix/config init --preset ssr

# Skip prompts, use defaults
npx @vyzorix/config init --yes
```

### Manual Setup

#### 1. Vite Configuration

```typescript
// vite.config.ts
import { defineViteConfig } from "@vyzorix/config/vite";

export default defineViteConfig({
  tanstackStart: {
    server: { entry: "src/server.ts" },
  },
  proxy: {
    "/v1": "http://localhost:3000",
    "/api": "http://localhost:3000",
  },
});
```

#### 2. API Client

```typescript
// lib/api.ts
import { createApiClient } from "@vyzorix/config/api";

export const api = createApiClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL,
  auth: {
    tokenStorage: "localStorage",
    tokenKey: "vyzorix_token",
  },
});

// Usage
const { data: devices } = await api.get("/v1/dashboard/devices");
await api.post("/v1/device/register", { deviceId: "123" });
```

#### 3. Auth Client

```typescript
// lib/auth.ts
import { createAuthClient } from "@vyzorix/config/auth";

export const auth = createAuthClient({
  backendUrl: "http://localhost:3000",
  callbackUrl: "http://localhost:5173/auth/callback",
  providers: [{ name: "google" }],
  storage: "localStorage",
});

// Sign in with Google
await auth.signInWithOAuth("google");

// Handle OAuth callback
auth.handleCallback(new URLSearchParams(window.location.search));

// Get current session
const session = auth.getSession();

// Sign out
await auth.signOut();
```

#### 4. Environment Management

```typescript
// lib/env.ts
import { loadEnv } from "@vyzorix/config/env";

export const env = loadEnv();

// Access typed environment variables
console.log(env.VITE_API_BASE_URL);
console.log(env.isDev());  // Check if in development
console.log(env.isProd()); // Check if in production

// Feature flags
if (env.isFeatureEnabled("analytics")) {
  // Enable analytics
}

// Get API URL
const apiUrl = env.getApiUrl("/v1/devices");
```

#### 5. Tailwind CSS

```javascript
// tailwind.config.js
import vyzorixConfig from "@vyzorix/config/tailwind";

export default {
  ...vyzorixConfig,
  // Your custom config
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
};
```

## Presets

### Available Presets

- **ssr** - Full-stack SSR app with TanStack Start
- **spa** - Single-page application with client-side rendering
- **lib** - React component library
- **go-api** - Go backend API service
- **minimal** - Just the essentials (Vite + ESLint + Prettier)

### Using Presets

```bash
# SSR with all services
npx @vyzorix/config init --preset ssr

# SPA preset
npx @vyzorix/config init --preset spa

# Library preset
npx @vyzorix/config init --preset lib

# Go API preset
npx @vyzorix/config init --preset go-api

# Minimal preset
npx @vyzorix/config init --preset minimal
```

## Configuration Files

When you run `npx @vyzorix/config init`, it will create the following files:

```
├── vite.config.ts              # Vite configuration
├── eslint.config.js            # ESLint configuration
├── .prettierrc                 # Prettier configuration
├── vitest.config.ts            # Vitest configuration
├── vitest.setup.ts             # Test setup and utilities
├── tailwind.config.js          # Tailwind CSS configuration
├── .husky/                     # Git hooks (if selected)
│   └── pre-commit
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI/CD
├── .vscode/
│   ├── settings.json           # VSCode settings
│   └── extensions.json         # Recommended extensions
└── docker-compose.yml          # Docker Compose (if selected)
```

## Environment Variables

### Required Variables

```env
# API Configuration
VITE_API_BASE_URL=http://localhost:3000
VITE_AUTH_REDIRECT_URI=http://localhost:5173/auth/callback

# Authentication
VITE_GOOGLE_CLIENT_ID=your-google-client-id

# UI Configuration
VITE_DEFAULT_THEME=system
VITE_PRIMARY_COLOR=oklch(0.645 0.246 16.439)

# Feature Flags
VITE_ENABLE_ANALYTICS=false
VITE_ENABLE_ERROR_REPORTING=false
VITE_ENABLE_EXPERIMENTAL_FEATURES=false
```

## Brand Colors

The Vyzorix design system uses OKLCH colors for better perceptional uniformity:

### Primary (Brand Orange)
```
Light: oklch(0.645 0.246 16.439)
Dark:  oklch(0.645 0.246 16.439)
```

### Secondary (Cool Blue-Gray)
```
Light: oklch(0.968 0.007 247.896)
Dark:  oklch(0.279 0.041 260.031)
```

### Accent (Teal)
```
Light: oklch(0.704 0.191 22.216)
Dark:  oklch(0.704 0.191 22.216)
```

## API Reference

### `defineViteConfig(options?)`

Creates a Vite configuration with all Vyzorix defaults.

**Options:**
```typescript
interface VyzorixViteConfig {
  tanstackStart?: {
    server?: { entry?: string };
  };
  proxy?: Record<string, string | { target: string; changeOrigin?: boolean }>;
  vite?: ViteUserConfig;
}
```

### `createApiClient(config)`

Creates a typed API client.

**Options:**
```typescript
interface ApiClientConfig {
  baseUrl: string;
  auth?: {
    tokenStorage?: "localStorage" | "sessionStorage" | "memory";
    tokenKey?: string;
    refreshEndpoint?: string;
  };
  timeout?: number;
}
```

### `createAuthClient(config)`

Creates an authentication client.

**Options:**
```typescript
interface AuthClientConfig {
  backendUrl: string;
  callbackUrl: string;
  providers?: AuthProviderConfig[];
  storage?: "localStorage" | "sessionStorage" | "memory";
  tokenKey?: string;
}
```

### `loadEnv(overrides?)`

Loads and validates environment variables.

**Returns:**
```typescript
interface EnvVariables {
  NODE_ENV: "development" | "production" | "test";
  VITE_API_BASE_URL: string;
  VITE_GOOGLE_CLIENT_ID?: string;
  isDev: () => boolean;
  isProd: () => boolean;
  getApiUrl: (path?: string) => string;
  isFeatureEnabled: (feature: "analytics" | "errorReporting" | "experimental") => boolean;
}
```

## Scripts

The init script supports the following commands:

```bash
# Interactive initialization
npx @vyzorix/config init

# With preset
npx @vyzorix/config init --preset ssr

# Non-interactive
npx @vyzorix/config init --yes

# Custom services
npx @vyzorix/config init --include vite,eslint,prettier,vitest

# Specific deployment target
npx @vyzorix/config init --target cloudflare
```


## License

MIT License - see LICENSE file for details.
