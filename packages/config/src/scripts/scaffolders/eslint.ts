// @vyzorix/config/scripts/scaffolders/eslint.ts - ESLint scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldESLint(_target: string): Promise<void> {
  const eslintConfig = `// @ts-check
import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends(
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    "plugin:import/errors",
    "plugin:import/warnings",
    "prettier"
  ),
  {
    files: ["**/*.{ts,tsx}"],
    plugins: {
      "@typescript-eslint": await import("@typescript-eslint/eslint-plugin"),
    },
    languageOptions: {
      parser: await import("@typescript-eslint/parser"),
    },
    rules: {
      "@typescript-eslint/no-unused-vars": ["warn", {
        argsIgnorePattern: "^_",
        varsIgnorePattern: "^_"
      }],
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },
  {
    files: ["**/*.{js,jsx}"],
    rules: {
      "no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
    },
  },
  {
    ignores: ["**/node_modules/**", "**/dist/**", "**/build/**", "**/.next/**"],
  },
];

export default eslintConfig;
`;

  await writeFile(join(process.cwd(), "eslint.config.js"), eslintConfig);
}