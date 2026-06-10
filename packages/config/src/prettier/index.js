// @vyzorix/config/prettier - Prettier Configuration
// Standardized Prettier configuration for Vyzorix projects

export default {
  // Print width
  printWidth: 100,
  
  // Tab width
  tabWidth: 2,
  
  // Use spaces instead of tabs
  useTabs: false,
  
  // Semicolons at the end of statements
  semi: true,
  
  // Use single quotes
  singleQuote: true,
  
  // Quote props only when needed
  quoteProps: "as-needed",
  
  // Use single quotes in JSX
  jsxSingleQuote: false,
  
  // Trailing commas
  trailingComma: "es5",
  
  // Spaces inside object literals
  bracketSpacing: true,
  
  // Put > on the same line in JSX
  bracketSameLine: false,
  
  // Format with prettier
  arrowParens: "always",
  
  // Range format
  rangeStart: 0,
  rangeEnd: Infinity,
  
  // Require pragma
  requirePragma: false,
  
  // Insert pragma
  insertPragma: false,
  
  // Prose wrap
  proseWrap: "preserve",
  
  // HTML whitespace sensitivity
  htmlWhitespaceSensitivity: "css",
  
  // Vue files script and style tags indentation
  vueIndentScriptAndStyle: false,
  
  // End of line
  endOfLine: "lf",
  
  // Embedded language formatting
  embeddedLanguageFormatting: "auto",
  
  // Single attribute per line
  singleAttributePerLine: false,
  
  // Overrides for specific file types
  overrides: [
    {
      files: "*.json",
      options: {
        printWidth: 120,
        tabWidth: 2,
      },
    },
    {
      files: "*.md",
      options: {
        proseWrap: "always",
        printWidth: 80,
      },
    },
    {
      files: "*.{yaml,yml}",
      options: {
        tabWidth: 2,
        singleQuote: false,
      },
    },
  ],
};