<script>
  import { onMount } from 'svelte';
  import { slide } from 'svelte/transition';
  import { updaterStore, checkForUpdates, doUpdate, updateComponent } from '../lib/store/updater';
  import { t } from 'svelte-i18n';

  onMount(() => {
    checkForUpdates();
  });

  function handleUpdate() {
    if ($updaterStore.version) {
      doUpdate($updaterStore.version);
    }
  }

  function handleComponentUpdate(component) {
    updateComponent(component);
  }
</script>

{#if $updaterStore.available || $updaterStore.componentUpdates.length > 0}
  <div class="updater-notification" transition:slide>
    <div class="content">
      {#if $updaterStore.available}
        <div class="info">
          <span class="title"
            >{$t('update_available', { default: 'Update Available' })}: {$updaterStore.version}</span
          >
          {#if $updaterStore.downloading}
            <div class="progress-bar">
              <div class="fill" style="width: {$updaterStore.progress}%"></div>
            </div>
            <span class="status">Downloading... {Math.round($updaterStore.progress)}%</span>
          {:else}
            <button class="update-btn" on:click={handleUpdate}>
              {$t('update_now', { default: 'Update Now' })}
            </button>
          {/if}
        </div>
      {/if}

      {#if $updaterStore.componentUpdates.length > 0}
        <div class="components-list">
          <span class="title">{$t('component_updates', { default: 'Component Updates' })}</span>
          {#each $updaterStore.componentUpdates as comp}
            <div class="component-item">
              <div class="comp-info">
                <span class="comp-name">{comp.type}: {comp.name} ({comp.version})</span>
                {#if comp.changelog}
                  <span class="comp-changelog">{comp.changelog}</span>
                {/if}
              </div>
              <button class="btn-small" on:click={() => handleComponentUpdate(comp)}>
                {$t('update', { default: 'Update' })}
              </button>
            </div>
          {/each}
        </div>
      {/if}

      {#if $updaterStore.error}
        <div class="error">{$updaterStore.error}</div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .updater-notification {
    position: fixed;
    bottom: 20px;
    right: 20px;
    background: rgba(30, 30, 30, 0.95);
    border: 1px solid #ffd700;
    border-radius: 8px;
    padding: 15px;
    z-index: 1000;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(10px);
    max-width: 300px;
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .title {
    color: #ffd700;
    font-weight: bold;
    display: block;
    margin-bottom: 5px;
  }

  .update-btn {
    background: #ffd700;
    color: #000;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
    transition: all 0.2s;
    width: 100%;
  }

  .update-btn:hover {
    background: #ffed4a;
    transform: translateY(-1px);
  }

  .progress-bar {
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 5px;
  }

  .fill {
    height: 100%;
    background: #ffd700;
    transition: width 0.3s ease;
  }

  .status {
    font-size: 12px;
    color: #aaa;
  }

  .error {
    color: #ff4444;
    font-size: 12px;
    margin-top: 5px;
  }

  .components-list {
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    padding-top: 10px;
    margin-top: 5px;
  }

  .component-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    margin-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    padding-bottom: 5px;
  }

  .comp-info {
    display: flex;
    flex-direction: column;
  }

  .comp-name {
    font-weight: bold;
  }

  .comp-changelog {
    font-size: 10px;
    color: #aaa;
    margin-top: 2px;
    font-style: italic;
  }

  .btn-small {
    background: #444;
    color: #fff;
    border: 1px solid #666;
    padding: 2px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 10px;
  }

  .btn-small:hover {
    background: #555;
  }
</style>
