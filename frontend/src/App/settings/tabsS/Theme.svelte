<script module>
  export const tabMetadata = {
    order: 4
  };
</script>

<script>
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import {
    getThemes,
    createTheme,
    applyTheme,
    themeState,
    currentTheme
  } from '../../lib/store/theming';
  import { appSettings, updateTheme } from '../../lib/store/appSettings';
  import { notifyError, notifySuccess } from '../../lib/store/notifications';
  import { OpenThemeEditor } from '/bindings/lce/backend/windows/windowservice';
  import { OpenThemesFolder } from '/bindings/lce/backend/utils/utils';
  import { toErrorMessage } from '../../lib/store/storeUtils';

  let themes = [];
  let selectedTheme = 'default';
  let localError = '';

  // Create state
  let showCreateModal = false;
  let newThemeName = '';
  let baseThemeForCreate = 'default';

  $: isApplying = $themeState.loading && $themeState.operation === 'theme:apply';
  $: isLoadingThemes = $themeState.loading && $themeState.operation === 'theme:list';
  $: isCreatingTheme = $themeState.loading && $themeState.operation === 'theme:create';
  $: controlsDisabled = isApplying || isLoadingThemes || isCreatingTheme;

  onMount(async () => {
    await loadThemes();
    selectedTheme = $appSettings.theme || $currentTheme || 'default';

    if (themes.length > 0 && !themes.includes(selectedTheme)) {
      selectedTheme = themes[0];
    }
  });

  async function openEditor() {
    localError = '';
    try {
      await OpenThemeEditor();
    } catch (error) {
      localError = toErrorMessage(error, $t('ERRORS.theme_editor.open'));
      notifyError(localError);
    }
  }

  async function openThemesFolder() {
    localError = '';
    try {
      await OpenThemesFolder();
    } catch (error) {
      localError = toErrorMessage(error, $t('ERRORS.theme.open_folder'));
      notifyError(localError);
    }
  }

  async function loadThemes() {
    localError = '';
    const result = await getThemes();
    if (!result.ok) {
      themes = [];
      localError = result.error || $t('ERRORS.theme.load_themes');
      return result;
    }

    themes = Array.isArray(result.data) ? result.data : [];

    if (themes.length === 0) {
      localError = $t('SETTINGS.THEME.no_themes_found');
    }

    return result;
  }

  async function handleThemeChange() {
    const nextTheme = String(selectedTheme || '').trim() || 'default';
    const previousTheme = $currentTheme || $appSettings.theme || 'default';

    if (nextTheme === previousTheme) {
      return;
    }

    localError = '';
    const applyResult = await applyTheme(nextTheme, { rollbackOnError: true });
    if (!applyResult.ok) {
      localError = applyResult.error || $t('ERRORS.theme.apply');
      notifyError(localError);
      selectedTheme = previousTheme;
      return;
    }

    const updateResult = await updateTheme(nextTheme);
    if (!updateResult.ok) {
      await applyTheme(previousTheme, { rollbackOnError: false });
      selectedTheme = previousTheme;
      localError = updateResult.error || $t('ERRORS.theme.save_selected');
      notifyError(localError);
      return;
    }

    notifySuccess(
      `${$t('SETTINGS.THEME.theme_prefix')} "${nextTheme}" ${$t('SETTINGS.THEME.status_applied')}`
    );
  }

  async function handleCreateTheme() {
    const name = String(newThemeName || '').trim();
    if (!name) return;

    localError = '';
    const createResult = await createTheme(name, baseThemeForCreate);
    if (!createResult.ok) {
      localError = createResult.error || $t('ERRORS.theme.create');
      notifyError(localError);
      return;
    }

    const loadResult = await loadThemes();
    if (!loadResult?.ok) return;

    selectedTheme = name;
    await handleThemeChange();
    if (localError) return;

    showCreateModal = false;
    newThemeName = '';
    notifySuccess(
      `${$t('SETTINGS.THEME.theme_prefix')} "${name}" ${$t('SETTINGS.THEME.status_created')}`
    );
  }
</script>

