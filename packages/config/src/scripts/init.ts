#!/usr/bin/env node
// @vyzorix/config/scripts/init.ts - Interactive CLI for scaffolding Vyzorix config
// Usage: npx @vyzorix/config init [--preset <name>] [--include <services>] [--target <platform>] [--yes]

import { cac } from "cac";
import inquirer from "inquirer";
import { scaffoldVite } from "./scaffolders/vite.js";
import { scaffoldESLint } from "./scaffolders/eslint.js";
import { scaffoldPrettier } from "./scaffolders/prettier.js";
import { scaffoldVitest } from "./scaffolders/vitest.js";
import { scaffoldTailwind } from "./scaffolders/tailwind.js";
import { scaffoldGitHooks } from "./scaffolders/git-hooks.js";
import { scaffoldDocker } from "./scaffolders/docker.js";
import { scaffoldGitHubActions } from "./scaffolders/github-actions.js";
import { scaffoldVSCode } from "./scaffolders/vscode.js";
import { updatePackageJson } from "./utils/package-json.js";
import { validateSetup } from "./utils/validate.js";

const cli = cac("vyzorix-config");

// Preset configurations
const presets = {
  ssr: {
    name: "SSR (TanStack Start)",
    description: "Full-stack app with server-side rendering",
    services: ["vite", "eslint", "prettier", "vitest", "tailwind", "git-hooks", "github-actions", "vscode"],
    target: "node",
  },
  spa: {
    name: "SPA (React + Vite)",
    description: "Single-page application with client-side rendering",
    services: ["vite", "eslint", "prettier", "vitest", "tailwind", "git-hooks", "github-actions", "vscode"],
    target: "static",
  },
  lib: {
    name: "Library (React Component)",
    description: "Shareable React component library",
    services: ["eslint", "prettier", "vitest", "git-hooks", "github-actions", "vscode"],
    target: "node",
  },
  "go-api": {
    name: "Go API",
    description: "Go backend API service",
    services: ["github-actions", "vscode"],
    target: "node",
  },
  minimal: {
    name: "Minimal",
    description: "Just the essentials - Vite + ESLint + Prettier",
    services: ["vite", "eslint", "prettier"],
    target: "node",
  },
};

async function promptPresets(presetName?: string) {
  if (presetName && presets[presetName as keyof typeof presets]) {
    const preset = presets[presetName as keyof typeof presets];
    return {
      preset: presetName,
      projectType: presetName,
      framework: presetName === "go-api" ? "go" : "react-tanstack",
      services: preset.services,
      target: preset.target,
      packageManager: "pnpm",
    };
  }

  const { preset } = await inquirer.prompt([
    {
      type: "list",
      name: "preset",
      message: "Which preset would you like to use?",
      choices: [
        ...Object.entries(presets).map(([key, value]) => ({
          name: `${value.name} - ${value.description}`,
          value: key,
        })),
        {
          name: "Custom (select services individually)",
          value: "custom",
        },
      ],
    },
  ]);

  if (preset === "custom") {
    return promptCustom();
  }

  return promptPresets(preset);
}

