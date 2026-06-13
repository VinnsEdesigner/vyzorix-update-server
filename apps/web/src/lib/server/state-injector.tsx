/**
 * state-injector.tsx - SSR state injection utilities
 *
 * This module handles injecting server-provided auth state into HTML
 * for React hydration.
 *
 * Based on Library's server.ts implementation pattern
 */

import type { Operator } from "./cookie-reader";

export interface HydratedState {
  isAuthenticated: boolean;
  operator: Operator | null;
}

/**
 * Generate script tag for state injection
 *
 * This injects state into window.__VYZORIX_PREFETCHED_STATE__
 * matching Library's server.ts pattern:
 *
 *   const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;
 *   return html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
 */
const generateStateScript = (state: HydratedState): string => {
  // Use JSON.stringify to properly escape the state for embedding in HTML
  const stateJson = JSON.stringify(state);

  return `<script id="__vyzorix-prefetched-state__" type="application/json">
window.__VYZORIX_PREFETCHED_STATE__ = ${stateJson};
</script>`;
};

export { generateStateScript };

/**
 * Inject state into HTML before the app mount point
 *
 * Library's pattern (index.html):
 *   <div id="root"><!--app-html--></div>
 *   <!--app-state-->
 *
 * This replaces <!--app-state--> with the actual state script
 */
const injectStateIntoHtml = (html: string, state: HydratedState): string => {
  const stateScript = generateStateScript(state);

  // Replace the placeholder with the actual state script
  return html.replace("<!--app-state-->", stateScript);
};

export { injectStateIntoHtml };

/**
 * Inject state into HTML right after the root div (Library's exact pattern)
 *
 * This matches Library's exact injection point:
 *   return html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
 */
const injectStateAfterRoot = (html: string, state: HydratedState): string => {
  const stateScript = generateStateScript(state);

  // Replace after <div id="root"> or <div id="app">
  // Both patterns are supported for flexibility
  let result = html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
  result = result.replace('<div id="app">', `${stateScript}\n<div id="app">`);

  return result;
};

export { injectStateAfterRoot };

/**
 * Get hydrated state from window object
 *
 * This is the client-side counterpart for reading server-injected state.
 * Matches Library's getHydratedState pattern from src/lib/config.ts:
 *
 *   export function getHydratedState<T>(key: string, defaultValue: T): T {
 *     if (typeof window !== 'undefined') {
 *       const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
 *       if (globalState && globalState[key] !== undefined) {
 *         return globalState[key];
 *       }
 *     }
 *     return defaultValue;
 *   }
 */
function getHydratedState<T>(key: string, defaultValue: T): T {
  if (typeof window !== "undefined") {
    const globalState = (window as unknown as { __VYZORIX_PREFETCHED_STATE__?: HydratedState })
      .__VYZORIX_PREFETCHED_STATE__;
    if (globalState && key in globalState) {
      return (globalState as unknown as Record<string, T>)[key] ?? defaultValue;
    }
  }
  return defaultValue;
}

export { getHydratedState };

/**
 * Get the full hydrated state object
 */
const getFullHydratedState = (): HydratedState | null => {
  if (typeof window !== "undefined") {
    return (
      (window as unknown as { __VYZORIX_PREFETCHED_STATE__?: HydratedState })
        .__VYZORIX_PREFETCHED_STATE__ ?? null
    );
  }
  return null;
};

export { getFullHydratedState };
