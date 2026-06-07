import js from "@eslint/js";
import eslintPluginPrettier from "eslint-plugin-prettier/recommended";
import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import tseslint from "typescript-eslint";
import unusedImports from "eslint-plugin-unused-imports";
import importPlugin from "eslint-plugin-import";
import { fileURLToPath } from "url";
import { dirname, resolve } from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = resolve(__filename, "..");

export default tseslint.config(
  { ignores: ["dist", ".output", ".vinxi", "node_modules"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: projectRoot,
        sourceType: "module",
      },
    },
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      "unused-imports": unusedImports,
      import: importPlugin,
    },
    settings: {
      "import/resolver": {
        typescript: {
          alwaysTryTypes: true,
        },
      },
    },
    rules: {
      // =============================================
      // STRICT MODE - Catch everything possible
      // =============================================

      // React Hooks (critical)
      ...reactHooks.configs.recommended.rules,
      "react-hooks/exhaustive-deps": "error",

      // =============================================
      // UNUSED CODE (these catch bugs early)
      // =============================================
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
          destructuredArrayIgnorePattern: "^_",
        },
      ],
      "unused-imports/no-unused-imports": "error",
      "unused-imports/no-unused-vars": [
        "error",
        { vars: "all", varsIgnorePattern: "^_", args: "after-used", argsIgnorePattern: "^_" },
      ],

      // =============================================
      // BEST PRACTICES (prevents bugs)
      // =============================================
      "no-implicit-coercion": "error",
      "no-console": ["warn", { allow: ["warn", "error", "info"] }],
      "no-alert": "error",
      "no-debugger": "error",
      "no-eval": "error",
      "no-new-wrappers": "error",
      "no-return-await": "error",
      "no-throw-literal": "error",
      "no-unused-expressions": ["error", { allowShortCircuit: false, allowTernary: false }],
      "no-useless-return": "error",
      "prefer-promise-reject-errors": ["error", { allowEmptyReject: false }],
      "require-await": "error",

      // =============================================
      // STYLISTIC (enforce consistency)
      // =============================================
      "array-bracket-newline": ["error", "consistent"],
      "comma-dangle": [
        "error",
        { arrays: "always-multiline", objects: "always-multiline", functions: "only-multiline" },
      ],
      "comma-spacing": "error",
      "func-style": ["warn", "expression", { allowArrowFunctions: true }],
      "implicit-arrow-linebreak": "error",
      "max-len": ["error", { code: 120, ignoreStrings: true, ignoreTemplateLiterals: true }],
      "no-bitwise": "error",
      "no-mixed-operators": "error",
      "no-multiple-empty-lines": ["error", { max: 1, maxBOF: 0, maxEOF: 0 }],
      "no-nested-ternary": "error",
      "no-trailing-spaces": "error",
      "object-curly-spacing": ["error", "always"],
      "one-var": ["error", "never"],
      "padded-blocks": ["error", "never"],
      "prefer-object-spread": "error",
      semi: ["error", "always"],
      "semi-style": "error",
      "space-before-blocks": "error",

      // =============================================
      // IMPORT RULES (dependency tracking)
      // =============================================
      "import/no-unresolved": "error",
      "import/named": "error",
      "import/default": "error",
      "import/export": "error",
      "import/namespace": "error",
      "import/newline-after-import": "error",
      "import/order": [
        "error",
        {
          groups: [["external", "builtin"], "internal", ["parent", "sibling", "index"]],
          pathGroups: [{ pattern: "@/**", group: "internal" }],
          pathGroupsExcludedImportTypes: ["internal"],
          "newlines-between": "always",
          alphabetize: { order: "asc", caseInsensitive: true },
        },
      ],
      "import/no-duplicates": "error",
      "import/no-self-import": "error",
      "import/no-cycle": "error",
      "import/no-useless-path-segments": "error",

      // =============================================
      // TANSTACK START SPECIFIC
      // =============================================
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "server-only",
              message: "Use *.server.ts pattern instead of server-only package",
            },
          ],
        },
      ],
      "react-refresh/only-export-components": "warn",

      // =============================================
      // TYPESCRIPT STRICT RULES
      // =============================================
      "@typescript-eslint/ban-ts-comment": [
        "error",
        {
          "ts-expect-error": "allow-with-description",
          "ts-ignore": false,
          "ts-nocheck": false,
          "ts-check": false,
        },
      ],
      "@typescript-eslint/no-confusing-void-expression": [
        "error",
        { ignoreArrowShorthand: true, ignoreVoidOperator: true },
      ],
      "@typescript-eslint/no-duplicate-enum-values": "error",
      "@typescript-eslint/no-extraneous-class": "error",
      "@typescript-eslint/no-invalid-void-type": "error",
      "@typescript-eslint/no-misused-spread": "error",
      "@typescript-eslint/no-non-null-asserted-nullish-coalescing": "error",
      "@typescript-eslint/no-non-null-asserted-optional-chain": "error",
      "@typescript-eslint/no-unnecessary-qualifier": "error",
      "@typescript-eslint/no-useless-empty-export": "error",
      "@typescript-eslint/prefer-enum-initializers": "error",
      "@typescript-eslint/prefer-for-of": "error",
      "@typescript-eslint/prefer-function-type": "error",
      "@typescript-eslint/prefer-includes": "error",
      "@typescript-eslint/prefer-literal-enum-member": "error",
      "@typescript-eslint/prefer-nullish-coalescing": [
        "error",
        { ignoreBooleanCoercion: true, ignoreConditionalTests: true },
      ],
      "@typescript-eslint/prefer-optional-chain": "error",
      "@typescript-eslint/prefer-readonly": "error",
      "@typescript-eslint/prefer-reduce-type-parameter": "error",
      "@typescript-eslint/prefer-string-starts-ends-with": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "@typescript-eslint/triple-slash-reference": "error",
      "@typescript-eslint/unified-signatures": "error",
      "@typescript-eslint/consistent-type-exports": "error",
      "@typescript-eslint/consistent-generic-constructors": "error",
      "@typescript-eslint/consistent-indexed-object-style": "error",
      "@typescript-eslint/consistent-type-definitions": ["error", "interface"],
      "@typescript-eslint/explicit-function-return-type": [
        "warn",
        { allowExpressions: true, allowConciseArrowFunctionExpressionsStartingWithVoid: true },
      ],
      "@typescript-eslint/explicit-module-boundary-types": "off",
      "@typescript-eslint/no-base-to-string": "error",
    },
  },
  eslintPluginPrettier,
);
