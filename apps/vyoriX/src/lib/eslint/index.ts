/**
 * Vyzorix ESLint Plugin
 *
 * Custom ESLint rules for the Vyzorix layered architecture.
 */

import { layerImportsRule, noUiInDomainRule } from "./rules";

export const vyzoRules = {
  "vyzo/layer-imports": layerImportsRule,
  "vyzo/no-ui-in-domain": noUiInDomainRule,
};

export const vyzo = {
  rules: vyzoRules,
};

export { layerImportsRule, noUiInDomainRule };
