<script>
  import { isConfigAvailable } from './lib/store/config.js';
  import { t } from 'svelte-i18n';
  import { openSettings } from './lib/store/settingsModal.js';
  import { appSettings } from './lib/store/appSettings.js';

  let _isConfigAvailable = $state(false);
  let showNotification = $state(false);

  // следим за изменением game_path
  $effect(() => {
    const path = $appSettings.game_path; // Зависимость для реактивности
    (async () => {
      if (!path) {
        _isConfigAvailable = false;
      } else {
        _isConfigAvailable = await isConfigAvailable();
      }
      showNotification = !_isConfigAvailable;
    })();
  });
</script>

{#if showNotification}
  <div class="notification">
    <div class="notification-content">
      <span>{$t('SETTINGS.PATHS.paths_not_found')}</span>
      <button onclick={openSettings}>{$t('SETTINGS.PATHS.select_path')}</button>
    </div>
  </div>
{/if}

<style>
  .notification {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    border-radius: 4px;
    padding: 12px 20px;
    z-index: 1000;
    backdrop-filter: blur(5px);
  }

  .notification-content {
    display: flex;
    align-items: center;
    gap: 16px;
    color: var(--status-error-text, #fecaca);
  }

  button {
    background: var(--action-primary-bg, #3ba475);
    border: 1px solid var(--status-success-border, rgba(36, 147, 79, 0.38));
    color: var(--text-color-primary, #fff);
    padding: 6px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
  }

  button:hover {
    background: var(--action-primary-bg-hover, #2e8b57);
  }
</style>
