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
      console.log('Config check:', _isConfigAvailable, 'Path:', path);
      showNotification = !_isConfigAvailable;
    })();
  });
</script>

{#if showNotification}
  <div class="notification">
    <div class="notification-content">
      <span>{$t('config_not_found')}</span>
      <button onclick={openSettings}>{$t('SETTING.GENERAL.select_path')}</button>
    </div>
  </div>
{/if}

<style>
  .notification {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(255, 0, 0, 0.5);
    border: 1px solid rgba(255, 0, 0, 0.3);
    border-radius: 4px;
    padding: 12px 20px;
    z-index: 1000;
    backdrop-filter: blur(5px);
  }

  .notification-content {
    display: flex;
    align-items: center;
    gap: 16px;
    color: #fff;
  }

  button {
    background: rgba(16, 245, 27, 0.212);
    border: 1px solid rgba(255, 255, 255, 0.9);
    color: #fff;
    padding: 6px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
  }

  button:hover {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
