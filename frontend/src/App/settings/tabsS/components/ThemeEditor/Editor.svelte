<script>
  import { Window } from '@wailsio/runtime';

  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import {
    draftTheme,
    isDirty,
    saveDraftTheme,
    resetDraftTheme,
    openThemeEditor,
    themeEditorState
  } from '../../../../lib/store/themeEditor';

  import { currentTheme } from '../../../../lib/store/theming';
  import { notifyError, notifyInfo, notifySuccess } from '../../../../lib/store/notifications';

  import TokenList from './TokenList.svelte';
  import TokenEditor from './TokenEditor.svelte';

  let themeName = '';
  let openRequest = 0;

  onMount(() => {
    const unsub = currentTheme.subscribe((id) => {
      if (id) {
        themeName = id;
        openRequest += 1;
        const requestID = openRequest;

        openThemeEditor(id).then((result) => {
          if (requestID !== openRequest || result.ok) return;
          notifyError(result.error || $t('ERRORS.theme_editor.open'));
        });
      }
    });

    return () => {
      openRequest += 1;
      unsub();
    };
  });

  async function handleSave() {
    const result = await saveDraftTheme();
    if (!result.ok) {
      notifyError(result.error || $t('ERRORS.theme_editor.save'));
      return;
    }
    notifySuccess($t('SETTINGS.THEME.saved_success'));
  }

  function handleReset() {
    const result = resetDraftTheme();
    if (!result.ok) {
      notifyError(result.error || $t('ERRORS.theme_editor.reset'));
      return;
    }
    notifyInfo($t('SETTINGS.THEME_EDITOR.reset_done'));
  }
</script>

<div class="editor-root">
  <header class="header">
    <div class="heading">
      <div class="title">{$t('SETTINGS.THEME_EDITOR.title')}: {themeName}</div>
      <div class="subtitle">
        {$t('SETTINGS.THEME_EDITOR.tokens')}: {$draftTheme?.tokens?.length || 0}
      </div>
    </div>

    <div class="preview" aria-hidden="true">
      <div class="preview-card">
        <div class="preview-line strong">{$t('SETTINGS.THEME_EDITOR.preview')}</div>
        <div class="preview-line">{$t('SETTINGS.THEME_EDITOR.preview_hint')}</div>
        <div class="preview-actions">
          <span class="pill primary"></span>
          <span class="pill secondary"></span>
          <span class="pill danger"></span>
        </div>
      </div>
    </div>

    <div class="actions">
      <button
        class="btn secondary"
        on:click={handleReset}
        disabled={!$isDirty || $themeEditorState.loading}
      >
        {$t('COMMON.reset')}
      </button>
      <button
        class="btn primary"
        on:click={handleSave}
        disabled={!$isDirty || $themeEditorState.loading || $themeEditorState.saving}
      >
        {$themeEditorState.saving ? `${$t('COMMON.saving')}...` : $t('COMMON.save')}
      </button>
      <button class="btn close" on:click={Window.Hide}> {$t('COMMON.close')} </button>
    </div>
  </header>

  {#if $themeEditorState.error}
    <div class="status status-error">{$themeEditorState.error}</div>
  {:else if $themeEditorState.loading}
    <div class="status">{$t('SETTINGS.THEME_EDITOR.loading_theme')}...</div>
  {:else if $themeEditorState.saving}
    <div class="status">{$t('COMMON.saving')}...</div>
  {:else if $themeEditorState.success && $themeEditorState.operation === 'editor:save'}
    <div class="status status-success">{$t('SETTINGS.THEME.saved_success')}</div>
  {/if}

  <div class="body">
    <TokenList />
    <TokenEditor />
  </div>
</div>

<style>
  .editor-root {
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--theme-editor-bg, var(--bg-color-dark, #1e1e1e));
    color: var(--text-color-primary, #eee);
  }

  .header {
    --wails-draggable: drag;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    border-bottom: 1px solid var(--theme-editor-card-border, var(--border-color, #333));
    gap: 12px;
  }

  .heading {
    min-width: 220px;
  }

  .title {
    font-weight: 600;
  }

  .subtitle {
    margin-top: 2px;
    font-size: 11px;
    color: var(--theme-editor-hint-text, var(--text-color-muted, #999));
  }

  .preview {
    --wails-draggable: drag;
    flex: 1;
    display: flex;
    justify-content: center;
    min-width: 0;
  }

  .preview-card {
    width: min(420px, 100%);
    border: 1px solid
      var(--theme-editor-card-border, var(--app-panel-border, rgba(64, 64, 64, 0.5)));
    border-radius: 10px;
    padding: 8px 12px;
    background: var(--theme-editor-preview-bg, var(--surface-panel, rgba(20, 20, 20, 0.85)));
    box-shadow: var(
      --theme-editor-preview-shadow,
      var(--theme-editor-card-shadow, 0 12px 30px rgba(0, 0, 0, 0.25))
    );
  }

  .preview-line {
    font-size: 11px;
    color: var(--theme-editor-hint-text, var(--text-color-muted, #9ca3af));
    line-height: 1.2;
  }

  .preview-line.strong {
    font-weight: 700;
    color: var(--text-color-primary, #fff);
  }

  .preview-actions {
    margin-top: 6px;
    display: flex;
    gap: 6px;
  }

  .pill {
    width: 24px;
    height: 8px;
    border-radius: 999px;
    display: inline-block;
  }

  .pill.primary {
    background: var(--action-primary-bg, #3ba475);
  }

  .pill.secondary {
    background: var(--action-secondary-bg, #555);
  }

  .pill.danger {
    background: var(--action-danger-bg, #ca3333);
  }

  .actions {
    --wails-draggable: no-drag;
    display: flex;
    gap: 8px;
  }

  .body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 280px 1fr;
    overflow: hidden;
    background: var(--theme-editor-body-bg, var(--surface-panel-muted, rgba(0, 0, 0, 0.28)));
  }

  .status {
    padding: 8px 14px;
    font-size: 12px;
    color: var(--theme-editor-hint-text, var(--text-color-muted, #b8b8b8));
    border-bottom: 1px solid var(--theme-editor-card-border, var(--border-color, #2c2c2c));
  }

  .status-error {
    color: var(--status-error-text, #fecaca);
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
  }

  .status-success {
    color: var(--status-success-text, #9be5be);
    background: var(--status-success-bg, rgba(36, 147, 79, 0.18));
  }

  .btn {
    padding: 6px 12px;
    border-radius: 4px;
    border: none;
    cursor: pointer;
  }

  .btn:hover {
    opacity: 0.85;
  }

  .btn.primary {
    background: var(--action-primary-bg, #3ba475);
    color: var(--text-color-primary, #fff);
  }

  .btn.secondary {
    background: var(--action-secondary-bg, #333);
    color: var(--text-color-primary, #fff);
  }

  .btn.close {
    background: var(--action-danger-bg, #ca3333);
    color: var(--text-color-primary, #fff);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
