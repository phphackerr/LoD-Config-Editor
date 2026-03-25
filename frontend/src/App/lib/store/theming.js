import {
  LoadThemeCSS,
  GetThemes,
  LoadRawTheme,
  CreateTheme,
  SaveTheme
} from '/bindings/lce/backend/theming/themeservice';
import { get, writable } from 'svelte/store';
import { toLocalizedError, tr } from './storeUtils';

export const currentTheme = writable('default');

const initialThemeState = {
  loading: false,
  success: false,
  error: null,
  warnings: [],
  operation: '',
  lastResult: null
};

export const themeState = writable({
  ...initialThemeState
});

let styleEl = null;
let operationSeq = 0;

const THEME_OP_APPLY = 'theme:apply';
const THEME_OP_LIST = 'theme:list';
const THEME_OP_LOAD_RAW = 'theme:load-raw';
const THEME_OP_CREATE = 'theme:create';
const THEME_OP_SAVE = 'theme:save';

function patchThemeState(patch) {
  themeState.update((state) => ({ ...state, ...patch }));
}

function normalizeThemeName(name, fallback = 'default') {
  const normalized = String(name || '').trim();
  return normalized || fallback;
}

function normalizeThemeList(input) {
  return Array.isArray(input) ? input.map((item) => String(item || '').trim()).filter(Boolean) : [];
}

function createThemeResult(
  operation,
  { ok, data = null, error = null, warnings, rolledBack = false }
) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    warnings: Array.isArray(warnings) ? warnings : undefined,
    rolledBack: Boolean(rolledBack)
  };
}

function startThemeOperation(operation, { preserveWarnings = true } = {}) {
  operationSeq += 1;
  const token = operationSeq;
  patchThemeState({
    loading: true,
    success: false,
    error: null,
    operation,
    ...(preserveWarnings ? {} : { warnings: [] })
  });
  return token;
}

function finishThemeOperation(token, result) {
  if (token !== operationSeq) {
    return result;
  }

  patchThemeState(
    (() => {
      const state = get(themeState);
      return {
        loading: false,
        success: result.ok,
        error: result.error,
        warnings: Array.isArray(result.warnings) ? result.warnings : state.warnings,
        operation: result.operation,
        lastResult: result
      };
    })()
  );

  return result;
}

function ensureStyleElement() {
  if (styleEl) return styleEl;

  styleEl = document.createElement('style');
  styleEl.id = 'app-theme';
  document.head.appendChild(styleEl);
  return styleEl;
}

function unpackThemeCSSResult(result) {
  if (Array.isArray(result)) {
    const css = typeof result[0] === 'string' ? result[0] : '';
    const warnings = Array.isArray(result[1]) ? result[1] : [];
    return { css, warnings };
  }

  if (typeof result === 'string') {
    return { css: result, warnings: [] };
  }

  return { css: '', warnings: [] };
}

function formatThemeWarnings(errors) {
  return (errors || [])
    .map((entry) => {
      if (typeof entry === 'string') return entry;
      if (entry && typeof entry.msg === 'string' && entry.msg) return entry.msg;
      return JSON.stringify(entry);
    })
    .filter(Boolean);
}

export async function applyTheme(name, options = {}) {
  const rollbackOnError = options.rollbackOnError !== false;

  const themeName = normalizeThemeName(name);
  const previousTheme = normalizeThemeName(get(currentTheme));
  const hadStyleElement = Boolean(styleEl);
  const previousCSS = hadStyleElement ? styleEl.textContent : '';
  const token = startThemeOperation(THEME_OP_APPLY, { preserveWarnings: false });

  try {
    const result = await LoadThemeCSS(themeName);
    const { css, warnings } = unpackThemeCSSResult(result);
    const formattedWarnings = formatThemeWarnings(warnings);

    if (!css) {
      const fromWarnings = formattedWarnings.join('; ').trim();
      throw new Error(fromWarnings || tr('ERRORS.theme.compile', { theme: themeName }));
    }

    ensureStyleElement().textContent = css;
    currentTheme.set(themeName);

    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_APPLY, {
        ok: true,
        data: { theme: themeName, css },
        warnings: formattedWarnings
      })
    );
  } catch (error) {
    const message = toLocalizedError(
      error,
      'ERRORS.theme.apply',
      tr('ERRORS.theme.apply', { theme: themeName }),
      { theme: themeName }
    );
    let rolledBack = false;

    if (rollbackOnError) {
      if (hadStyleElement && styleEl) {
        styleEl.textContent = previousCSS || '';
      }
      currentTheme.set(previousTheme);
      rolledBack = true;
    }

    const result = createThemeResult(THEME_OP_APPLY, {
      ok: false,
      error: message,
      warnings: [],
      rolledBack
    });
    return finishThemeOperation(token, result);
  }
}

export async function getThemes() {
  const token = startThemeOperation(THEME_OP_LIST, { preserveWarnings: true });

  try {
    const themes = normalizeThemeList(await GetThemes());
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_LIST, {
        ok: true,
        data: themes
      })
    );
  } catch (error) {
    const message = toLocalizedError(error, 'ERRORS.theme.load_themes');
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_LIST, {
        ok: false,
        error: message,
        data: []
      })
    );
  }
}

export async function getRawTheme(name) {
  const themeName = normalizeThemeName(name);
  const token = startThemeOperation(THEME_OP_LOAD_RAW, { preserveWarnings: true });

  try {
    const theme = await LoadRawTheme(themeName);
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_LOAD_RAW, {
        ok: true,
        data: theme
      })
    );
  } catch (error) {
    const message = toLocalizedError(
      error,
      'ERRORS.theme.load_theme',
      tr('ERRORS.theme.load_theme', { theme: themeName }),
      { theme: themeName }
    );
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_LOAD_RAW, {
        ok: false,
        error: message
      })
    );
  }
}

export async function createTheme(name, baseTheme) {
  const themeName = normalizeThemeName(name, '');
  const baseName = normalizeThemeName(baseTheme);
  const token = startThemeOperation(THEME_OP_CREATE, { preserveWarnings: true });

  if (!themeName) {
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_CREATE, {
        ok: false,
        error: tr('ERRORS.theme.name_required')
      })
    );
  }

  try {
    await CreateTheme(themeName, baseName);
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_CREATE, {
        ok: true,
        data: { name: themeName, baseTheme: baseName }
      })
    );
  } catch (error) {
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_CREATE, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.theme.create',
          tr('ERRORS.theme.create', { theme: themeName }),
          { theme: themeName }
        )
      })
    );
  }
}

export async function saveTheme(name, themeData) {
  const themeName = normalizeThemeName(name, '');
  const token = startThemeOperation(THEME_OP_SAVE, { preserveWarnings: true });

  if (!themeName) {
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_SAVE, {
        ok: false,
        error: tr('ERRORS.theme.name_required')
      })
    );
  }

  try {
    await SaveTheme(themeName, themeData);
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_SAVE, {
        ok: true,
        data: { name: themeName }
      })
    );
  } catch (error) {
    return finishThemeOperation(
      token,
      createThemeResult(THEME_OP_SAVE, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.theme.save',
          tr('ERRORS.theme.save', { theme: themeName }),
          { theme: themeName }
        )
      })
    );
  }
}
