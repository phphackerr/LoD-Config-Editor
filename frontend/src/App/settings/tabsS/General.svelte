<script module>
  export const tabMetadata = {
    order: 1
  };
</script>

<script>
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { appSettings, runScanner, deletePath } from '../../lib/store/appSettings';
  import Radio from './components/Radio.svelte';
  import AddFolderButton from './components/AddFolderButton.svelte';
  import ScannerOverlay from './components/ScannerOverlay.svelte';
  import { t } from 'svelte-i18n';

  let gamePathOptions = [];
  let selectedGamePath = '';
  let isLoadingPaths = false;

  // подписка на изменения стора
  appSettings.subscribe((settings) => {
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

  onMount(async () => {
    const isFirstRun = get(appSettings).first_run;

    if (isFirstRun) {
      isLoadingPaths = true;
      try {
        await runScanner();
      } finally {
        isLoadingPaths = false;
      }
    }
  });

  async function handleDeletePath(path) {
    await deletePath(path);
  }

  async function handleRunScanner() {
    isLoadingPaths = true;
    await runScanner();
    isLoadingPaths = false;
  }
</script>

<ScannerOverlay show={isLoadingPaths} text={$t('SETTING.GENERAL.scanning')} />

{#if !isLoadingPaths}
  <div class="general-settings">
    {#if gamePathOptions.length > 0}
      <h3 class="choose-text">{$t('SETTING.GENERAL.select_path')}</h3>
      <Radio
        options={gamePathOptions}
        name="game-path-selector"
        bind:selectedValue={selectedGamePath}
        onDelete={handleDeletePath}
      />
    {:else}
      <p class="not-found">
        {$t('SETTING.GENERAL.paths_not_found')}
        <button class="run-scanner-button" on:click={handleRunScanner}>
          {$t('SETTING.GENERAL.run_scanner')}
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

  .run-scanner-button {
    padding: 10px 20px;
    background-color: #007bff;
    color: white;
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
    background-color: #0056b3;
    transform: translateY(-2px);
  }
</style>