async function promptCustom() {
  const { projectType, framework, selectedServices, target, packageManager } = await inquirer.prompt([
    {
      type: "list",
      name: "projectType",
      message: "What type of project is this?",
      choices: [
        { name: "SSR Application", value: "ssr" },
        { name: "Single Page Application (SPA)", value: "spa" },
        { name: "React Component Library", value: "lib" },
        { name: "Go API Service", value: "go-api" },
      ],
    },
    {
      type: "list",
      name: "framework",
      message: "What framework are you using?",
      choices: [
        { name: "React + TanStack", value: "react-tanstack" },
        { name: "React + Next.js", value: "react-next" },
        { name: "React (Vite only)", value: "react-vite" },
        { name: "Go", value: "go" },
      ],
    },
    {
      type: "checkbox",
      name: "selectedServices",
      message: "Which services would you like to include?",
      choices: [
        { name: "Vite Configuration", value: "vite", checked: true },
        { name: "ESLint (Code Linting)", value: "eslint", checked: true },
        { name: "Prettier (Code Formatting)", value: "prettier", checked: true },
        { name: "Vitest (Testing)", value: "vitest", checked: true },
        { name: "Tailwind CSS (Styling)", value: "tailwind", checked: true },
        { name: "Git Hooks (Husky)", value: "git-hooks", checked: false },
        { name: "GitHub Actions (CI/CD)", value: "github-actions", checked: true },
        { name: "VSCode Settings", value: "vscode", checked: false },
        { name: "Docker Compose", value: "docker", checked: false },
      ],
    },
    {
      type: "list",
      name: "target",
      message: "What is your deployment target?",
      choices: [
        { name: "Node.js Server", value: "node" },
        { name: "Cloudflare Workers", value: "cloudflare" },
        { name: "Static Hosting (Vercel, Netlify)", value: "static" },
        { name: "Docker Container", value: "docker" },
      ],
    },
    {
      type: "list",
      name: "packageManager",
      message: "Which package manager do you use?",
      choices: [
        { name: "pnpm (recommended)", value: "pnpm" },
        { name: "npm", value: "npm" },
        { name: "yarn", value: "yarn" },
      ],
    },
  ]);

  return {
    preset: "custom",
    projectType,
    framework,
    services: selectedServices,
    target,
    packageManager,
  };
}

async function scaffoldAll(answers: any) {
  console.log("\n🚀 Scaffolding Vyzorix config...\n");

  const { services, target } = answers;

  const scaffolders = {
    vite: scaffoldVite,
    eslint: scaffoldESLint,
    prettier: scaffoldPrettier,
    vitest: scaffoldVitest,
    tailwind: scaffoldTailwind,
    "git-hooks": scaffoldGitHooks,
    docker: scaffoldDocker,
    "github-actions": scaffoldGitHubActions,
    vscode: scaffoldVSCode,
  };

  for (const service of services) {
    if (scaffolders[service as keyof typeof scaffolders]) {
      console.log(`  📦 Scaffolding ${service}...`);
      try {
        await scaffolders[service as keyof typeof scaffolders](target);
        console.log(`  ✅ ${service} scaffolded successfully`);
      } catch (error) {
        console.error(`  ❌ Failed to scaffold ${service}:`, error);
      }
    }
  }

  console.log("\n  📝 Updating package.json...");
  await updatePackageJson(answers);

  console.log("\n  🔍 Validating setup...");
  await validateSetup();
}

async function main() {
  cli
    .command("init", "Initialize Vyzorix config in your project")
    .option("--preset <name>", "Use a preset (ssr, spa, lib, go-api, minimal)")
    .option("--include <services>", "Comma-separated services to include")
    .option("--target <platform>", "Deployment target (node, cloudflare, static, docker)")
    .option("--yes", "Skip prompts, use defaults")
    .action(async (options) => {
      try {
        let answers;

        if (options.preset) {
          answers = await promptPresets(options.preset);
        } else if (options.yes) {
          answers = await promptPresets("ssr");
        } else {
          answers = await promptPresets();
        }

        if (options.include) {
          answers.services = options.include.split(",").map((s: string) => s.trim());
        }

        if (options.target) {
          answers.target = options.target;
        }

        await scaffoldAll(answers);

        console.log("\n" + "=".repeat(60));
        console.log("✅ Vyzorix config initialized successfully!");
        console.log("=".repeat(60));
        console.log("\n📋 Next steps:");
        console.log("  1. Run `pnpm install` to install new dependencies");
        console.log("  2. Review the generated configuration files");
        console.log("  3. Run `pnpm dev` to start development");
        console.log("\n📚 Documentation:");
        console.log("  - @vyzorix/config README: ./node_modules/@vyzorix/config/README.md");
        console.log("  - Vite config: vite.config.ts");
        console.log("  - ESLint config: eslint.config.js");
        console.log("  - Prettier config: .prettierrc");
        console.log("=".repeat(60) + "\n");
      } catch (error) {
        console.error("❌ Initialization failed:", error);
        process.exit(1);
      }
    });

  cli.help();
  cli.parse(process.argv);
}

main();