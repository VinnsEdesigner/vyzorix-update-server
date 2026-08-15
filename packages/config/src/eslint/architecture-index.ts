/**
 * Vyzorix Architecture Rules Plugin
 * 
 * Shared architectural dependency enforcement rules
 * used by both VyzoriX_web and VyzoriX_mobile.
 * 
 * Usage in eslint.config.js:
 *   import { vyzoRules } from '@vyzorix/config/eslint/architecture-index';
 */
import { layerImportsRule, noReactInApiClientRule } from "./architecture-rules.js";

export const vyzoRules = {
  "layer-imports": layerImportsRule,
  "no-react-in-api-client": noReactInApiClientRule,
};

export const vyzo = {
  rules: vyzoRules,
};

export { layerImportsRule, noReactInApiClientRule };
