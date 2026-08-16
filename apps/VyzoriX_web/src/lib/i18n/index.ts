export {
  createVyzorI18n,
  getDefaultVyzorI18n,
  changeVyzorLanguage,
  VYZOR_I18N_DEFAULT_NS,
  VYZOR_I18N_FALLBACK_LNG,
  VYZOR_AVAILABLE_LANGUAGES,
  type VyzorLanguageValue,
} from './vyzor-i18n-config';
export { loadLocale, loadLocaleInto } from './vyzor-i18n-loader';
export type {
  VyzorNamespace,
  VyzorCommonKey,
  VyzorUpdatesKey,
  VyzorTranslationKey,
  VyzorLocaleResource,
  VyzorLocaleResources,
} from './vyzor-i18n-types';
