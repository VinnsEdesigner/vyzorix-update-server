/**
 * Vyzorix Layered Architecture ESLint Rules
 * UI Layer → Hooks → Domain → Data Layer (inward only)
 */
import type { TSESLint, TSESTree } from "@typescript-eslint/utils";

const LAYERS = {
  ui: { name: "UI Layer", path: "src/ui", canImportFrom: ["hooks", "ui", "react"], canNotImportFrom: ["domain", "vyzorServer"] },
  hooks: { name: "Presentation Layer (Hooks)", path: "src/hooks", canImportFrom: ["domain", "hooks", "vyzorServer", "react"], canNotImportFrom: ["ui"] },
  domain: { name: "Domain Layer", path: "src/domain", canImportFrom: ["domain"], canNotImportFrom: ["hooks", "vyzorServer", "ui", "react"] },
  data: { name: "Data Layer", path: "src/vyzorServer", canImportFrom: ["domain", "hooks", "vyzorServer", "react"], canNotImportFrom: ["ui"] },
} as const;

type LayerKey = keyof typeof LAYERS;

function getFileLayer(filePath: string): LayerKey | null {
  const normalizedPath = filePath.replace(/\\/g, "/");
  if (normalizedPath.includes("/src/ui/")) return "ui";
  if (normalizedPath.includes("/src/hooks/")) return "hooks";
  if (normalizedPath.includes("/src/domain/")) return "domain";
  if (normalizedPath.includes("/src/vyzorServer/")) return "data";
  return null;
}

function getImportLayer(importPath: string): LayerKey | null {
  const normalizedPath = importPath.replace(/\\/g, "/").replace(/^@\//, "");
  if (normalizedPath.startsWith("components/") || normalizedPath.startsWith("ui/")) return "ui";
  if (normalizedPath.startsWith("hooks/")) return "hooks";
  if (normalizedPath.startsWith("domain/")) return "domain";
  if (normalizedPath.startsWith("vyzorServer/")) return "data";
  return null;
}

function isSameLayer(filePath: string, importPath: string): boolean {
  return getFileLayer(filePath) === getImportLayer(importPath);
}

export const layerImportsRule: TSESLint.RuleModule<"forbidden" | "message"> = {
  defaultOptions: [],
  meta: {
    type: "problem",
    docs: { description: "Enforce layered architecture - dependencies must flow inward only", recommended: "error" },
    schema: [],
    messages: { forbidden: "{{ fromLayer }} cannot import from {{ toLayer }}. {{ rule }}" },
  },
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
          if (fromLayer === "ui" && toLayer === "domain") {
            rule = "UI must use hooks from @/hooks instead.";
          } else if (fromLayer === "ui" && toLayer === "data") {
            rule = "UI must use hooks from @/hooks instead.";
          } else if (fromLayer === "hooks" && toLayer === "ui") {
            rule = "Hooks must be pure logic without UI dependencies.";
          } else if (fromLayer === "domain") {
            rule = "Domain must be pure types/transforms only.";
          }

          context.report({ node, messageId: "forbidden", data: { fromLayer: fromLayerInfo.name, fromPath: filename, toLayer: toLayerInfo.name, toPath: source, rule } });
        }
      },
    };
  },
};

export const noUiInDomainRule: TSESLint.RuleModule<"forbidden"> = {
  defaultOptions: [],
  meta: {
    type: "problem",
    docs: { description: "Domain layer cannot import React or UI-related code", recommended: "error" },
    schema: [],
    messages: { forbidden: "Domain layer cannot import {{ package }}. Domain must be pure TypeScript." },
  },
  create(context) {
    const filename = context.filename ?? context.getFilename();
    const fileLayer = getFileLayer(filename);
    if (fileLayer !== "domain") return {};

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
          { pattern: /^@\/vyzorServer\/(?!api\/)/, name: "@/vyzorServer (non-API)" },
        ];

        for (const { pattern, name } of forbiddenImports) {
          if (pattern.test(source)) {
            context.report({ node, messageId: "forbidden", data: { package: name } });
          }
        }
      },
    };
  },
};
