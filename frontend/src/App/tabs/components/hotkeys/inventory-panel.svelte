<script>
  import { t } from 'svelte-i18n';
  import HotkeyButton from './hotkey-button.svelte';
  import Radio from './radio.svelte';

  let radioState = $state('cast');

  const radioOptions = [
    { value: 'cast', label: 'HOTKEYS.inventory_cast' },
    { value: 'quickcast', label: 'HOTKEYS.inventory_quickcast' }
  ];

  const inventoryButtons = [
    {
      id: 'ItemSlot1',
      quickcast: 'QuickCastInventorySlot1',
      icon: 'BTNItemSlot.png'
    },
    {
      id: 'ItemSlot2',
      quickcast: 'QuickCastInventorySlot2',
      icon: 'BTNItemSlot.png'
    },
    {
      id: 'ItemSlot3',
      quickcast: 'QuickCastInventorySlot3',
      icon: 'BTNItemSlot.png'
    },
    {
      id: 'ItemSlot4',
      quickcast: 'QuickCastInventorySlot4',
      icon: 'BTNItemSlot.png'
    },
    {
      id: 'ItemSlot5',
      quickcast: 'QuickCastInventorySlot5',
      icon: 'BTNItemSlot.png'
    },
    {
      id: 'ItemSlot6',
      quickcast: 'QuickCastInventorySlot6',
      icon: 'BTNItemSlot.png'
    }
  ];

  // Получаем итоговые кнопки в зависимости от режима
  const renderedInventoryButtons = $derived(
    inventoryButtons.map((btn) => ({
      id: radioState === 'cast' ? btn.id : btn.quickcast,
      icon: btn.icon
    }))
  );
</script>

<div class="panel">
  <div class="header"><span>{$t('HOTKEYS.inventory_panel')}</span></div>
  <div class="content">
    <div class="radio">
      <Radio bind:group={radioState} options={radioOptions} />
    </div>
    <div class="inventory-grid">
      {#each renderedInventoryButtons as button (button.id)}
        <HotkeyButton section="HOTKEYS" option={button.id} imageSrc={`htk_icons/${button.icon}`} />
      {/each}
    </div>
  </div>
</div>

<style>
  .panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: rgba(30, 30, 30, 0.5);
    border: 1px solid transparent;
    border-radius: 10px;
    box-sizing: border-box;
  }

  .header {
    text-align: center;
    padding: 10px;
  }

  .header span {
    font-size: 24px;
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 10px;
    transform: translateY(-12%);
    padding: 10px;
  }

  .radio {
    margin-bottom: 10px;
  }

  .inventory-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr); /* 2 колонки */
    grid-template-rows: repeat(3, 1fr); /* 3 строки */
    gap: 10px; /* почти вплотную, можно 0 */
  }

  @media (max-width: 1100px) {
    .content {
      transform: translateY(0%);
    }
  }
</style>
