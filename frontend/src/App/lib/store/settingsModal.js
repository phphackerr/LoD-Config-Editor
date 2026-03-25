import { get, writable } from 'svelte/store';

export const isSettingsOpen = writable(false);
export const activeSettingsTab = writable('');

const SETTINGS_MODAL_OP_OPEN = 'settings-modal:open';
const SETTINGS_MODAL_OP_CLOSE = 'settings-modal:close';
const SETTINGS_MODAL_OP_TOGGLE = 'settings-modal:toggle';
const SETTINGS_MODAL_OP_SET_TAB = 'settings-modal:set-tab';

const initialSettingsModalState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null,
  open: false,
  activeTab: ''
};

export const settingsModalState = writable({
  ...initialSettingsModalState
});

let operationSeq = 0;

function patchSettingsModalState(patch) {
  settingsModalState.update((state) => ({ ...state, ...patch }));
}

function createSettingsModalResult(
  operation,
  { ok, data = null, error = null, rolledBack = false }
) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function normalizeTabName(value) {
  return String(value ?? '').trim();
}

function startSettingsModalOperation(operation, { loading = false } = {}) {
  operationSeq += 1;
  const token = operationSeq;

  patchSettingsModalState({
    loading,
    success: false,
    error: null,
    operation
  });

  return token;
}

function finishSettingsModalOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  patchSettingsModalState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    open: get(isSettingsOpen),
    activeTab: get(activeSettingsTab),
    ...patch
  });
  return result;
}

function syncSettingsModalSnapshot() {
  patchSettingsModalState({
    open: get(isSettingsOpen),
    activeTab: get(activeSettingsTab)
  });
}

export function setActiveSettingsTab(tabName = '') {
  const token = startSettingsModalOperation(SETTINGS_MODAL_OP_SET_TAB, { loading: false });
  const nextTab = normalizeTabName(tabName);
  activeSettingsTab.set(nextTab);
  return finishSettingsModalOperation(
    token,
    createSettingsModalResult(SETTINGS_MODAL_OP_SET_TAB, {
      ok: true,
      data: {
        activeTab: nextTab
      }
    })
  );
}

export function openSettings(tabName = '') {
  const token = startSettingsModalOperation(SETTINGS_MODAL_OP_OPEN, { loading: false });
  const nextTab = normalizeTabName(tabName);
  if (nextTab) {
    activeSettingsTab.set(nextTab);
  }
  isSettingsOpen.set(true);
  return finishSettingsModalOperation(
    token,
    createSettingsModalResult(SETTINGS_MODAL_OP_OPEN, {
      ok: true,
      data: {
        open: true,
        activeTab: get(activeSettingsTab)
      }
    })
  );
}

export function closeSettings() {
  const token = startSettingsModalOperation(SETTINGS_MODAL_OP_CLOSE, { loading: false });
  isSettingsOpen.set(false);
  return finishSettingsModalOperation(
    token,
    createSettingsModalResult(SETTINGS_MODAL_OP_CLOSE, {
      ok: true,
      data: {
        open: false,
        activeTab: get(activeSettingsTab)
      }
    })
  );
}

export function toggleSettings() {
  const token = startSettingsModalOperation(SETTINGS_MODAL_OP_TOGGLE, { loading: false });
  isSettingsOpen.update((value) => !value);
  return finishSettingsModalOperation(
    token,
    createSettingsModalResult(SETTINGS_MODAL_OP_TOGGLE, {
      ok: true,
      data: {
        open: get(isSettingsOpen),
        activeTab: get(activeSettingsTab)
      }
    })
  );
}

const unsubscribeIsOpen = isSettingsOpen.subscribe(() => {
  syncSettingsModalSnapshot();
});
const unsubscribeActiveTab = activeSettingsTab.subscribe(() => {
  syncSettingsModalSnapshot();
});

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    unsubscribeIsOpen();
    unsubscribeActiveTab();
  });
}
