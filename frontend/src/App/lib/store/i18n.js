//@ts-nocheck

import { addMessages, init, locale } from "svelte-i18n";
import {
  GetTranslations,
  GetCurrentLanguage,
} from "/bindings/lce/backend/i18n/i18n";

async function loadGoTranslations(lang) {
  try {
    const translations = await GetTranslations(lang);
    addMessages(lang, translations);
    locale.set(lang);
  } catch (e) {
    console.error(`Failed to load locale ${lang}:`, e);
  }
}

export async function initGoI18n() {
  console.log("init go i18n called");
  let lang = "en";
  try {
    lang = await GetCurrentLanguage();
    console.log("lang: " + lang);
  } catch (e) {
    console.warn("Не удалось получить язык из настроек, используем en");
  }
  init({
    fallbackLocale: "en",
    initialLocale: lang,
  });
  await loadGoTranslations(lang);
}
