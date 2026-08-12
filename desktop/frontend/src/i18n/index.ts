import i18next, { type TOptions } from 'i18next';
import { resources, type TranslationKey } from './resources.ts';

export const supportedLocales = ['en', 'es', 'pt-BR', 'it'] as const;
export type SupportedLocale = (typeof supportedLocales)[number];
export const languageStorageKey = 'corsarr.desktop.language';

export function normalizeLocale(locale: string | null | undefined): SupportedLocale | undefined {
  const normalized = locale?.trim().replace(/_/g, '-').toLowerCase();
  if (!normalized) return undefined;
  if (normalized === 'en' || normalized.startsWith('en-')) return 'en';
  if (normalized === 'es' || normalized.startsWith('es-')) return 'es';
  if (normalized === 'pt' || normalized.startsWith('pt-')) return 'pt-BR';
  if (normalized === 'it' || normalized.startsWith('it-')) return 'it';
  return undefined;
}

export function detectLocale(
  storedLocale?: string | null,
  browserLocales: readonly string[] = [],
): SupportedLocale {
  const stored = normalizeLocale(storedLocale);
  if (stored) return stored;
  for (const locale of browserLocales) {
    const detected = normalizeLocale(locale);
    if (detected) return detected;
  }
  return 'en';
}

export function initializeLocalization(locale: SupportedLocale): void {
  void i18next.init({
    lng: locale,
    fallbackLng: 'en',
    initAsync: false,
    interpolation: { escapeValue: false },
    resources: Object.fromEntries(
      Object.entries(resources).map(([code, translation]) => [code, { translation }]),
    ),
  });
  document.documentElement.lang = locale;
}

export function translate(key: TranslationKey, options?: TOptions): string {
  return i18next.t(key, options);
}

export function currentLocale(): SupportedLocale {
  return normalizeLocale(i18next.resolvedLanguage) ?? 'en';
}
