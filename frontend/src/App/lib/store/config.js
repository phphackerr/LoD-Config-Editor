import { get, writable } from 'svelte/store';
import {
  GetConfigValue,
  IsConfigAvailable,
  LoadConfig,
  SetConfigValue
} from '/bindings/lce/backend/config_editor/configeditor';
import { appSettings } from './appSettings';
import { normalizePath, samePath, toLocalizedError, tr } from './storeUtils';

const initialConfigStore = {
  loading: false,
  error: null,
  data: null,
  path: null
};

const initialConfigState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null
};

const CONFIG_OP_RESET = 'config:reset';
const CONFIG_OP_LOAD = 'config:load';
const CONFIG_OP_CHECK_AVAILABLE = 'config:check-available';
const CONFIG_OP_SAVE = 'config:save';

export const configStore = writable({ ...initialConfigStore });
export const configState = writable({ ...initialConfigState });

let loadRequestId = 0;
let lastLoadedPath = null;
let operationSeq = 0;

function updateConfigStore(patch) {
  configStore.update((state) => ({ ...state, ...patch }));
}

function configPathFromGamePath(gamePath) {
  const base = String(gamePath || '').trim();
  if (!base) return null;
  return `${base}\\config.lod.ini`;
}

function updateConfigState(patch) {
  configState.update((state) => ({ ...state, ...patch }));
}

function createConfigResult(operation, { ok, data = null, error = null, rolledBack = false }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function startConfigOperation(operation, { loading = true } = {}) {
  operationSeq += 1;
  const token = operationSeq;
  updateConfigState({
    loading,
    success: false,
    error: null,
    operation
  });
  return token;
}

function finishConfigOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  updateConfigState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    ...patch
  });
  return result;
}

export function resetConfig() {
  const token = startConfigOperation(CONFIG_OP_RESET, { loading: false });
  loadRequestId += 1;
  lastLoadedPath = null;
  configStore.set({ ...initialConfigStore });
  return finishConfigOperation(
    token,
    createConfigResult(CONFIG_OP_RESET, {
      ok: true,
      data: null
    })
  );
}

export async function loadConfig(path) {
  const token = startConfigOperation(CONFIG_OP_LOAD);
  const normalizedPath = normalizePath(path);
  if (!normalizedPath) {
    loadRequestId += 1;
    lastLoadedPath = null;
    configStore.set({ ...initialConfigStore });
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_LOAD, {
        ok: true,
        data: null
      })
    );
  }

  const requestId = ++loadRequestId;
  configStore.set({
    loading: true,
    error: null,
    data: null,
    path
  });

  try {
    const config = await LoadConfig(path);
    if (requestId !== loadRequestId) {
      return createConfigResult(CONFIG_OP_LOAD, {
        ok: false,
        error: tr('ERRORS.config.stale_load_ignored')
      });
    }

    updateConfigStore({
      loading: false,
      error: null,
      data: config,
      path
    });
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_LOAD, {
        ok: true,
        data: {
          path,
          config
        }
      })
    );
  } catch (error) {
    if (requestId !== loadRequestId) {
      return createConfigResult(CONFIG_OP_LOAD, {
        ok: false,
        error: tr('ERRORS.config.stale_load_ignored')
      });
    }

    const message = toLocalizedError(error, 'ERRORS.config.load');

    updateConfigStore({
      loading: false,
      error: message,
      data: null,
      path
    });
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_LOAD, {
        ok: false,
        data: {
          path,
          config: null
        },
        error: message
      })
    );
  }
}

export async function checkConfigAvailability() {
  const token = startConfigOperation(CONFIG_OP_CHECK_AVAILABLE, { loading: false });

  try {
    const available = await IsConfigAvailable();
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_CHECK_AVAILABLE, {
        ok: true,
        data: Boolean(available)
      })
    );
  } catch (error) {
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_CHECK_AVAILABLE, {
        ok: false,
        data: false,
        error: toLocalizedError(error, 'ERRORS.config.check_availability')
      })
    );
  }
}

export async function isConfigAvailable() {
  const result = await checkConfigAvailability();
  return result.ok ? Boolean(result.data) : false;
}

export async function getConfigValue(section, option) {
  try {
    return await GetConfigValue(section, option);
  } catch {
    return '';
  }
}

export async function setConfigValue(section, option, value) {
  const result = await saveConfigValue(section, option, value);
  return result.ok;
}

export async function saveConfigValue(section, option, value, fallbackMessage = '') {
  const token = startConfigOperation(CONFIG_OP_SAVE, { loading: false });

  try {
    await SetConfigValue(section, option, value);

    configStore.update((state) => {
      if (!state.data) return { ...state, error: null };

      const sectionData = state.data[section] ? { ...state.data[section] } : {};
      sectionData[option] = value;

      return {
        ...state,
        error: null,
        data: {
          ...state.data,
          [section]: sectionData
        }
      };
    });
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_SAVE, {
        ok: true,
        data: {
          section,
          option,
          value
        }
      })
    );
  } catch (error) {
    const fallback = fallbackMessage || tr('ERRORS.config.save_option', { section, option });
    const message = toLocalizedError(error, 'ERRORS.config.save_option', fallback, {
      section,
      option
    });
    updateConfigStore({ error: message });
    return finishConfigOperation(
      token,
      createConfigResult(CONFIG_OP_SAVE, {
        ok: false,
        data: {
          section,
          option,
          value
        },
        error: message
      })
    );
  }
}

const unsubscribeSettings = appSettings.subscribe((settings) => {
  const nextPath = configPathFromGamePath(settings?.game_path);

  if (samePath(nextPath, lastLoadedPath)) {
    return;
  }

  if (nextPath) {
    loadConfig(nextPath);
    lastLoadedPath = nextPath;
    return;
  }

  resetConfig();
});

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    unsubscribeSettings();
  });
}

export function getCurrentConfigPath() {
  return get(configStore).path;
}
