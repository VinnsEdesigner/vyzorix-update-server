import { dirname, resolve } from "path";
import { fileURLToPath } from "url";
import js from "@eslint/js";
import tseslint from "typescript-eslint";
import { vyzo } from "../../packages/config/dist/eslint/architecture-index.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = resolve(__filename, "..");

export default tseslint.config(
  { ignores: ["dist", "node_modules", ".tsbuildinfo", "mocks/**"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["src/**/*.ts", "src/**/*.tsx"],
    languageOptions: {
      ecmaVersion: 2020,
      sourceType: "module",
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: projectRoot,
      },
    },
    plugins: {
      vyzo,
    },
    rules: {
      "vyzo/layer-imports": "error",
      "vyzo/no-react-in-api-client": "error",
      "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_", caughtErrorsIgnorePattern: "^_" }],
      "@typescript-eslint/explicit-function-return-type": "off",
      "@typescript-eslint/no-explicit-any": "warn",
      "no-unused-vars": "off",
    },
  },
);
