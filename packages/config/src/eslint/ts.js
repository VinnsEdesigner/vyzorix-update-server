// @vyzorix/config/eslint/ts.js - TypeScript-specific ESLint rules

export const tsRules = {
  // TypeScript rules
  "@typescript-eslint/no-explicit-any": "warn",
  "@typescript-eslint/no-non-null-assertion": "off",
  "@typescript-eslint/explicit-function-return-type": "off",
  "@typescript-eslint/explicit-module-boundary-types": "off",
  "@typescript-eslint/no-unused-vars": [
    "warn",
    {
      argsIgnorePattern: "^_",
      varsIgnorePattern: "^_",
      caughtErrorsIgnorePattern: "^_",
    },
  ],
  "@typescript-eslint/no-implied-eval": "error",
  "@typescript-eslint/dot-notation": "error",
  "@typescript-eslint/no-floating-promises": "error",
  "@typescript-eslint/no-misused-promises": "error",
  "@typescript-eslint/no-unnecessary-type-assertion": "warn",
  "@typescript-eslint/prefer-optional-chain": "warn",
  "@typescript-eslint/prefer-nullish-coalescing": "warn",
  "@typescript-eslint/require-await": "warn",
  "@typescript-eslint/await-thenable": "error",
  "@typescript-eslint/no-for-in-array": "error",
  "@typescript-eslint/no-inferrable-types": "off",
  "@typescript-eslint/no-namespace": "off",
  "@typescript-eslint/no-parameter-properties": "off",
  "@typescript-eslint/triple-slash-reference": "off",
};

export default tsRules;