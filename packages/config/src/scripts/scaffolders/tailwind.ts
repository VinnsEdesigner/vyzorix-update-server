// @vyzorix/config/scripts/scaffolders/tailwind.ts - Tailwind scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldTailwind(_target: string): Promise<void> {
  const tailwindConfig = `import vyzorixConfig from "@vyzorix/config/tailwind";

export default {
  ...vyzorixConfig,
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./pages/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
  ],
};
`;

  await writeFile(join(process.cwd(), "tailwind.config.js"), tailwindConfig);
}