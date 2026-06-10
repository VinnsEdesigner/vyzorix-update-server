// @vyzorix/config/scripts/scaffolders/vitest.ts - Vitest scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldVitest(_target: string): Promise<void> {
  const vitestConfig = `import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  test: {
    environment: "jsdom",
    include: ["**/*.{test,spec}.{ts,tsx}"],
    exclude: ["**/node_modules/**", "**/dist/**", "**/build/**"],
    setupFiles: ["./vitest.setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: ["**/node_modules/**", "**/dist/**", "**/*.test.{ts,tsx}"],
    },
  },
});
`;

  const vitestSetup = `import "@testing-library/jest-dom";
import { beforeEach, afterEach, vi } from "vitest";

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});
`;

  await writeFile(join(process.cwd(), "vitest.config.ts"), vitestConfig);
  await writeFile(join(process.cwd(), "vitest.setup.ts"), vitestSetup);
}