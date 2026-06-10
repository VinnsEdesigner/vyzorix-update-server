// @vyzorix/config/nitro/index.ts - Nitro Configuration Helpers
import { cloudflarePreset } from "./targets/cloudflare";
import { nodePreset } from "./targets/node";
import { staticPreset } from "./targets/static";

export const nitroTargets = {
  cloudflare: cloudflarePreset,
  node: nodePreset,
  static: staticPreset,
};

export type NitroTarget = keyof typeof nitroTargets;

export function getNitroTarget(name: NitroTarget) {
  return nitroTargets[name];
}

export { cloudflarePreset, nodePreset, staticPreset };
export default nitroTargets;