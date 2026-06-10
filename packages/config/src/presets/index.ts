// @vyzorix/config/presets/index.ts - Preset Registry & Exports
import { ssrPreset } from "./ssr";
import { spaPreset } from "./spa";
import { libPreset } from "./lib";

export const presets = {
  ssr: ssrPreset,
  spa: spaPreset,
  lib: libPreset,
};

export type PresetName = keyof typeof presets;

export function getPreset(name: PresetName) {
  return presets[name];
}

export { ssrPreset, spaPreset, libPreset };
export default presets;