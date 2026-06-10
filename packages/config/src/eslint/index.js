// @vyzorix/config/eslint/index.js - Main ESLint Configuration
import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";
import { tsRules } from "./ts.js";
import { reactRules } from "./react.js";
import { perfectionistRules } from "./perfectionist.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  // Base configuration
  ...compat.extends(
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    "plugin:import/errors",
    "plugin:import/warnings",
    "prettier"
  ),
  
  // React configuration
  {
    files: ["**/*.{jsx,tsx}"],
    plugins: {
      react: await import("eslint-plugin-react"),
      "react-hooks": await import("eslint-plugin-react-hooks"),
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      ...reactRules,
    },
  },
  
  // TypeScript configuration
  {
    files: ["**/*.{ts,tsx}"],
    plugins: {
      "@typescript-eslint": await import("@typescript-eslint/eslint-plugin"),
    },
    languageOptions: {
      parser: await import("@typescript-eslint/parser"),
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    rules: {
      ...tsRules,
    },
  },
  
  // Import organization
  {
    plugins: {
      import: await import("eslint-plugin-import"),
    },
    rules: {
      ...perfectionistRules,
    },
  },
  
  // Test files
  {
    files: ["**/*.test.{ts,tsx}", "**/*.spec.{ts,tsx}"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unused-vars": "off",
      "no-unused-vars": "off",
    },
  },
  
  // Node files
  {
    files: ["**/node.{js,cjs,mjs}"],
    rules: {
      "no-unused-vars": "off",
      "no-console": "off",
    },
  },
  
  // Ignore patterns
  {
    ignores: [
      "**/node_modules/**",
      "**/dist/**",
      "**/build/**",
      "**/.next/**",
      "**/coverage/**",
      "**/*.min.js",
      "**/CHANGELOG.md",
      "**/LICENSE.md",
      ".git/**",
      ".husky/**",
      ".vscode/**",
      "**/*.d.ts",
    ],
  },
];

export default eslintConfig;