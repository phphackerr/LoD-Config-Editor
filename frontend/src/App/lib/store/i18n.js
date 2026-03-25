import { get, writable } from 'svelte/store';
import { addMessages, init, locale } from 'svelte-i18n';
import {
  CreateLanguage,
  GetCurrentLanguage,
  GetLanguages,
  GetTranslations,
  SaveTranslation
} from '/bindings/lce/backend/i18n/i18n';
import { toLocalizedError, tr } from './storeUtils';

const FALLBACK_LOCALE = 'en';
const initialI18nState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null,
  currentLocale: FALLBACK_LOCALE,
  languages: []
};

const I18N_OP_INIT = 'i18n:init';
const I18N_OP_CHANGE_LANGUAGE = 'i18n:change-language';
const I18N_OP_GET_LANGUAGES = 'i18n:get-languages';
const I18N_OP_CREATE_LANGUAGE = 'i18n:create-language';
const I18N_OP_SAVE_TRANSLATIONS = 'i18n:save-translations';
const I18N_OP_GET_TRANSLATIONS = 'i18n:get-translations';

let operationSeq = 0;

export const i18nState = writable({ ...initialI18nState });

function patchI18nState(patch) {
  i18nState.update((state) => ({ ...state, ...patch }));
}

function createI18nResult(operation, { ok, data = null, error = null, rolledBack = false }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function startI18nOperation(operation) {
  operationSeq += 1;
  const token = operationSeq;
  patchI18nState({
    loading: true,
    success: false,
    error: null,
    operation
  });
  return token;
}

function finishI18nOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  patchI18nState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    ...patch
  });
  return result;
}

function normalizeLangCode(langCode, fallback = FALLBACK_LOCALE) {
  const code = String(langCode || '').trim();
  return code || fallback;
}

function normalizeLanguages(input) {
  return Array.isArray(input) ? input : [];
}

async function fetchTranslations(langCode) {
  const code = normalizeLangCode(langCode);
  const translations = await GetTranslations(code);
  return { code, translations: translations || {} };
}

async function applyLocale(langCode) {
  const loaded = await fetchTranslations(langCode);
  addMessages(loaded.code, loaded.translations);
  locale.set(loaded.code);
  return loaded;
}

export async function changeLanguage(langCode) {
  const code = normalizeLangCode(langCode);
  const token = startI18nOperation(I18N_OP_CHANGE_LANGUAGE);
  const previousLocale = normalizeLangCode(get(locale), FALLBACK_LOCALE);

  try {
    const loaded = await applyLocale(code);
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_CHANGE_LANGUAGE, {
        ok: true,
        data: loaded
      }),
      { currentLocale: loaded.code }
    );
  } catch (error) {
    locale.set(previousLocale);
    const message = toLocalizedError(
      error,
      'ERRORS.i18n.load_locale',
      tr('ERRORS.i18n.load_locale', { code }),
      { code }
    );

    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_CHANGE_LANGUAGE, {
        ok: false,
        error: message,
        data: { code: previousLocale },
        rolledBack: true
      }),
      { currentLocale: previousLocale }
    );
  }
}

export async function getAvailableLanguages() {
  const token = startI18nOperation(I18N_OP_GET_LANGUAGES);

  try {
    const langs = await GetLanguages();
    const normalized = normalizeLanguages(langs);

    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_GET_LANGUAGES, {
        ok: true,
        data: normalized
      }),
      { languages: normalized }
    );
  } catch (error) {
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_GET_LANGUAGES, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.i18n.load_languages'),
        data: []
      }),
      { languages: [] }
    );
  }
}

export async function createNewLanguage(code, name, author) {
  const langCode = normalizeLangCode(code, '');
  const langName = String(name || '').trim();
  const langAuthor = String(author || '').trim();
  const token = startI18nOperation(I18N_OP_CREATE_LANGUAGE);

  if (!langCode || !langName) {
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_CREATE_LANGUAGE, {
        ok: false,
        error: tr('ERRORS.i18n.language_code_and_name_required')
      })
    );
  }

  try {
    await CreateLanguage(langCode, langName, langAuthor);
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_CREATE_LANGUAGE, {
        ok: true,
        data: { code: langCode, name: langName, author: langAuthor }
      })
    );
  } catch (error) {
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_CREATE_LANGUAGE, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.i18n.create_language',
          tr('ERRORS.i18n.create_language', { code: langCode }),
          { code: langCode }
        )
      })
    );
  }
}

export async function getTranslationsForLanguage(langCode) {
  const code = normalizeLangCode(langCode);
  const token = startI18nOperation(I18N_OP_GET_TRANSLATIONS);

  try {
    const loaded = await fetchTranslations(code);
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_GET_TRANSLATIONS, {
        ok: true,
        data: loaded.translations
      })
    );
  } catch (error) {
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_GET_TRANSLATIONS, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.i18n.load_translations',
          tr('ERRORS.i18n.load_translations', { code }),
          { code }
        ),
        data: {}
      })
    );
  }
}

export async function saveTranslations(langCode, translations) {
  const code = normalizeLangCode(langCode);
  const token = startI18nOperation(I18N_OP_SAVE_TRANSLATIONS);

  try {
    await SaveTranslation(code, translations);
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_SAVE_TRANSLATIONS, {
        ok: true,
        data: { code }
      })
    );
  } catch (error) {
    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_SAVE_TRANSLATIONS, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.i18n.save_translations',
          tr('ERRORS.i18n.save_translations', { code }),
          { code }
        )
      })
    );
  }
}

export async function initGoI18n() {
  const token = startI18nOperation(I18N_OP_INIT);
  let currentLocale = FALLBACK_LOCALE;

  try {
    currentLocale = (await GetCurrentLanguage()) || FALLBACK_LOCALE;
  } catch {
    currentLocale = FALLBACK_LOCALE;
  }
  currentLocale = normalizeLangCode(currentLocale);

  try {
    init({
      fallbackLocale: FALLBACK_LOCALE,
      initialLocale: currentLocale,
      handleMissingMessage: ({ id }) => id
    });

    await applyLocale(FALLBACK_LOCALE);
    let resolvedLocale = FALLBACK_LOCALE;

    if (currentLocale !== FALLBACK_LOCALE) {
      await applyLocale(currentLocale);
      resolvedLocale = currentLocale;
    }

    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_INIT, {
        ok: true,
        data: { locale: resolvedLocale }
      }),
      { currentLocale: resolvedLocale }
    );
  } catch (error) {
    locale.set(FALLBACK_LOCALE);
    const message = toLocalizedError(error, 'ERRORS.i18n.init');

    return finishI18nOperation(
      token,
      createI18nResult(I18N_OP_INIT, {
        ok: false,
        error: message,
        data: { locale: FALLBACK_LOCALE },
        rolledBack: true
      }),
      { currentLocale: FALLBACK_LOCALE }
    );
  }
}
