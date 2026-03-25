import { get, writable } from 'svelte/store';
import {
  GetDefaultSettings,
  GetOption,
  GetSettings,
  UpdateSettings
} from '/bindings/lce/backend/app_settings/appsettings';
import { CheckAndFindPaths } from '/bindings/lce/backend/paths_scanner/scanner';
import { samePath, toLocalizedError, uniquePaths } from './storeUtils';

export const DEFAULT_APP_SETTINGS = {
  width: 1600,
  height: 900,
  language: 'en',
  game_path: '',
  first_run: true,
  all_paths: [],
  theme: 'default',
  windowed_mode: false
};

export const appSettings = writable({ ...DEFAULT_APP_SETTINGS });
export const initialAppSettingsState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null
};
export const appSettingsState = writable({ ...initialAppSettingsState });

const APP_SETTINGS_OP_LOAD = 'appSettings:load';
const APP_SETTINGS_OP_UPDATE = 'appSettings:update';
const APP_SETTINGS_OP_RESET = 'appSettings:reset';
const APP_SETTINGS_OP_GET_OPTION = 'appSettings:get-option';
const APP_SETTINGS_OP_RUN_SCANNER = 'appSettings:run-scanner';
const APP_SETTINGS_OP_DELETE_PATH = 'appSettings:delete-path';

let operationSeq = 0;

function normalizeSettings(input) {
  const base = { ...DEFAULT_APP_SETTINGS, ...(input || {}) };
  base.all_paths = uniquePaths(base.all_paths);
  return base;
}

function setAppSettings(next) {
  appSettings.set(normalizeSettings(next));
}

function patchAppSettingsState(patch) {
  appSettingsState.update((state) => ({ ...state, ...patch }));
}

function createAppSettingsResult(operation, { ok, data = null, error = null, rolledBack = false }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function startAppSettingsOperation(operation) {
  operationSeq += 1;
  const token = operationSeq;
  patchAppSettingsState({
    loading: true,
    success: false,
    error: null,
    operation
  });
  return token;
}

function finishAppSettingsOperation(token, result) {
  if (token !== operationSeq) {
    return result;
  }

  patchAppSettingsState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result
  });
  return result;
}

async function loadDefaultSettingsNoState() {
  try {
    return normalizeSettings(await GetDefaultSettings());
  } catch {
    return { ...DEFAULT_APP_SETTINGS };
  }
}

async function persistSettings(next) {
  const updated = normalizeSettings(await UpdateSettings(next));
  setAppSettings(updated);
  return updated;
}

export async function loadSettings() {
  const token = startAppSettingsOperation(APP_SETTINGS_OP_LOAD);

  try {
    const settings = await GetSettings();
    setAppSettings(settings);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_LOAD, {
        ok: true,
        data: get(appSettings)
      })
    );
  } catch (error) {
    const fallback = await loadDefaultSettingsNoState();
    setAppSettings(fallback);
    const message = toLocalizedError(error, 'ERRORS.app_settings.load');

    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_LOAD, {
        ok: false,
        data: get(appSettings),
        error: message,
        rolledBack: true
      })
    );
  }
}

export async function updateSettings(patch) {
  const current = get(appSettings);
  const next = normalizeSettings({ ...current, ...(patch || {}) });
  const token = startAppSettingsOperation(APP_SETTINGS_OP_UPDATE);

  try {
    const updated = await persistSettings(next);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_UPDATE, {
        ok: true,
        data: updated
      })
    );
  } catch (error) {
    setAppSettings(current);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_UPDATE, {
        ok: false,
        data: current,
        error: toLocalizedError(error, 'ERRORS.app_settings.update'),
        rolledBack: true
      })
    );
  }
}

export async function getSetting(key) {
  const token = startAppSettingsOperation(APP_SETTINGS_OP_GET_OPTION);

  try {
    const value = await GetOption(key);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_GET_OPTION, {
        ok: true,
        data: value
      })
    );
  } catch (error) {
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_GET_OPTION, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.app_settings.get_setting')
      })
    );
  }
}

export async function resetSettings() {
  const token = startAppSettingsOperation(APP_SETTINGS_OP_RESET);

  try {
    const defaults = await loadDefaultSettingsNoState();
    const updated = await persistSettings(defaults);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_RESET, {
        ok: true,
        data: updated
      })
    );
  } catch (error) {
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_RESET, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.app_settings.reset')
      })
    );
  }
}

export async function updateWindowSize(width, height) {
  return updateSettings({ width, height });
}

export async function updateLanguage(language) {
  return updateSettings({ language });
}

export async function updateGamePath(game_path) {
  return updateSettings({ game_path });
}

export async function updateFirstRun(first_run) {
  return updateSettings({ first_run });
}

export async function updateAllPaths(all_paths) {
  return updateSettings({ all_paths: uniquePaths(all_paths) });
}

export async function updateTheme(theme) {
  return updateSettings({ theme });
}

export async function updateWindowedMode(windowed_mode) {
  return updateSettings({ windowed_mode });
}

export async function runScanner() {
  const token = startAppSettingsOperation(APP_SETTINGS_OP_RUN_SCANNER);

  try {
    const found = uniquePaths(await CheckAndFindPaths());
    const current = get(appSettings);

    const patch = {
      all_paths: found,
      first_run: false
    };

    if (found.length === 1) {
      patch.game_path = found[0];
    } else if (current.game_path && !found.some((p) => samePath(p, current.game_path))) {
      patch.game_path = '';
    }

    const desired = normalizeSettings({ ...current, ...patch });
    const updated = await persistSettings(desired);
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_RUN_SCANNER, {
        ok: true,
        data: {
          found,
          settings: updated
        }
      })
    );
  } catch (error) {
    const message = toLocalizedError(error, 'ERRORS.app_settings.run_scanner');
    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_RUN_SCANNER, {
        ok: false,
        data: {
          found: []
        },
        error: message
      })
    );
  }
}

export async function deletePath(pathToDelete) {
  const token = startAppSettingsOperation(APP_SETTINGS_OP_DELETE_PATH);
  const current = get(appSettings);
  const updatedPaths = current.all_paths.filter((p) => !samePath(p, pathToDelete));
  const patch = { all_paths: updatedPaths };

  if (samePath(current.game_path, pathToDelete)) {
    patch.game_path = updatedPaths[0] || '';
  }

  try {
    const desired = normalizeSettings({ ...current, ...patch });
    const updated = await persistSettings(desired);

    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_DELETE_PATH, {
        ok: true,
        data: {
          deletedPath: pathToDelete,
          settings: updated
        }
      })
    );
  } catch (error) {
    setAppSettings(current);

    return finishAppSettingsOperation(
      token,
      createAppSettingsResult(APP_SETTINGS_OP_DELETE_PATH, {
        ok: false,
        data: {
          deletedPath: pathToDelete,
          settings: current
        },
        error: toLocalizedError(error, 'ERRORS.app_settings.delete_path'),
        rolledBack: true
      })
    );
  }
}
