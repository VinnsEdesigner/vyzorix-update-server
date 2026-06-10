// @vyzorix/config/scripts/utils/package-json.ts - Package.json utilities
import { readFile } from "fs/promises";
import { join } from "path";

export async function updatePackageJson(answers: any): Promise<void> {
  const packageJsonPath = join(process.cwd(), "package.json");
  
  try {
    const packageJsonContent = await readFile(packageJsonPath, "utf-8");
    const packageJson = JSON.parse(packageJsonContent);
    
    // Update scripts
    packageJson.scripts = {
      ...packageJson.scripts,
      dev: "vite dev",
      build: "vite build",
      lint: "eslint .",
      typecheck: "tsc --noEmit",
      test: "vitest run",
      prepare: "husky install",
    };
    
    // Add lint-staged if git hooks selected
    if (answers.services.includes("git-hooks")) {
      packageJson["lint-staged"] = {
        "*.{ts,tsx,js,jsx}": ["eslint --fix", "prettier --write"],
        "*.{json,md,css}": ["prettier --write"],
      };
    }
    
    // Add new dependencies to install later
    const newDeps: string[] = [];
    const newDevDeps: string[] = [];
    
    if (answers.services.includes("vite")) {
      newDevDeps.push("@vyzorix/config", "vite", "typescript", "@vitejs/plugin-react");
    }
    if (answers.services.includes("eslint")) {
      newDevDeps.push("eslint", "@typescript-eslint/eslint-plugin", "@typescript-eslint/parser");
    }
    if (answers.services.includes("prettier")) {
      newDevDeps.push("prettier", "eslint-config-prettier");
    }
    if (answers.services.includes("vitest")) {
      newDevDeps.push("vitest", "@testing-library/jest-dom", "@testing-library/react");
    }
    if (answers.services.includes("tailwind")) {
      newDevDeps.push("tailwindcss", "@tailwindcss/vite");
    }
    if (answers.services.includes("git-hooks")) {
      newDevDeps.push("husky", "lint-staged");
    }
    
    // Write updated package.json
    await Bun.write(packageJsonPath, JSON.stringify(packageJson, null, 2));
    
    console.log("  📝 Dependencies to install:");
    if (newDeps.length > 0) {
      console.log("    Dependencies:", newDeps.join(", "));
    }
    if (newDevDeps.length > 0) {
      console.log("    Dev Dependencies:", newDevDeps.join(", "));
    }
  } catch (error) {
    console.error("  ⚠️  Could not update package.json:", error);
  }
}