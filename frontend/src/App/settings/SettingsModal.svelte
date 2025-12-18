<script>
  import { fade, scale } from 'svelte/transition';
  import { isSettingsOpen, closeSettings } from '../lib/store/settingsModal.js';
  import TabsS from './TabsS.svelte';
  import { CloseIc } from '../lib/icons.js';
  import { t } from 'svelte-i18n';

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeSettings();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if $isSettingsOpen}
  <div class="modal-backdrop" on:click={closeSettings} transition:fade={{ duration: 200 }}>
    <div
      class="modal-content"
      on:click|stopPropagation
      transition:scale={{ duration: 200, start: 0.95 }}
    >
      <div class="modal-header">
        <h2>{$t('TITLE.settings_tooltip')}</h2>
        <button class="close-btn" on:click={closeSettings}>
          <div class="icon-wrapper">
            <CloseIc />
          </div>
        </button>
      </div>
      <div class="modal-body">
        <TabsS />
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    backdrop-filter: blur(2px);
  }

  .modal-content {
    background: var(--settings-bg-color);
    color: var(--settings-text-color);
    width: 80%;
    height: 80%;
    max-width: 1000px;
    max-height: 800px;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid var(--border-color, #444);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px 20px;
    background: var(--titlebar-bg-color);
    border-bottom: 1px solid var(--border-color, #444);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 600;
  }

  .close-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 5px;
    border-radius: 4px;
    color: var(--text-color);
    transition: background 0.2s;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .icon-wrapper {
    width: 24px;
    height: 24px;
  }

  .modal-body {
    flex: 1;
    overflow: hidden;
    position: relative;
  }
</style>
