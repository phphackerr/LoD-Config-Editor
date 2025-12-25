<script>
  import Tabs from './App/Tabs.svelte';
  import TitleBar from './App/titlebar/TitleBar.svelte';
  import SettingsModal from './App/settings/SettingsModal.svelte';
  import Watcher from './App/footer/Watcher.svelte';
  import Actions from './App/footer/Actions.svelte';
  import ConfigNotification from './App/ConfigNotification.svelte';
  import UpdaterNotification from './App/UpdaterNotification.svelte';
  import { onMount } from 'svelte';
  import { appSettings } from './App/lib/store/appSettings';
  import { openSettings, isSettingsOpen } from './App/lib/store/settingsModal';

  import { OpenDevTools } from '/bindings/lce/backend/utils/utils';

  onMount(() => {
    const handleKeydown = (e) => {
      if (e.ctrlKey && e.shiftKey && e.key === 'F12') {
        OpenDevTools();
      }
    };

    window.addEventListener('keydown', handleKeydown);

    const unsubscribe = appSettings.subscribe((settings) => {
      if (settings.first_run) {
        openSettings('Paths');
      }
    });

    return () => {
      window.removeEventListener('keydown', handleKeydown);
      unsubscribe();
    };
  });
</script>

<div class="main">
  <TitleBar />
  <div class="tabs">
    <Tabs />
  </div>
  <div class="footer">
    <Watcher />
    <Actions />
  </div>
  <SettingsModal bind:isOpen={$isSettingsOpen} />
  <ConfigNotification />
  <UpdaterNotification />
</div>

<style>
  .main {
    height: 100%;
    background: var(--background-color);
    color: var(--text-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .tabs {
    flex-grow: 1;
    overflow: hidden;
  }
</style>
