<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';

  // Импортируем все компоненты из папки tabs
  const tabs = import.meta.glob('./tabsS/*.svelte', { eager: true });

  let activeTab = '';
  let tabComponents = [];

  // Загружаем табы и их метаданные
  function loadTabsAndComponents() {
    tabComponents = Object.entries(tabs)
      .map(([path, module]) => {
        const component = module.default;
        const metadata = module.tabMetadata || {};
        const id = path.split('/').pop().replace('.svelte', '');
        return {
          id,
          order: metadata.order || 999,
          component,
          icon: metadata.icon
        };
      })
      .sort((a, b) => a.order - b.order);

    if (
      tabComponents.length > 0 &&
      (!activeTab || !tabComponents.find((tab) => tab.id === activeTab))
    ) {
      activeTab = tabComponents[0].id;
    } else if (tabComponents.length === 0) {
      activeTab = '';
      console.warn('No tab components found!');
    }
  }

  onMount(() => {
    loadTabsAndComponents();
  });

  function openTab(tabId) {
    if (tabId !== activeTab) {
      activeTab = tabId;
    }
  }
</script>

<div class="tabs-container">
  <div class="tab-bar">
    {#each tabComponents as { id, icon } (id)}
      <button class="tab-btn" class:active={activeTab === id} on:click={() => openTab(id)}>
        {#if icon}
          <span class="tab-icon">{@html icon}</span>
        {/if}
        <span class="tab-label">{$t(`SETTING.${id.toUpperCase()}.${id.toLowerCase()}_tab`)}</span>
      </button>
    {/each}
  </div>

  <div class="tab-content-wrapper">
    {#each tabComponents as { id, component: TabComponent } (id)}
      <div class="tab-content" style="display: {activeTab === id ? 'block' : 'none'};">
        <TabComponent />
      </div>
    {/each}
  </div>
</div>

<style>
  .tabs-container {
    display: flex;
    height: 100%;
  }

  .tab-bar {
    display: flex;
    flex-direction: column;
    background: var(--settings-tabbar-background);
    width: 15%;
    min-width: 150px;
    padding-top: 10px;
    box-shadow: var(--settings-tabbar-box-shadow);
    z-index: 1;
  }
  .tab-btn {
    background: var(--settings-tab-bg-color);
    border: var(--settings-tab-border);
    color: var(--settings-tab-text-color);
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    padding: 10px 20px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    transition: var(--settings-tab-transition);
    width: 100%;
    text-align: left;
  }
  .tab-btn.active {
    background: var(--settings-tab-bg-color-active);
    color: var(--settings-tab-text-color-active);
    border-right: var(--settings-tab-right-border);
    border-bottom: none;
  }
  .tab-icon {
    margin-right: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
  }
  .tab-label {
    white-space: nowrap;
  }
  .tab-content-wrapper {
    flex-grow: 1;
    overflow-y: auto;
  }
</style>
