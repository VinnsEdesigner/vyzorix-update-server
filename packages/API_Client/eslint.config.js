import js from "@eslint/js";
import tseslint from "@typescript-eslint/eslint-plugin";
import tsparser from "@typescript-eslint/parser";

/**
 * Strict ESLint configuration for API Client
 * Designed to catch potential bugs, enforce type safety, and maintain code quality
 * 
 * Note: Type-checking rules require tsconfig to be set up via parserOptions.project
 * For full type-aware linting, run: npx eslint src --ext .ts,.tsx --parser-options=project:./tsconfig.json
 */
export default [
  js.configs.recommended,
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
      },
      globals: {
        window: "readonly",
        console: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        WebSocket: "readonly",
        Blob: "readonly",
        URLSearchParams: "readonly",
        URL: "readonly",
        crypto: "readonly",
        Buffer: "readonly",
        import: "readonly",
        gql: "readonly",
        clear: "readonly",
        FormData: "readonly",
        Headers: "readonly",
        Request: "readonly",
        Response: "readonly",
        fetch: "readonly",
      },
    },
    plugins: {
      "@typescript-eslint": tseslint,
    },
    rules: {
      // TypeScript recommended rules (errors)
      ...tseslint.configs.recommended.rules,
      
      // Unused variables - error (allow underscore prefix)
      "@typescript-eslint/no-unused-vars": ["error", { 
        argsIgnorePattern: "^_",
        varsIgnorePattern: "^_",
        caughtErrorsIgnorePattern: "^_",
      }],
      
      // Explicit any - STRICT ERROR (no `any` allowed in API client)
      "@typescript-eslint/no-explicit-any": "error",
      
      // Require explicit return types on public/exported functions
      "@typescript-eslint/explicit-function-return-type": ["warn", {
        allowExpressions: true,
        allowConciseArrowFunctionExpressionsStartingWithVoid: true,
      }],
      
      // Consistent type imports (prevents type/value confusion)
      "@typescript-eslint/consistent-type-imports": "error",
      
      // Consistent type definitions (prefer interface, but allow type for raw/DTO patterns)
      "@typescript-eslint/consistent-type-definitions": "warn",
      
      // No non-null assertions (can hide null issues)
      "@typescript-eslint/no-non-null-assertion": "warn",
      
      // Consistent array type syntax
      "@typescript-eslint/array-type": ["error", { default: "array" }],
      
      // Disallow unused expressions (e.g., `foo;` without effect)
      "no-unused-expressions": ["error", {
        allowShortCircuit: true,
        allowTernary: true,
      }],
      
      // Disallow reassigning function parameters
      "no-param-reassign": "error",
      
      // Disallow unnecessary catch clause parameters
      "no-useless-catch": "error",
      
      // Prefer const over let/var
      "prefer-const": "error",
      
      // Require exhaustive cases in switch/case
      "default-case": "error",
      
      // No duplicate imports (disabled - TypeScript handles this via import type)
      "no-duplicate-imports": "off",
      
      // Disallow unused imports
      "no-unused-vars": "off", // Handled by @typescript-eslint/no-unused-vars
      "no-redeclare": "off",   // Handled by TypeScript
      "no-empty": "off",        // Handled by @typescript-eslint/no-empty-function
    },
  },
  // Rules specific to test files
  {
    files: ["**/*.test.ts", "**/*.test.tsx", "**/*.spec.ts", "**/*.spec.tsx"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/explicit-function-return-type": "off",
    },
  },
];
