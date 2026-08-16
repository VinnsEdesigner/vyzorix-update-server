import { devtools, type DevtoolsOptions } from 'zustand/middleware';

export function isDevtoolsEnabled(): boolean {
  return !import.meta.env.PROD;
}

export function buildDevtoolsOptions(name: string): DevtoolsOptions {
  return {
    name,
    enabled: isDevtoolsEnabled(),
  };
}

export { devtools };
