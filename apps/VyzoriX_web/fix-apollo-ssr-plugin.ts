import type { Plugin } from "vite";

const apolloNoExternal = ["@apollo/client", "@apollo/client/*", "graphql", "graphql-tag", "zen-observable-ts"];

export function fixApolloSsrPlugin(): Plugin {
  return {
    name: "fix:apollo-ssr-noexternal",
    enforce: "post",
    configResolved(config) {
      const ssr = (config.environments as any)?.ssr;
      if (!ssr) return;
      ssr.resolve = ssr.resolve || {};
      const existing = Array.isArray(ssr.resolve.noExternal) ? ssr.resolve.noExternal : [];
      ssr.resolve.noExternal = [...existing, ...apolloNoExternal];
    },
  };
}
