import { writable, get } from 'svelte/store';
import { Events } from '@wailsio/runtime';
import { LoadRawTheme, SaveTheme } from '/bindings/lce/backend/theming/themeservice';
import { toLocalizedError, tr } from './storeUtils';
import { applyTheme, currentTheme } from './theming';

// --- STATE ---

export const editorThemeId = writable(null);
export const originalTheme = writable(null);
export const draftTheme = writable(null);
export const selectedTokenId = writable(null);
export const isDirty = writable(false);

const initialThemeEditorState = {
  loading: false,
  saving: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null
};

export const themeEditorState = writable({
  ...initialThemeEditorState
});

let openSeq = 0;

const EDITOR_OP_OPEN = 'editor:open';
const EDITOR_OP_SAVE = 'editor:save';
const EDITOR_OP_RESET = 'editor:reset';
const EDITOR_OP_UPDATE_TOKEN = 'editor:update-token';
const EDITOR_OP_SELECT_TOKEN = 'editor:select-token';

function patchThemeEditorState(patch) {
  themeEditorState.update((state) => ({ ...state, ...patch }));
}

function cloneTheme(value) {
  if (value == null) return value;
  if (typeof structuredClone === 'function') {
    return structuredClone(value);
  }
  return JSON.parse(JSON.stringify(value));
}

function createEditorResult(operation, { ok, data = null, error = null, rolledBack = false }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function finishEditorOperation(result, patch = {}) {
  patchThemeEditorState({
    loading: false,
    saving: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    ...patch
  });
  return result;
}

// --- ACTIONS ---

export async function openThemeEditor(themeId) {
  const nextThemeID = String(themeId || '').trim();
  if (!nextThemeID) {
    return finishEditorOperation(
      createEditorResult(EDITOR_OP_OPEN, {
        ok: false,
        error: tr('ERRORS.theme_editor.theme_id_required')
      })
    );
  }

  const previous = {
    themeId: get(editorThemeId),
    original: get(originalTheme),
    draft: get(draftTheme),
    selectedTokenId: get(selectedTokenId),
    isDirty: get(isDirty)
  };

  openSeq += 1;
  const requestID = openSeq;
  patchThemeEditorState({
    loading: true,
    saving: false,
    success: false,
    error: null,
    operation: EDITOR_OP_OPEN
  });

  try {
    const theme = await LoadRawTheme(nextThemeID);

    if (requestID !== openSeq) {
      return createEditorResult(EDITOR_OP_OPEN, {
        ok: false,
        error: tr('ERRORS.theme_editor.stale_open_ignored')
      });
    }

    editorThemeId.set(nextThemeID);
    originalTheme.set(theme);
    draftTheme.set(cloneTheme(theme));
    selectedTokenId.set(null);
    isDirty.set(false);

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_OPEN, {
        ok: true,
        data: theme
      })
    );
  } catch (error) {
    if (requestID !== openSeq) {
      return createEditorResult(EDITOR_OP_OPEN, {
        ok: false,
        error: tr('ERRORS.theme_editor.stale_open_ignored')
      });
    }

    editorThemeId.set(previous.themeId);
    originalTheme.set(previous.original);
    draftTheme.set(previous.draft);
    selectedTokenId.set(previous.selectedTokenId);
    isDirty.set(previous.isDirty);

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_OPEN, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.theme_editor.open',
          tr('ERRORS.theme_editor.open', { theme: nextThemeID }),
          { theme: nextThemeID }
        ),
        rolledBack: true
      })
    );
  }
}

export function selectToken(tokenId) {
  selectedTokenId.set(tokenId || null);

  const result = createEditorResult(EDITOR_OP_SELECT_TOKEN, {
    ok: true,
    data: { tokenId: tokenId || null }
  });
  patchThemeEditorState({
    success: true,
    error: null,
    operation: result.operation,
    lastResult: result
  });
  return result;
}