<div class="content">
  {#if localError}
    <div class="error-message">{localError}</div>
  {/if}

  {#if isLoadingThemes}
    <div class="warning-message">{$t('COMMON.loading')}...</div>
  {/if}

  {#if $themeState.warnings.length > 0}
    <div class="warning-message">{$themeState.warnings.join('; ')}</div>
  {/if}

  <div class="setting-row">
    <div class="label">{$t('SETTINGS.THEME.theme_tab')}</div>
    <div class="dropdown-wrapper">
      <label class="dropdown">
        <select
          bind:value={selectedTheme}
          on:change={handleThemeChange}
          disabled={controlsDisabled}
        >
          {#each themes as theme}
            <option value={theme}>{theme}</option>
          {/each}
        </select>
      </label>
    </div>
  </div>

  <div class="setting-row">
    <div class="label">{$t('SETTINGS.THEME.tools')}</div>
    <div class="buttons">
      <button class="btn" on:click={() => (showCreateModal = true)} disabled={controlsDisabled}
        >{$t('SETTINGS.THEME.add_theme')}</button
      >
      <button class="btn" on:click={openEditor} disabled={controlsDisabled}
        >{$t('SETTINGS.THEME.editor')}</button
      >
      <button class="btn" on:click={openThemesFolder} disabled={controlsDisabled}
        >{$t('SETTINGS.THEME.open_folder')}</button
      >
    </div>
  </div>
</div>

{#if showCreateModal}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal-backdrop" on:click|self={() => (showCreateModal = false)}>
    <div class="modal" role="dialog" aria-modal="true">
      <h3>{$t('SETTINGS.THEME.add_new_theme')}</h3>
      <div class="field">
        <label>
          {$t('SETTINGS.THEME.name_placeholder')}
          <input bind:value={newThemeName} placeholder={$t('SETTINGS.THEME.example_name')} />
        </label>
      </div>
      <div class="field">
        <label>
          {$t('SETTINGS.THEME.base_theme')}
          <select bind:value={baseThemeForCreate}>
            {#each themes as theme}
              <option value={theme}>{theme}</option>
            {/each}
          </select>
        </label>
      </div>
      <div class="actions">
        <button
          class="btn"
          on:click={handleCreateTheme}
          disabled={isCreatingTheme || String(newThemeName || '').trim().length === 0}
        >
          {isCreatingTheme ? `${$t('COMMON.creating')}...` : $t('SETTINGS.LANGUAGE.create')}
        </button>
        <button
          class="btn secondary"
          on:click={() => (showCreateModal = false)}
          disabled={isCreatingTheme}>{$t('SETTINGS.LANGUAGE.cancel')}</button
        >
      </div>
    </div>
  </div>
{/if}

<style>
  .content {
    padding: 20px;
    color: var(--text-color);
  }

  .error-message,
  .warning-message {
    margin-bottom: 16px;
    padding: 10px 12px;
    border-radius: 6px;
    font-size: 14px;
  }

  .error-message {
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    color: var(--status-error-text, #fecaca);
  }

  .warning-message {
    background: var(--status-warning-bg, rgba(245, 158, 11, 0.12));
    border: 1px solid var(--status-warning-border, rgba(245, 158, 11, 0.35));
    color: var(--status-warning-text, #fde68a);
  }

  .setting-row {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 20px;
  }

  .label {
    font-size: 16px;
    min-width: 100px;
  }

  /* Styles borrowed from Dropdown.svelte */
  .dropdown-wrapper {
    display: inline-flex;
    width: fit-content;
  }

  .buttons {
    display: flex;
    gap: 10px;
  }

  .dropdown {
    display: flex;
    gap: 10px;
    align-items: center;
    cursor: pointer;
    user-select: none;
    padding: 5px 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    width: fit-content;
  }

  .dropdown:hover {
    background: var(--element-bg-hover-color);
  }

  .dropdown select {
    background-color: var(--element-bg-color, rgba(255, 255, 255, 0.1));
    border: 1px solid var(--dd-select-border-color);
    border-radius: 4px;
    color: var(--text-color-primary);
    cursor: pointer;
    transition: all 0.2s ease;
    padding: 5px;
    font-size: 15px;
    min-width: 150px;
  }

  .dropdown select:hover {
    background-color: var(--element-bg-hover-color, rgba(255, 255, 255, 0.15));
    border-color: var(--dd-select-border-color, rgba(255, 255, 255, 0.2));
  }

  .dropdown select:focus {
    outline: none;
    border-color: var(--accent-color, #ffd700);
    box-shadow: 0 0 0 2px rgba(255, 215, 0, 0.2);
  }

  .dropdown select option {
    background-color: var(--bg-color-medium, #2a2a2a);
    color: var(--text-color-primary, #fff);
  }

  .btn {
    background: var(--action-secondary-bg, #444);
    color: var(--text-color-primary, #fff);
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
  }
  .btn:hover {
    background: var(--action-secondary-bg-hover, #555);
  }
  .btn.secondary {
    background: var(--action-secondary-bg, #333);
  }

  /* Modal Styles */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: var(--surface-overlay, rgba(0, 0, 0, 0.7));
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal {
    background: var(--bg-color-dark, #23282e);
    padding: 20px;
    border-radius: 8px;
    width: 400px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border-color, #444);
  }

  .modal h3 {
    margin: 0 0 20px 0;
    color: var(--text-color-primary, #fff);
    font-size: 18px;
  }

  .field {
    margin-bottom: 15px;
  }

  .field label {
    display: block;
    color: var(--text-color-muted, #aaa);
    margin-bottom: 5px;
    font-size: 13px;
  }

  .field input,
  .field select {
    width: 100%;
    background: var(--bg-color, #1a1a1a);
    border: 1px solid var(--border-color, #333);
    color: var(--text-color-primary, #fff);
    padding: 8px;
    border-radius: 4px;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 20px;
  }
</style>
