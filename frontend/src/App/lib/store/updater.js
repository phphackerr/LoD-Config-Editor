import { writable } from 'svelte/store';
import { Events } from '@wailsio/runtime';
import {
  CheckForComponentUpdates,
  CheckForUpdates,
  DoUpdate,
  RestartApp,
  UpdateComponent
} from '/bindings/lce/backend/updater/updater';
import { toLocalizedError, tr } from './storeUtils';

const initialUpdaterState = {
  loading: false,
  success: false,
  operation: '',
  lastResult: null,
  available: false,
  version: '',
  body: '',
  checking: false,
  downloading: false,
  progress: 0,
  error: null,
  componentUpdates: [],
  readyToRestart: false
};

export const updaterStore = writable({ ...initialUpdaterState });
let operationSeq = 0;

const UPDATER_OP_CHECK = 'updater:check';
const UPDATER_OP_DOWNLOAD_APP = 'updater:download-app';
const UPDATER_OP_RESTART = 'updater:restart';
const UPDATER_OP_UPDATE_COMPONENT = 'updater:update-component';

function patchUpdaterStore(patch) {
  updaterStore.update((state) => ({ ...state, ...patch }));
}

function createUpdaterResult(operation, { ok, data = null, error = null }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null
  };
}

function startUpdaterOperation(operation, patch = {}) {
  operationSeq += 1;
  const token = operationSeq;

  patchUpdaterStore({
    loading: true,
    success: false,
    error: null,
    operation,
    ...patch
  });
  return token;
}

function finishUpdaterOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  patchUpdaterStore({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    ...patch
  });
  return result;
}

export async function checkForUpdates() {
  const token = startUpdaterOperation(UPDATER_OP_CHECK, { checking: true });

  try {
    const [appResult, componentUpdates] = await Promise.all([
      CheckForUpdates(),
      CheckForComponentUpdates()
    ]);

    const nextState = {
      checking: false,
      available: Boolean(appResult?.available),
      version: appResult?.version ?? '',
      body: appResult?.body ?? '',
      componentUpdates: Array.isArray(componentUpdates) ? componentUpdates : [],
      error: appResult?.error || null
    };

    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_CHECK, {
        ok: !nextState.error,
        data: {
          app: appResult || {},
          components: nextState.componentUpdates
        },
        error: nextState.error || null
      }),
      nextState
    );
  } catch (error) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_CHECK, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.updater.check_updates')
      }),
      { checking: false }
    );
  }
}

export async function doUpdate(version) {
  const targetVersion = String(version || '').trim();
  const token = startUpdaterOperation(UPDATER_OP_DOWNLOAD_APP, {
    downloading: true,
    readyToRestart: false,
    progress: 0
  });

  if (!targetVersion) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_DOWNLOAD_APP, {
        ok: false,
        error: tr('ERRORS.updater.version_required')
      }),
      { downloading: false }
    );
  }

  try {
    await DoUpdate(targetVersion);
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_DOWNLOAD_APP, {
        ok: true,
        data: { version: targetVersion }
      }),
      {
        downloading: false,
        progress: 100,
        readyToRestart: true
      }
    );
  } catch (error) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_DOWNLOAD_APP, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.updater.download_update')
      }),
      { downloading: false }
    );
  }
}

export async function restartApp() {
  const token = startUpdaterOperation(UPDATER_OP_RESTART);

  try {
    await RestartApp();
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_RESTART, {
        ok: true
      })
    );
  } catch (error) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_RESTART, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.updater.restart_app')
      })
    );
  }
}

export async function updateComponent(component) {
  const token = startUpdaterOperation(UPDATER_OP_UPDATE_COMPONENT, {
    downloading: true
  });

  if (!component || !component.name) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_UPDATE_COMPONENT, {
        ok: false,
        error: tr('ERRORS.updater.invalid_component_request')
      }),
      { downloading: false }
    );
  }

  try {
    await UpdateComponent(component);

    let remainingComponents = [];
    updaterStore.update((state) => {
      remainingComponents = state.componentUpdates.filter(
        (c) => c.name !== component.name || c.type !== component.type
      );
      return {
        ...state,
        componentUpdates: remainingComponents
      };
    });

    const result = finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_UPDATE_COMPONENT, {
        ok: true,
        data: {
          updatedComponent: component,
          remainingComponents
        }
      }),
      { downloading: false }
    );

    window.location.reload();
    return result;
  } catch (error) {
    return finishUpdaterOperation(
      token,
      createUpdaterResult(UPDATER_OP_UPDATE_COMPONENT, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.updater.update_component')
      }),
      { downloading: false }
    );
  }
}

let offProgressListener = null;

function ensureProgressListener() {
  if (offProgressListener) return;

  const listener = (event) => {
    const payload = Array.isArray(event?.data) ? event.data[0] : event?.data;
    if (!payload || payload.status !== 'downloading') return;
    patchUpdaterStore({ progress: Number(payload.percent) || 0 });
  };

  offProgressListener = Events.On('update:progress', listener);
}

ensureProgressListener();

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    if (offProgressListener) {
      offProgressListener();
      offProgressListener = null;
    }
  });
}
