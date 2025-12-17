<script>
  import Tabs from './App/Tabs.svelte';
  import TitleBar from './App/titlebar/TitleBar.svelte';
  import SettingsModal from './App/settings/SettingsModal.svelte';
  import Watcher from './App/footer/Watcher.svelte';
  import Actions from './App/footer/Actions.svelte';
  import ConfigNotification from './App/ConfigNotification.svelte';
  import { onMount } from 'svelte';
  import { appSettings } from './App/lib/store/appSettings';
  import { openSettings } from './App/lib/store/settingsModal';

  onMount(() => {
    const unsubscribe = appSettings.subscribe((settings) => {
      if (settings.first_run) {
        openSettings();
      }
    });
    return unsubscribe;
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
  <SettingsModal />
  <ConfigNotification />
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
