<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { activeTab } from './lib/store/search';

  // Импортируем все компоненты из папки tabs
  const tabs = import.meta.glob('./tabs/*.svelte', { eager: true });

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

    if (tabComponents.length > 0) {
      activeTab.set(tabComponents[0].id);
    } else {
      activeTab.set('');
      console.warn('No tab components found!');
    }
  }

  onMount(() => {
    loadTabsAndComponents();
  });

  function openTab(tabId) {
    activeTab.set(tabId);
  }
</script>

<div class="tab-bar">
  {#each tabComponents as { id, icon } (id)}
    <button class="tab-btn" class:active={$activeTab === id} on:click={() => openTab(id)}>
      {#if icon}
        <span class="tab-icon">{@html icon}</span>
      {/if}
      <span class="tab-label">{$t(`${id.toUpperCase()}.${id.toLowerCase()}_tab`)}</span>
    </button>
  {/each}
</div>

{#each tabComponents as { id, component: TabComponent } (id)}
  <div class="tab-content" style="display: {$activeTab === id ? 'block' : 'none'};">
    <TabComponent />
  </div>
{/each}

<style>
  .tab-bar {
    display: flex;
    background: var(--tab-background, #23282e);
    height: 50px;
    align-items: stretch;
  }

  .tab-btn {
    flex: 1;
    background: none;
    border: none;
    color: var(--tab-color, #ccc);
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    padding: 0 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
      background 0.2s,
      color 0.2s;
  }

  .tab-btn:hover {
    background: var(--tab-hover-background, #181c20);
  }

  .tab-btn:focus-visible {
    background: var(--tab-hover-background, #181c20);
  }

  .tab-btn.active {
    background: var(--tab-active-background, #181c20);
    color: var(--tab-active-color, #3ba475);
    border-bottom: 3px solid #3ba475;
  }

  .tab-btn.active:focus-visible {
    border-bottom: 3px solid #3ba475;
  }

  .tab-icon {
    margin-right: 8px;
    display: flex;
    align-items: center;
  }

  .tab-label {
    white-space: nowrap;
  }

  .tab-content {
    height: calc(100% - 50px - 30px);
    padding: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
  }
</style>
