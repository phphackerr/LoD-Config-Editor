<script>
  import { Window } from '@wailsio/runtime';
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { appSettings } from '../../../../lib/store/appSettings';
  import {
    changeLanguage,
    getAvailableLanguages,
    getTranslationsForLanguage,
    saveTranslations
  } from '../../../../lib/store/i18n';
  import { notifyError, notifySuccess } from '../../../../lib/store/notifications';
  import { tt } from '../../../../lib/tooltip';

  let languages = [];
  let currentLang = 'en';

  let transSource = 'en';
  let transTarget = '';
  let sourceValues = {};
  let targetValues = {};
  let transKeys = [];
  let filterText = '';
  let loading = false;
  let translatorError = '';
  let translatorStatus = '';

  onMount(async () => {
    currentLang = $appSettings.language || 'en';
    transSource = currentLang;
    await refreshLanguages();
  });

  async function refreshLanguages() {
    loading = true;
    translatorError = '';

    const result = await getAvailableLanguages();
    if (!result.ok) {
      loading = false;
      translatorError = result.error || $t('ERRORS.i18n.load_languages');
      return;
    }

    languages = Array.isArray(result.data) ? result.data : [];
    const availableCodes = languages.map((lang) => lang.code);

    if (!availableCodes.includes(transSource)) {
      transSource = availableCodes[0] || 'en';
    }

    if (!availableCodes.includes(transTarget)) {
      transTarget =
        availableCodes.find((code) => code !== transSource) ||
        availableCodes[0] ||
        transSource ||
        'en';
    }

    await loadTranslatorData();
    loading = false;
  }

  async function loadTranslatorData() {
    if (!transSource || !transTarget) return;

    loading = true;
    translatorError = '';
    translatorStatus = '';

    const [sourceResult, targetResult] = await Promise.all([
      getTranslationsForLanguage(transSource),
      getTranslationsForLanguage(transTarget)
    ]);

    if (!sourceResult.ok || !targetResult.ok) {
      loading = false;
      translatorError =
        sourceResult.error || targetResult.error || $t('ERRORS.i18n.load_selected_translations');
      return;
    }

    sourceValues = sourceResult.data || {};
    targetValues = targetResult.data || {};
    const keys = new Set([...Object.keys(sourceValues), ...Object.keys(targetValues)]);
    transKeys = Array.from(keys).sort();
    loading = false;
  }

  async function saveTranslation() {
    translatorError = '';
    translatorStatus = '';

    const result = await saveTranslations(transTarget, targetValues);
    if (!result.ok) {
      translatorError = result.error || $t('ERRORS.i18n.save_translations');
      notifyError(translatorError);
      return;
    }

    translatorStatus = `${$t('SETTINGS.LANGUAGE_EDITOR.saved_prefix')} ${transTarget.toUpperCase()} ${$t('SETTINGS.LANGUAGE_EDITOR.saved_suffix')}`;
    notifySuccess(translatorStatus);

    if (transTarget === currentLang) {
      const applyResult = await changeLanguage(currentLang);
      if (!applyResult.ok) {
        translatorError = applyResult.error || $t('ERRORS.i18n.saved_but_reload_failed');
        notifyError(translatorError);
      }
    }
  }

  function resize(node, _value) {
    const handleInput = () => {
      node.style.height = 'auto';
      node.style.height = node.scrollHeight + 'px';
    };

    node.addEventListener('input', handleInput);
    setTimeout(handleInput, 0);

    return {
      update() {
        setTimeout(handleInput, 0);
      },
      destroy() {
        node.removeEventListener('input', handleInput);
      }
    };
  }

  $: filteredKeys = transKeys.filter((key) => {
    const q = filterText.toLowerCase();
    return (
      key.toLowerCase().includes(q) ||
      (sourceValues[key] && sourceValues[key].toLowerCase().includes(q)) ||
      (targetValues[key] && targetValues[key].toLowerCase().includes(q))
    );
  });
</script>

