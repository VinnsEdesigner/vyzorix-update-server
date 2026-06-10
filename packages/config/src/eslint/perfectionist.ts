// @vyzorix/config/eslint/perfectionist.ts - Import sorting and perfectionist rules

export const perfectionistRules = {
  // Import organization
  "import/order": [
    "error",
    {
      groups: [
        "builtin",
        "external",
        "internal",
        "parent",
        "sibling",
        "index",
        "object",
        "type",
      ],
      "newlines-between": "always",
      alphabetize: {
        order: "asc",
        caseInsensitive: true,
      },
     warnOnUnassignedImports: true,
    },
  ],
  "import/no-unresolved": "off", // TypeScript handles this
  "import/named": "error",
  "import/default": "error",
  "import/namespace": "off",
  "import/no-named-as-default": "warn",
  "import/no-named-as-default-member": "warn",
  "import/no-deprecated": "warn",
  "import/no-extraneous-dependencies": [
    "error",
    {
      devDependencies: [
        "**/*.test.ts",
        "**/*.test.tsx",
        "**/*.spec.ts",
        "**/*.spec.tsx",
        "**/test/**",
        "**/tests/**",
        "**/mock/**",
        "**/__mocks__/**",
        "**/fixtures/**",
        "**/examples/**",
        "vite.config.ts",
        "vitest.config.ts",
        "eslint.config.js",
        "tailwind.config.js",
        "tsconfig.json",
      ],
    },
  ],
  "import/no-restricted-paths": "off",
  "import/no-cycle": "warn",
  "import/no-useless-path-segments": "warn",
  "import/dynamic-import-chunkname": "off",
  "import/no-relative-parent-imports": "off",
  "import/no-unused-modules": "warn",
};

export default perfectionistRules;