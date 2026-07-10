import { dirname, resolve } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";
import vyzo from "../../packages/config/src/eslint/architecture-index";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("expo", "prettier"),
  {
    plugins: {
      vyzo,
    },
    rules: {
      ...vyzo.rules,
      "prettier/prettier": "warn",
    },
  },
];

export default eslintConfig;
