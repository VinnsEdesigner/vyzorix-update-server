import { useTranslation } from 'react-i18next';
import { VYZOR_I18N_DEFAULT_NS } from '@/lib/i18n';

export function useVyzorTranslation() {
  const { t, i18n, ready } = useTranslation(VYZOR_I18N_DEFAULT_NS);
  return { t, i18n, ready };
}
