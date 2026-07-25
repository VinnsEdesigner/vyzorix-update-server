/**
 * Vyzorix Layered Architecture ESLint Rules
 * 
 * Architecture Dependency Flow (inward only):
 * ┌─────────────────────────────────────────────────────┐
 * │                    UI Layer                         │
 * │         (components, pages, layouts)                │
 * │                    ↓ only                          │
 * │                Hooks Layer                          │
 * │         (queries, mutations, state)                 │
 * │                    ↓ only                          │
 * │   ┌─────────────┐       ┌────────────────┐       │
 * │   │ API_Client   │       │     Lib        │       │
 * │   │ (domain,api) │       │  (utilities)   │       │
 * │   └─────────────┘       └────────────────┘       │
 * └─────────────────────────────────────────────────────┘
 * 
 * Enforced Rules:
 * - UI can ONLY import from Hooks
 * - Hooks can import from API_Client (domain, api), Hooks, Lib
 * - API_Client is PURE TypeScript (no React, no UI)
 * - Build FAILS if rules are violated
 */
import type { TSESLint, TSESTree } from "@typescript-eslint/utils";

type LayerKey = "ui" | "hooks" | "domain" | "vyzorServer";

const LAYERS: Record<LayerKey, { 
  name: string; 
  path: string; 
  canImportFrom: LayerKey[]; 
  canNotImportFrom: LayerKey[];
}> = {
  ui: { 
    name: "UI Layer", 
    path: "src/ui", 
    canImportFrom: ["hooks", "ui", "react"] as LayerKey[],
    canNotImportFrom: ["domain", "vyzorServer"] as LayerKey[]
  },
  hooks: { 
    name: "Hooks Layer", 
    path: "src/hooks", 
    canImportFrom: ["domain", "vyzorServer", "hooks", "react"] as LayerKey[],
    canNotImportFrom: ["ui"] as LayerKey[]
  },
  domain: { 
    name: "Domain Layer (API_Client)", 
    path: "packages/API_Client/src/domain", 
    canImportFrom: ["domain"] as LayerKey[],
    canNotImportFrom: ["hooks", "ui", "vyzorServer", "react"] as LayerKey[]
  },
  vyzorServer: { 
    name: "VyzoServer Layer (API_Client)", 
    path: "packages/API_Client/src/vyzorServer", 
    canImportFrom: ["domain", "vyzorServer", "react"] as LayerKey[],
    canNotImportFrom: ["hooks", "ui"] as LayerKey[]
  },
};

function getFileLayer(filePath: string): LayerKey | null {
  const normalizedPath = filePath.replace(/\\/g, "/");
  
  if (normalizedPath.includes("/src/ui/")) return "ui";
  if (normalizedPath.includes("/src/hooks/")) return "hooks";
  if (normalizedPath.includes("/packages/API_Client/src/domain/")) return "domain";
  if (normalizedPath.includes("/packages/API_Client/src/vyzorServer/")) return "vyzorServer";
  
  return null;
}

function getImportLayer(importPath: string): LayerKey | null {
  const normalizedPath = importPath.replace(/\\/g, "/").replace(/^@\//, "");
  
  if (normalizedPath.startsWith("components/") || normalizedPath.startsWith("ui/") || normalizedPath.startsWith("pages/")) return "ui";
  if (normalizedPath.startsWith("hooks/")) return "hooks";
  if (normalizedPath.startsWith("@vyzorix/api-client/domain/") || normalizedPath.startsWith("domain/")) return "domain";
  if (normalizedPath.startsWith("@vyzorix/api-client/vyzorServer/") || normalizedPath.startsWith("vyzorServer/")) return "vyzorServer";
  
  return null;
}

function isSameLayer(filePath: string, importPath: string): boolean {
  return getFileLayer(filePath) === getImportLayer(importPath);
}

export const layerImportsRule: TSESLint.RuleModule<"forbidden", []> = {
  defaultOptions: [],
  meta: {
    type: "problem",
    docs: { description: "Enforce layered architecture - dependencies must flow inward only" },
    schema: [],
    messages: { forbidden: "{{ fromLayer }} cannot import from {{ toLayer }}. {{ rule }}" },
  },
  create(context: TSESLint.RuleContext<"forbidden", []>) {
    const filename = (context as unknown as { filename?: string }).filename ?? (context.getFilename as () => string)();

    return {
      ImportDeclaration(node: TSESTree.ImportDeclaration) {
        const source = node.source.value;
        if (typeof source !== "string") return;
        if (!source.startsWith("@/") && !source.startsWith("@vyzorix/")) return;

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
          if (fromLayer === "ui" && (toLayer === "domain" || toLayer === "vyzorServer")) {
            rule = "UI must use hooks from @/hooks instead.";
          } else if (fromLayer === "hooks" && toLayer === "ui") {
            rule = "Hooks must be pure logic without UI dependencies.";
          } else if (fromLayer === "domain" || fromLayer === "vyzorServer") {
            rule = "API_Client must be pure TypeScript - no React or UI imports.";
          }

          context.report({ node, messageId: "forbidden", data: { fromLayer: fromLayerInfo.name, toLayer: toLayerInfo.name, rule } });
        }
      },
    };
  },
};

export const noReactInApiClientRule: TSESLint.RuleModule<"forbidden", []> = {
  defaultOptions: [],
  meta: {
    type: "problem",
    docs: { description: "API_Client package cannot import React or UI-related code" },
    schema: [],
    messages: { forbidden: "API_Client cannot import {{ package }}. Must be pure TypeScript." },
  },
  create(context: TSESLint.RuleContext<"forbidden", []>) {
    const filename = (context as unknown as { filename?: string }).filename ?? (context.getFilename as () => string)();
    const fileLayer = getFileLayer(filename);
    
    if (fileLayer !== "domain" && fileLayer !== "vyzorServer") return {};

    return {
      ImportDeclaration(node: TSESTree.ImportDeclaration) {
        const source = node.source.value;
        if (typeof source !== "string") return;

        const forbiddenImports = [
          { pattern: /^react$/, name: "React" },
          { pattern: /^react-dom/, name: "react-dom" },
          { pattern: /^react-native$/, name: "react-native" },
          { pattern: /^@\/components/, name: "@/components" },
          { pattern: /^@\/hooks/, name: "@/hooks" },
          { pattern: /^@\/ui/, name: "@/ui" },
          { pattern: /^@vyzorix\/ui/, name: "@vyzorix/ui" },
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
