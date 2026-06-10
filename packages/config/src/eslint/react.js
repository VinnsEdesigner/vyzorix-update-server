// @vyzorix/config/eslint/react.js - React-specific ESLint rules

export const reactRules = {
  // React rules
  "react/jsx-uses-react": "error",
  "react/jsx-uses-vars": "error",
  "react/react-in-jsx-scope": "off", // React 17+ JSX transform
  "react/prop-types": "off", // Use TypeScript instead
  "react/require-default-props": "off", // Use TypeScript instead
  "react/display-name": "off", // Not needed with new JSX transform
  "react/jsx-fragments": "error",
  "react/jsx-no-target-blank": "error",
  "react/no-multi-comp": "off", // Allow multiple components in same file
  "react/no-unstable-nested-components": "error",
  "react/self-closing-comp": "warn",
  "react/prefer-read-only-props": "warn",
  "react/hook-use-state": "warn",
  
  // React Hooks rules
  "react-hooks/rules-of-hooks": "error",
  "react-hooks/exhaustive-deps": "warn",
};

export default reactRules;