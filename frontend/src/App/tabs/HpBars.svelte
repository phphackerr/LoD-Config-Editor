<script module>
  // @ts-nocheck
  export const tabMetadata = {
    order: 4
  };
</script>

<script>
  //@ts-nocheck
  import Checkbox from './components/Checkbox.svelte';
  import Dropdown from './components/Dropdown.svelte';
  import ColorPicker from './components/ColorPicker.svelte';
  import { onMount } from 'svelte';
  import { getConfigValue } from '../lib/store/config';
  import { setContext } from 'svelte';

  const tabId = 'HpBars';
  setContext('tabId', tabId);

  const SECTION = 'DOTA2HPBARSOPTIONS';
  const GAME_SECTION = 'GAMEOPTIONS';

  // Состояния
  let isHpBarsEnabled = $state(false);
  let customPreset = $state('1');
  let isCustomSelected = $state(false);

  $effect(() => {
    // Явно приводим к строке, чтобы сравнение было надёжным
    isCustomSelected = String(customPreset) === '0';
  });

  onMount(async () => {
    try {
      const hpBarsConfigValue = await getConfigValue(GAME_SECTION, 'DotA2HPBars');
      isHpBarsEnabled = hpBarsConfigValue.toLowerCase() === 'true';

      const presetConfigValue = await getConfigValue(SECTION, 'CustomBarPresetNumber');
      customPreset = presetConfigValue;
    } catch (error) {
      console.error('HpBars: Ошибка загрузки конфига', error);
      isHpBarsEnabled = false;
    }
  });

  // Ключи для цветовых пикеров
  const buttonsKeys = [
    'CustomBarAllyPlayerColor_Hero',
    'CustomBarAllyPlayerColor_Unit',
    'CustomBarAllyPlayerColor_Struct',
    'CustomBarEnemyPlayerColor_Hero',
    'CustomBarEnemyPlayerColor_Unit',
    'CustomBarEnemyPlayerColor_Struct',
    'CustomBarLocalPlayerColor_Hero',
    'CustomBarLocalPlayerColor_Unit',
    'CustomBarLocalPlayerColor_Struct',
    'CustomBarNeutralPlayerColor_Unit'
  ];

  const midIndex = Math.ceil(buttonsKeys.length / 2);
  const firstHalfKeys = buttonsKeys.slice(0, midIndex);
  const secondHalfKeys = buttonsKeys.slice(midIndex);
</script>

<div class="hp-bars-grid">
  <div class="options-container">
    <!-- Верхний ряд -->
    <div class="top-row">
      <Checkbox
        bind:checked={isHpBarsEnabled}
        label="DotA2HPBars"
        section={GAME_SECTION}
        option="DotA2HPBars"
      />
    </div>

    <!-- Средний ряд -->
    {#if isHpBarsEnabled}
      <div class="middle-row">
        <Dropdown
          label="CustomBarPresetNumber"
          section={SECTION}
          option="CustomBarPresetNumber"
          options={[
            'Custom',
            'Dota Pale',
            'LoD Bright',
            'League of Legends',
            'Dusk',
            'Gray',
            'Colorblind',
            'CMYK'
          ]}
          options_keys={['0', '1', '2', '3', '4', '5', '6', '7']}
          bind:value={customPreset}
        />
        <Checkbox
          label="CustomBarFixedSides"
          section={SECTION}
          option="CustomBarFixedSides"
          visible={true}
          ttKey={'CustomBarFixedSides_tooltip'.toLowerCase()}
        />
      </div>
    {/if}

    <!-- Нижний ряд с цветовыми пикерами -->
    {#if isHpBarsEnabled && isCustomSelected}
      <div class="bottom-row">
        <div class="color-pickers-column">
          {#each firstHalfKeys as key}
            <ColorPicker label={key} section={SECTION} option={key} visible={true} />
          {/each}
        </div>
        <div class="color-pickers-column">
          {#each secondHalfKeys as key}
            <ColorPicker label={key} section={SECTION} option={key} visible={true} />
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .hp-bars-grid {
    display: grid;
    grid-template-columns: 1fr;
    flex: 1;
  }

  .options-container {
    display: flex;
    flex-direction: column;
    gap: 30px;
    align-items: center;
  }

  .top-row {
    display: flex;
    justify-content: center;
  }

  .middle-row {
    display: flex;
    gap: 20px;
    justify-content: center;
    align-items: center;
  }

  .bottom-row {
    display: flex;
    gap: 40px;
    justify-content: center;
  }

  .color-pickers-column {
    display: flex;
    flex-direction: column;
    gap: 40px;
  }

  @media (max-width: 768px) {
    .bottom-row {
      flex-direction: column;
      gap: 20px;
    }
    .color-pickers-column {
      width: 100%;
    }
  }
</style>
