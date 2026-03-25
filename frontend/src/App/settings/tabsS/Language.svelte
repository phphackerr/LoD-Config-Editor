<script module>
  export const tabMetadata = {
    order: 3
  };
</script>

<script>
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { appSettings, appSettingsState, updateLanguage } from '../../lib/store/appSettings';
  import {
    changeLanguage,
    getAvailableLanguages,
    createNewLanguage,
    i18nState
  } from '../../lib/store/i18n';
  import { OpenLanguageEditor } from '/bindings/lce/backend/windows/windowservice';
  import { OpenLocalesFolder } from '/bindings/lce/backend/utils/utils';
  import { toErrorMessage } from '../../lib/store/storeUtils';

  let languages = [];
  let currentLang = 'en';
  let localError = '';
  let localStatus = '';

  let showAddModal = false;
  let newLangData = { code: '', name: '', author: '' };

  onMount(async () => {
    await refreshLanguages();
    currentLang = $appSettings.language;
  });

  async function refreshLanguages() {
    const result = await getAvailableLanguages();
    if (!result.ok) {
      localError = result.error || $t('ERRORS.i18n.load_languages');
      languages = [];
      return;
    }

    localError = '';
    languages = Array.isArray(result.data) ? result.data : [];
  }

  async function handleLanguageChange(event) {
    const newLang = event.target.value;
    const previousLang = currentLang || $appSettings.language || 'en';

    currentLang = newLang;
    localError = '';
    localStatus = '';

    const settingsResult = await updateLanguage(newLang);
    if (!settingsResult.ok) {
      currentLang = previousLang;
      localError = settingsResult.error || $t('ERRORS.language.save_setting');
      return;
    }

    const languageResult = await changeLanguage(newLang);
    if (!languageResult.ok) {
      await updateLanguage(previousLang);
      currentLang = previousLang;
      localError = languageResult.error || $t('ERRORS.language.apply');
      return;
    }

    localStatus = `${newLang.toUpperCase()} ${$t('SETTINGS.LANGUAGE.status_selected')}`;
  }

  async function createLanguage() {
    if (!newLangData.code || !newLangData.name) return;

    localError = '';
    const createdCode = String(newLangData.code || '')
      .trim()
      .toUpperCase();
    const result = await createNewLanguage(newLangData.code, newLangData.name, newLangData.author);
    if (!result.ok) {
      localError = result.error || $t('ERRORS.i18n.create_language');
      return;
    }

    await refreshLanguages();
    showAddModal = false;
    newLangData = { code: '', name: '', author: '' };
    localStatus = `${createdCode} ${$t('SETTINGS.LANGUAGE.status_created')}`;
  }

  async function openLanguageEditor() {
    localError = '';
    try {
      await OpenLanguageEditor();
    } catch (error) {
      localError = toErrorMessage(error, $t('ERRORS.language.open_editor'));
    }
  }

  async function openLocalesFolder() {
    localError = '';
    try {
      await OpenLocalesFolder();
    } catch (error) {
      localError = toErrorMessage(error, $t('ERRORS.language.open_folder'));
    }
  }
</script>

<div class="content">
  {#if localError}
    <div class="error-message">{localError}</div>
  {/if}
  {#if localStatus}
    <div class="success-message">{localStatus}</div>
  {/if}
  {#if $i18nState.error}
    <div class="error-message">{$i18nState.error}</div>
  {/if}

  <div class="setting-row">
    <div class="label">{$t('SETTINGS.LANGUAGE.lang')}</div>
    <div class="dropdown-wrapper">
      <label class="dropdown">
        <select
          value={currentLang}
          on:change={handleLanguageChange}
          disabled={$i18nState.loading || $appSettingsState.loading}
        >
          {#each languages as lang}
            <option value={lang.code}>{lang.name}</option>
          {/each}
        </select>
      </label>
    </div>
  </div>

  <div class="setting-row">
    <div class="label">{$t('SETTINGS.LANGUAGE.tools')}</div>
    <div class="buttons">
      <button class="btn" on:click={() => (showAddModal = true)}
        >{$t('SETTINGS.LANGUAGE.add_language')}</button
      >
      <button class="btn" on:click={openLanguageEditor}>{$t('SETTINGS.LANGUAGE.translator')}</button
      >
      <button class="btn" on:click={openLocalesFolder}>{$t('SETTINGS.LANGUAGE.open_folder')}</button
      >
    </div>
  </div>
</div>

{#if showAddModal}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal-backdrop" on:click|self={() => (showAddModal = false)}>
    <div class="modal" role="dialog" aria-modal="true">
      <h3>{$t('SETTINGS.LANGUAGE.add_new_language')}</h3>
      <div class="field">
        <label>
          {$t('SETTINGS.LANGUAGE.code_placeholder')}
          <input bind:value={newLangData.code} placeholder={$t('SETTINGS.LANGUAGE.example_code')} />
        </label>
      </div>
      <div class="field">
        <label>
          {$t('SETTINGS.LANGUAGE.name_placeholder')}
          <input bind:value={newLangData.name} placeholder={$t('SETTINGS.LANGUAGE.example_name')} />
        </label>
      </div>
      <div class="field">
        <label>
          {$t('SETTINGS.LANGUAGE.author_placeholder')}
          <input
            bind:value={newLangData.author}
            placeholder={$t('SETTINGS.LANGUAGE.example_author')}
          />
        </label>
      </div>
      <div class="actions">
        <button class="btn" on:click={createLanguage}>{$t('SETTINGS.LANGUAGE.create')}</button>
        <button class="btn secondary" on:click={() => (showAddModal = false)}
          >{$t('SETTINGS.LANGUAGE.cancel')}</button
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

  .setting-row {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 20px;
  }

  .error-message,
  .success-message {
    margin-bottom: 12px;
    padding: 8px 10px;
    border-radius: 8px;
    font-size: 13px;
  }

  .error-message {
    color: var(--status-error-text, #fecaca);
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
  }

  .success-message {
    color: var(--status-success-text, #9be5be);
    border: 1px solid var(--status-success-border, rgba(36, 147, 79, 0.38));
    background: var(--status-success-bg, rgba(36, 147, 79, 0.18));
  }

  .label {
    font-size: 16px;
    min-width: 100px;
  }

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
    display: block;
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

  .field input {
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
