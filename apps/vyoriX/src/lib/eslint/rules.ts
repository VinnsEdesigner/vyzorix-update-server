/**
 * Vyzorix Layered Architecture ESLint Rules
 * UI Layer â Hooks â Domain â Data Layer (inward only)
 */
import type { TSESLint, TSESTree } from "@typescript-eslint/utils";

const LAYERS = {
  ui: { name: "UI Layer", path: "src/ui", canImportFrom: ["hooks", "ui", "react"], canNotImportFrom: ["domain", "lib"] },
  hooks: { name: "Presentation Layer (Hooks)", path: "src/hooks", canImportFrom: ["domain", "hooks", "lib", "react"], canNotImportFrom: ["ui"] },
  domain: { name: "Domain Layer", path: "src/domain", canImportFrom: ["domain"], canNotImportFrom: ["hooks", "lib", "ui", "react"] },
  data: { name: "Data Layer", path: "src/lib", canImportFrom: ["domain", "hooks", "lib", "react"], canNotImportFrom: ["ui"] },
} as const;

type LayerKey = keyof typeof LAYERS;

function getFileLayer(filePath: string): LayerKey | null {
  const normalizedPath = filePath.replace(/\/g, "/");
  if (normalizedPath.includes("/src/ui/")) return "ui";
  if (normalizedPath.includes("/src/hooks/")) return "hooks";
  if (normalizedPath.includes("/src/domain/")) return "domain";
  if (normalizedPath.includes("/src/lib/")) return "data";
  return null;
}

function getImportLayer(importPath: string): LayerKey | null {
  const normalizedPath = importPath.replace(/\/g, "/").replace(/^@\//, "");
  if (normalizedPath.startsWith("components/") || normalizedPath.startsWith("ui/")) return "ui";
  if (normalizedPath.startsWith("hooks/")) return "hooks";
  if (normalizedPath.startsWith("domain/")) return "domain";
  if (normalizedPath.startsWith("lib/")) return "data";
  return null;
}

function isSameLayer(filePath: string, importPath: string): boolean {
  return getFileLayer(filePath) === getImportLayer(importPath);
}

export const layerImportsRule: TSESLint.RuleModule<"forbidden" | "message"> = {
  defaultOptions: [],
  meta: { type: "problem", docs: { description: "Enforce layered architecture import rules", recommended: "error" }, schema: [], messages: { forbidden: "{{ fromLayer }} cannot import from {{ toLayer }}. {{ rule }}" } },
  create(context) {
    const filename = context.filename ?? context.getFilename();
    return {
      ImportDeclaration(node: TSESTree.ImportDeclaration) {
        const source = node.source.value;
        if (typeof source !== "string") return;
        if (!source.startsWith("@/")) return;
        if (source === "@/") return;
        const fromLayer = getFileLayer(filename);
        if (!fromLayer) return;
        const toLayer = getImportLayer(source);
        if (!toLayer) return;
        if (isSameLayer(filename, source)) return;
        const fromLayerInfo = LAYERS[fromLayer];
        const toLayerInfo = LAYERS[toLayer];
        const isForbidden = fromLayerInfo.canNotImportFrom.includes(toLayer);
        if (isForbidden) {
          let rule = "";
          if (fromLayer === "ui" && toLayer === "domain") rule = "Use hooks from @/hooks instead.";
          else if (fromLayer === "ui" && toLayer === "data") rule = "Use hooks from @/hooks instead.";
          else if (fromLayer === "hooks" && toLayer === "ui") rule = "Hooks must be pure logic without UI dependencies.";
          else if (fromLayer === "domain") rule = "Domain must be pure types/transforms only.";
          context.report({ node, messageId: "forbidden", data: { fromLayer: fromLayerInfo.name, fromPath: filename, toLayer: toLayerInfo.name, toPath: source, rule } });
        }
      },
    };
  },
};

export const noUiInDomainRule: TSESLint.RuleModule<"forbidden"> = {
  defaultOptions: [],
  meta: { type: "problem", docs: { description: "Domain cannot import React or UI", recommended: "error" }, schema: [], messages: { forbidden: "Domain cannot import {{ package }}. Domain must be pure TypeScript." } },
  create(context) {
    const filename = context.filename ?? context.getFilename();
    if (getFileLayer(filename) !== "domain") return {};
    return {
      ImportDeclaration(node: TSESTree.ImportDeclaration) {
        const source = node.source.value;
        if (typeof source !== "string") return;
        if (source.startsWith("@/domain/")) return;
        const forbiddenImports = [
          { pattern: /^react$/, name: "React" },
          { pattern: /^react-dom/, name: "react-dom" },
          { pattern: /^@\/components/, name: "@/components" },
          { pattern: /^@\/hooks/, name: "@/hooks" },
          { pattern: /^@\/lib\/(?!api\/)/, name: "@/lib (non-API)" },
        ];
        for (const { pattern, name } of forbiddenImports) {
          if (pattern.test(source)) {
            context.report({
              node,
              messageId: "forbidden",
              data: { package: name },
            });
          }
        }
      },
    };
  },
};
