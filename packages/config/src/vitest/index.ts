// @vyzorix/config/vitest - Vitest Configuration
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

/**
 * Vitest Configuration for Vyzorix projects
 */
export default defineConfig({
  plugins: [
    react(),
    tsconfigPaths(),
  ],
  
  test: {
    // Test environment
    environment: "jsdom",
    
    // Glob patterns for test files
    include: [
      "**/*.{test,spec}.{ts,tsx}",
      "**/*.test.{ts,tsx}",
      "**/*.spec.{ts,tsx}",
    ],
    
    // Exclude patterns
    exclude: [
      "**/node_modules/**",
      "**/dist/**",
      "**/build/**",
      "**/.next/**",
      "**/coverage/**",
    ],
    
    // Coverage configuration
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html", "lcov"],
      exclude: [
        "**/node_modules/**",
        "**/dist/**",
        "**/build/**",
        "**/.next/**",
        "**/coverage/**",
        "**/*.test.{ts,tsx}",
        "**/*.spec.{ts,tsx}",
        "**/*.d.ts",
        "**/types/**",
        "**/*.config.ts",
        "**/vitest.config.ts",
      ],
      thresholds: {
        statements: 80,
        branches: 80,
        functions: 80,
        lines: 80,
      },
    },
    
    // Setup files
    setupFiles: ["./vitest.setup.ts"],
    
    // Global test timeout
    testTimeout: 10000,
    
    // Hook timeout
    hookTimeout: 10000,
    
    //pool: "forks",
    //poolOptions: {
    //  forks: {
    //    singleFork: true,
    //  },
    //},
    
    // Environment match
    env: {
      NODE_ENV: "test",
    },
    
    // globals: true,
    // Uncomment above if you want to use Jest-style globals (describe, it, expect, etc.)
    
    // reporters
    reporters: ["default", "verbose"],
    
    // Type checking
    typecheck: {
      checker: "tsc",
      include: ["**/*.test.{ts,tsx}", "**/*.spec.{ts,tsx}"],
    },
  },
  
  // Resolve aliases
  resolve: {
    alias: {
      "@": "/src",
      "@vyzorix/config": "/packages/config/src",
    },
  },
});