<div class="editor-root">
  <header class="header">
    <div class="title">{$t('SETTINGS.LANGUAGE_EDITOR.title')}</div>

    <div class="controls">
      <select bind:value={transSource} on:change={loadTranslatorData} disabled={loading}>
        {#each languages as lang}
          <option value={lang.code}>{lang.name} ({lang.code})</option>
        {/each}
      </select>

      <span class="arrow">→</span>

      <select bind:value={transTarget} on:change={loadTranslatorData} disabled={loading}>
        {#each languages as lang}
          <option value={lang.code}>{lang.name} ({lang.code})</option>
        {/each}
      </select>

      <input class="search" bind:value={filterText} placeholder={$t('COMMON.search')} />

      <button class="btn primary" on:click={saveTranslation} disabled={loading}
        >{$t('COMMON.save')}</button
      >
      <button class="btn close" on:click={Window.Hide}>{$t('COMMON.close')}</button>
    </div>
  </header>

  {#if loading}
    <div class="status">{$t('COMMON.loading')}...</div>
  {/if}
  {#if translatorError}
    <div class="status status-error">{translatorError}</div>
  {/if}
  {#if translatorStatus}
    <div class="status status-success">{translatorStatus}</div>
  {/if}

  <div class="grid-header">
    <div class="col">{$t('SETTINGS.LANGUAGE.key')}</div>
    <div class="col">{transSource}</div>
    <div class="col">{transTarget}</div>
  </div>

  <div class="grid-body">
    {#each filteredKeys as key}
      <div class="grid-row">
        <div class="col key" use:tt={{ content: key }}>{key}</div>
        <div class="col source">{sourceValues[key] || ''}</div>
        <div class="col target">
          <textarea use:resize={targetValues[key]} bind:value={targetValues[key]} rows="1"
          ></textarea>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .editor-root {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--app-bg, rgb(47, 47, 47));
    color: var(--app-text, rgb(246, 246, 246));
  }

  .header {
    --wails-draggable: drag;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .title {
    font-size: 15px;
    font-weight: 600;
  }

  .controls {
    --wails-draggable: no-drag;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .controls select,
  .search {
    background: var(--element-bg-color, rgba(255, 255, 255, 0.1));
    border: 1px solid var(--dd-select-border-color, rgba(255, 255, 255, 0.2));
    border-radius: 6px;
    color: var(--text-color-primary, #fff);
    padding: 6px 8px;
    min-width: 180px;
  }

  .search {
    min-width: 220px;
  }

  .arrow {
    opacity: 0.8;
  }

  .btn {
    border: none;
    border-radius: 6px;
    padding: 7px 12px;
    cursor: pointer;
    color: var(--text-color-primary, #fff);
  }

  .btn.primary {
    background: var(--action-primary-bg, #3ba475);
  }

  .btn.close {
    background: var(--action-danger-bg, #ca3333);
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .status {
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color, #333);
    font-size: 12px;
  }

  .status-error {
    color: var(--status-error-text, #fecaca);
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
  }

  .status-success {
    color: var(--status-success-text, #9be5be);
    background: var(--status-success-bg, rgba(36, 147, 79, 0.18));
  }

  .grid-header {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    background: var(--bg-color, #1a1a1a);
    padding: 10px;
    font-weight: 600;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .grid-body {
    flex: 1;
    overflow-y: auto;
    background: var(--bg-color-dark, #151515);
  }

  .grid-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    border-bottom: 1px solid var(--border-color, #222);
  }

  .grid-row:hover {
    background: var(--bg-color-medium, #1e1e1e);
  }

  .col {
    padding: 10px;
    color: var(--text-color-secondary, #ccc);
    display: flex;
    align-items: center;
    word-break: break-word;
  }

  .col.key {
    color: var(--primary-color, #3ba475);
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  textarea {
    width: 100%;
    resize: vertical;
    min-height: 30px;
    overflow: hidden;
    box-sizing: border-box;
    background: var(--bg-color, #1a1a1a);
    border: 1px solid var(--border-color, #333);
    color: var(--text-color-secondary, #ccc);
    padding: 6px 10px;
    border-radius: 4px;
  }
</style>
