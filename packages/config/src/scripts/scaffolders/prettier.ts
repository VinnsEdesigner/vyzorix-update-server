// @vyzorix/config/scripts/scaffolders/prettier.ts - Prettier scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldPrettier(_target: string): Promise<void> {
  const prettierConfig = `{
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "semi": true,
  "singleQuote": true,
  "quoteProps": "as-needed",
  "jsxSingleQuote": false,
  "trailingComma": "es5",
  "bracketSpacing": true,
  "bracketSameLine": false,
  "arrowParens": "always",
  "endOfLine": "lf",
  "proseWrap": "preserve",
  "htmlWhitespaceSensitivity": "css"
}
`;

  await writeFile(join(process.cwd(), ".prettierrc"), prettierConfig);
}