// updater: (token) => token
export function updateToken(tokenId, updater) {
  const normalizedID = String(tokenId || '').trim();
  if (!normalizedID) {
    const result = createEditorResult(EDITOR_OP_UPDATE_TOKEN, {
      ok: false,
      error: tr('ERRORS.theme_editor.token_id_required')
    });
    patchThemeEditorState({
      success: false,
      error: result.error,
      operation: result.operation,
      lastResult: result
    });
    return result;
  }

  if (typeof updater !== 'function') {
    const result = createEditorResult(EDITOR_OP_UPDATE_TOKEN, {
      ok: false,
      error: tr('ERRORS.theme_editor.token_updater_required')
    });
    patchThemeEditorState({
      success: false,
      error: result.error,
      operation: result.operation,
      lastResult: result
    });
    return result;
  }

  let updatedToken = null;
  let updated = false;

  draftTheme.update((theme) => {
    if (!theme) return theme;

    const index = theme.tokens.findIndex((t) => t.id === normalizedID);
    if (index === -1) return theme;

    const next = cloneTheme(theme);
    next.tokens[index] = updater(next.tokens[index]);
    updatedToken = next.tokens[index];
    updated = true;
    return next;
  });

  if (!updated) {
    const result = createEditorResult(EDITOR_OP_UPDATE_TOKEN, {
      ok: false,
      error: tr('ERRORS.theme_editor.token_not_found', { token: normalizedID })
    });
    patchThemeEditorState({
      success: false,
      error: result.error,
      operation: result.operation,
      lastResult: result
    });
    return result;
  }

  isDirty.set(true);

  const result = createEditorResult(EDITOR_OP_UPDATE_TOKEN, {
    ok: true,
    data: updatedToken
  });
  patchThemeEditorState({
    success: true,
    error: null,
    operation: result.operation,
    lastResult: result
  });
  return result;
}

export async function saveDraftTheme() {
  const themeId = get(editorThemeId);
  const theme = get(draftTheme);
  const previousOriginal = get(originalTheme);
  const previousDirty = get(isDirty);

  if (!themeId || !theme) {
    return finishEditorOperation(
      createEditorResult(EDITOR_OP_SAVE, {
        ok: false,
        error: tr('ERRORS.theme_editor.no_opened_theme')
      })
    );
  }

  patchThemeEditorState({
    loading: false,
    saving: true,
    success: false,
    error: null,
    operation: EDITOR_OP_SAVE
  });

  try {
    await SaveTheme(themeId, theme);

    originalTheme.set(cloneTheme(theme));
    isDirty.set(false);

    const activeTheme = get(currentTheme);
    if (activeTheme === themeId) {
      const applyResult = await applyTheme(themeId, { rollbackOnError: false });
      if (!applyResult.ok) {
        return finishEditorOperation(
          createEditorResult(EDITOR_OP_SAVE, {
            ok: false,
            error:
              applyResult.error || tr('ERRORS.theme_editor.save_apply_failed', { theme: themeId })
          })
        );
      }

      // Broadcast to other windows so the active theme updates without restart.
      try {
        await Events.Emit('theme:updated', { theme: themeId });
      } catch {
        // Best-effort event; do not fail save flow if broadcast fails.
      }
    }

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_SAVE, {
        ok: true,
        data: theme
      })
    );
  } catch (error) {
    originalTheme.set(previousOriginal);
    isDirty.set(previousDirty);

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_SAVE, {
        ok: false,
        error: toLocalizedError(
          error,
          'ERRORS.theme_editor.save',
          tr('ERRORS.theme_editor.save', { theme: themeId }),
          { theme: themeId }
        ),
        rolledBack: true
      })
    );
  }
}

export function resetDraftTheme() {
  const orig = get(originalTheme);
  if (!orig) {
    return finishEditorOperation(
      createEditorResult(EDITOR_OP_RESET, {
        ok: false,
        error: tr('ERRORS.theme_editor.no_source_theme')
      })
    );
  }

  const previousDraft = get(draftTheme);
  const previousDirty = get(isDirty);

  try {
    draftTheme.set(cloneTheme(orig));
    isDirty.set(false);

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_RESET, {
        ok: true,
        data: get(draftTheme)
      })
    );
  } catch (error) {
    draftTheme.set(previousDraft);
    isDirty.set(previousDirty);

    return finishEditorOperation(
      createEditorResult(EDITOR_OP_RESET, {
        ok: false,
        error: toLocalizedError(error, 'ERRORS.theme_editor.reset'),
        rolledBack: true
      })
    );
  }
}
