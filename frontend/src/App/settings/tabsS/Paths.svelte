<script module>
  export const tabMetadata = {
    order: 2
  };
</script>

<script>
  import { onDestroy, onMount } from 'svelte';
  import { get } from 'svelte/store';
  import {
    appSettings,
    appSettingsState,
    runScanner,
    deletePath
  } from '../../lib/store/appSettings';
  import Radio from './components/Radio.svelte';
  import AddFolderButton from './components/AddFolderButton.svelte';
  import ScannerOverlay from './components/ScannerOverlay.svelte';
  import { t } from 'svelte-i18n';

  let gamePathOptions = [];
  let selectedGamePath = '';
  let isLoadingPaths = false;
  let localError = '';

  const unsubscribe = appSettings.subscribe((settings) => {
    if (settings.all_paths && settings.all_paths.length > 0) {
      gamePathOptions = settings.all_paths.map((path) => ({
        label: path,
        value: path
      }));
      selectedGamePath =
        settings.game_path && settings.all_paths.includes(settings.game_path)
          ? settings.game_path
          : '';
    } else {
      gamePathOptions = [];
      selectedGamePath = '';
    }
  });

  onDestroy(() => {
    unsubscribe?.();
  });

  onMount(async () => {
    const isFirstRun = get(appSettings).first_run;

    if (isFirstRun) {
      isLoadingPaths = true;
      const result = await runScanner();
      if (!result.ok) {
        localError = result.error || $t('ERRORS.paths.scan_folders');
      }
      isLoadingPaths = false;
    }
  });

  async function handleDeletePath(path) {
    localError = '';
    const result = await deletePath(path);
    if (!result.ok) {
      localError = result.error || $t('ERRORS.paths.delete_path');
    }
  }

  async function handleRunScanner() {
    isLoadingPaths = true;
    localError = '';
    const result = await runScanner();
    if (!result.ok) {
      localError = result.error || $t('ERRORS.paths.run_scanner');
    }
    isLoadingPaths = false;
  }
</script>

<ScannerOverlay show={isLoadingPaths} text={$t('SETTINGS.PATHS.scanning')} />

{#if !isLoadingPaths}
  <div class="general-settings">
    {#if localError || ($appSettingsState.error && $appSettingsState.operation === 'appSettings:run-scanner')}
      <div class="error-message">{localError || $appSettingsState.error}</div>
    {/if}

    {#if gamePathOptions.length > 0}
      <h3 class="choose-text">{$t('SETTINGS.PATHS.select_path')}</h3>
      <Radio
        options={gamePathOptions}
        name="game-path-selector"
        bind:selectedValue={selectedGamePath}
        onDelete={handleDeletePath}
      />
    {:else}
      <p class="not-found">
        {$t('SETTINGS.PATHS.paths_not_found')}
        <button class="run-scanner-button" on:click={handleRunScanner}>
          {$t('SETTINGS.PATHS.run_scanner')}
        </button>
      </p>
    {/if}

    <AddFolderButton />
  </div>
{/if}

<style>
  .general-settings {
    padding: 20px;
    overflow-x: hidden;
    align-items: center;
  }

  .choose-text {
    text-align: center;
  }

  .not-found {
    text-align: center;
    position: absolute;
    top: 45%;
    left: 55%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
  }

  .error-message {
    margin: 0 auto 14px;
    max-width: 720px;
    color: var(--status-error-text, #fecaca);
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    border-radius: 8px;
    padding: 10px 12px;
    text-align: center;
  }

  .run-scanner-button {
    padding: 10px 20px;
    background-color: var(--action-accent-bg, #e6c200);
    color: var(--surface-base, #1a1a1a);
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-size: 1em;
    transition:
      background-color 0.3s ease,
      transform 0.2s ease;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  .run-scanner-button:hover {
    background-color: var(--action-accent-bg-hover, #e6a800);
    transform: translateY(-2px);
  }
</style>
