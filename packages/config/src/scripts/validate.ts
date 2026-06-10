// @vyzorix/config/scripts/validate.ts - Config validator for pre-commit
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

export async function validateConfig(): Promise<{ valid: boolean; errors: string[]; warnings: string[] }> {
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
  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export async function validateSetup(): Promise<void> {
  const { errors, warnings } = await validateConfig();
  
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

// CLI entry point
if (import.meta.url === `file://${process.argv[1]}`) {
  validateSetup()
    .then(() => process.exit(0))
    .catch(() => process.exit(1));
}

export default validateConfig;