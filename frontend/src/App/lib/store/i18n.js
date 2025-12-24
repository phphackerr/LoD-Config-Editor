//@ts-nocheck

import { addMessages, init, locale } from 'svelte-i18n';
import { GetTranslations, GetCurrentLanguage, GetLanguages } from '/bindings/lce/backend/i18n/i18n';

async function loadGoTranslations(lang) {
  try {
    const translations = await GetTranslations(lang);
    addMessages(lang, translations);
    console.log(translations);
    locale.set(lang);
  } catch (e) {
    console.error(`Failed to load locale ${lang}:`, e);
  }
}

export async function changeLanguage(lang) {
  await loadGoTranslations(lang);
}

export async function getAvailableLanguages() {
  try {
    const langs = await GetLanguages();
    return langs;
  } catch (e) {
    console.error('Failed to get available languages:', e);
    return [];
  }
}

export async function initGoI18n() {
  console.log('init go i18n called');
  let lang = 'en';
  try {
    lang = await GetCurrentLanguage();
    console.log('lang: ' + lang);
  } catch (e) {
    console.warn('Не удалось получить язык из настроек, используем en');
  }
  init({
    fallbackLocale: 'en',
    initialLocale: lang,
    handleMissingMessage: ({ id }) => id
  });

  // Всегда грузим английский как фоллбэк
  await loadGoTranslations('en');

  if (lang !== 'en') {
    await loadGoTranslations(lang);
  }
}
