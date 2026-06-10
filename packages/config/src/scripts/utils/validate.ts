// @vyzorix/config/scripts/utils/validate.ts - Validation utilities
import { access } from "fs/promises";
import { join } from "path";

const CONFIG_FILES = {
  vite: "vite.config.ts",
  eslint: "eslint.config.js",
  prettier: ".prettierrc",
  vitest: "vitest.config.ts",
  tailwind: "tailwind.config.js",
  "github-actions": ".github/workflows/ci.yml",
  vscode: ".vscode/settings.json",
  docker: "docker-compose.yml",
};

export async function validateSetup(): Promise<void> {
  const errors: string[] = [];
  const warnings: string[] = [];

  // Check for generated config files
  for (const [service, filename] of Object.entries(CONFIG_FILES)) {
    try {
      await access(join(process.cwd(), filename));
    } catch {
      warnings.push(`${service}: ${filename} not found`);
    }
  }

  // Check for package.json
  try {
    await access(join(process.cwd(), "package.json"));
  } catch {
    errors.push("package.json not found");
  }

  // Check for source directory
  try {
    await access(join(process.cwd(), "src"));
  } catch {
    warnings.push("src/ directory not found (may not be initialized yet)");
  }

  // Report results
  if (errors.length > 0) {
    console.error("\n❌ Validation failed:");
    errors.forEach((error) => console.error(`  - ${error}`));
    throw new Error("Setup validation failed");
  }

  if (warnings.length > 0) {
    console.warn("\n⚠️  Validation warnings:");
    warnings.forEach((warning) => console.warn(`  - ${warning}`));
  }

  if (errors.length === 0 && warnings.length === 0) {
    console.log("  ✅ All configuration files validated successfully");
  }
}