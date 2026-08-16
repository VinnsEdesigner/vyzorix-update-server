import { vyzorCommonEn } from './locales/en/vyzor-common-en';
import { vyzorUpdatesEn } from './locales/en/vyzor-updates-en';

export const VYZOR_I18N_NAMESPACES = ['common', 'updates'] as const;
export type VyzorNamespace = (typeof VYZOR_I18N_NAMESPACES)[number];

export type VyzorCommonKey = keyof typeof vyzorCommonEn;
export type VyzorUpdatesKey = keyof typeof vyzorUpdatesEn;

export type VyzorTranslationKey = VyzorCommonKey | VyzorUpdatesKey;

export type VyzorLocaleResource = Record<string, string>;

export interface VyzorLocaleResources {
  common: VyzorLocaleResource;
  updates: VyzorLocaleResource;
}
