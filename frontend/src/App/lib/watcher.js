import { Events } from '@wailsio/runtime';
import { StartWatching, StopWatching } from '/bindings/lce/backend/config_watcher/configwatcher';
import { get, writable } from 'svelte/store';
import { isInternalChange } from './store/internalChange';
import { samePath, toLocalizedError, tr } from './store/storeUtils';

let currentPath = '';
const callbacks = new Set();
let hasRuntimeListener = false;
let offRuntimeListener = null;
let operationSeq = 0;

const WATCHER_OP_START = 'watcher:start';
const WATCHER_OP_STOP = 'watcher:stop';
const WATCHER_OP_CALLBACK = 'watcher:callback';

const initialWatcherState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null,
  activePath: '',
  listening: false,
  callbackCount: 0
};

export const watcherState = writable({
  ...initialWatcherState
});

function patchWatcherState(patch) {
  watcherState.update((state) => ({ ...state, ...patch }));
}

function createWatcherResult(
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

function startWatcherOperation(operation) {
  operationSeq += 1;
  const token = operationSeq;

  patchWatcherState({
    loading: true,
    success: false,
    error: null,
    operation
  });

  return token;
}

function finishWatcherOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  patchWatcherState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    activePath: currentPath,
    listening: hasRuntimeListener,
    callbackCount: callbacks.size,
    ...patch
  });

  return result;
}

function syncWatcherMeta(patch = {}) {
  patchWatcherState({
    activePath: currentPath,
    listening: hasRuntimeListener,
    callbackCount: callbacks.size,
    ...patch
  });
}

function getEventPath(event) {
  const payload = Array.isArray(event?.data) ? event.data[0] : event?.data;
  return String(payload || '').trim();
}

function handleConfigChangedEvent(event) {
  const filePath = getEventPath(event);
  if (!filePath) return;

  if (get(isInternalChange)) {
    return;
  }

  for (const callback of callbacks) {
    try {
      callback(filePath);
    } catch (error) {
      const message = toLocalizedError(error, 'ERRORS.watcher.callback_failed');
      patchWatcherState({
        loading: false,
        success: false,
        error: message,
        operation: WATCHER_OP_CALLBACK,
        lastResult: createWatcherResult(WATCHER_OP_CALLBACK, {
          ok: false,
          error: message,
          data: { filePath }
        }),
        activePath: currentPath,
        listening: hasRuntimeListener,
        callbackCount: callbacks.size
      });
    }
  }
}

function ensureRuntimeListener() {
  if (hasRuntimeListener) return;

  offRuntimeListener = Events.On('config-changed', handleConfigChangedEvent);
  hasRuntimeListener = true;
  syncWatcherMeta();
}

function removeRuntimeListenerIfUnused() {
  if (!hasRuntimeListener || callbacks.size > 0) return;

  if (typeof offRuntimeListener === 'function') {
    offRuntimeListener();
  }
  offRuntimeListener = null;
  hasRuntimeListener = false;
  syncWatcherMeta();
}

// --- API для UI --- //
export function onConfigChanged(callback) {
  if (typeof callback !== 'function') {
    return { off: () => {} };
  }

  ensureRuntimeListener();
  callbacks.add(callback);
  syncWatcherMeta();

  let disposed = false;
  return {
    off: () => {
      if (disposed) return;
      disposed = true;
      callbacks.delete(callback);
      removeRuntimeListenerIfUnused();
      syncWatcherMeta();
    }
  };
}

export async function startWatcher(path) {
  const token = startWatcherOperation(WATCHER_OP_START);
  const nextPath = String(path || '').trim();
  if (!nextPath) {
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_START, {
        ok: false,
        error: tr('ERRORS.watcher.path_required')
      })
    );
  }

  if (samePath(currentPath, nextPath)) {
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_START, {
        ok: true,
        data: {
          path: currentPath,
          started: false
        }
      })
    );
  }

  const warnings = [];
  if (currentPath) {
    try {
      await StopWatching();
    } catch (err) {
      warnings.push(toLocalizedError(err, 'ERRORS.watcher.stop_previous'));
    } finally {
      currentPath = '';
    }
  }

  try {
    await StartWatching(nextPath, 200);
    currentPath = nextPath;
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_START, {
        ok: true,
        data: {
          path: nextPath,
          started: true
        },
        warnings
      })
    );
  } catch (err) {
    const message = toLocalizedError(err, 'ERRORS.watcher.start');
    currentPath = '';
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_START, {
        ok: false,
        error: message,
        data: {
          path: nextPath,
          started: false
        },
        warnings
      })
    );
  }
}

export async function stopWatcher() {
  const token = startWatcherOperation(WATCHER_OP_STOP);
  if (!currentPath) {
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_STOP, {
        ok: true,
        data: {
          path: '',
          stopped: false
        }
      })
    );
  }

  const previousPath = currentPath;
  try {
    await StopWatching();
    currentPath = '';
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_STOP, {
        ok: true,
        data: {
          path: previousPath,
          stopped: true
        }
      })
    );
  } catch (err) {
    const message = toLocalizedError(err, 'ERRORS.watcher.stop');
    currentPath = '';
    return finishWatcherOperation(
      token,
      createWatcherResult(WATCHER_OP_STOP, {
        ok: false,
        error: message,
        data: {
          path: previousPath,
          stopped: false
        }
      })
    );
  } finally {
    syncWatcherMeta();
  }
}

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    callbacks.clear();
    removeRuntimeListenerIfUnused();
    currentPath = '';
    patchWatcherState({ ...initialWatcherState });
  });
}
