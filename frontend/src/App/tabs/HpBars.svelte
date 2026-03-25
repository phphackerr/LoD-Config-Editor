<script module>
  export const tabMetadata = {
    order: 4
  };
</script>

<script>
  import Checkbox from './components/Checkbox.svelte';
  import Dropdown from './components/Dropdown.svelte';
  import ColorPicker from './components/ColorPicker.svelte';
  import { onMount } from 'svelte';
  import { getConfigValue } from '../lib/store/config';
  import { notifyError } from '../lib/store/notifications';
  import { setContext } from 'svelte';
  import { toErrorMessage } from '../lib/store/storeUtils';
  import { t } from 'svelte-i18n';

  const tabId = 'HpBars';
  setContext('tabId', tabId);

  const SECTION = 'DOTA2HPBARSOPTIONS';
  const GAME_SECTION = 'GAMEOPTIONS';

  // Состояния
  let isHpBarsEnabled = $state(false);
  let customPreset = $state('1');
  let isCustomSelected = $state(false);
  let loadError = $state('');

  $effect(() => {
    // Явно приводим к строке, чтобы сравнение было надёжным
    isCustomSelected = String(customPreset) === '0';
  });

  onMount(async () => {
    try {
      loadError = '';
      const hpBarsConfigValue = await getConfigValue(GAME_SECTION, 'DotA2HPBars');
      isHpBarsEnabled = hpBarsConfigValue.toLowerCase() === 'true';

      const presetConfigValue = await getConfigValue(SECTION, 'CustomBarPresetNumber');
      customPreset = presetConfigValue;
    } catch (error) {
      loadError = toErrorMessage(error, $t('ERRORS.hpbars.load_settings'));
      notifyError(loadError);
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
    {#if loadError}
      <div class="load-error">{loadError}</div>
    {/if}

    <!-- Верхний ряд -->
    <div class="top-row">
      <Checkbox
        bind:checked={isHpBarsEnabled}
        label="HPBARS.DotA2HPBars"
        section={GAME_SECTION}
        option="DotA2HPBars"
      />
    </div>

    <!-- Средний ряд -->
    {#if isHpBarsEnabled}
      <div class="middle-row">
        <Dropdown
          label="HPBARS.CustomBarPresetNumber"
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
          label="HPBARS.CustomBarFixedSides"
          section={SECTION}
          option="CustomBarFixedSides"
          visible={true}
          ttKey={'HPBARS.TOOLTIPS.CustomBarFixedSides_tooltip'}
        />
      </div>
    {/if}

    <!-- Нижний ряд с цветовыми пикерами -->
    {#if isHpBarsEnabled && isCustomSelected}
      <div class="bottom-row">
        <div class="color-pickers-column">
          {#each firstHalfKeys as key}
            <ColorPicker label={`HPBARS.${key}`} section={SECTION} option={key} visible={true} />
          {/each}
        </div>
        <div class="color-pickers-column">
          {#each secondHalfKeys as key}
            <ColorPicker label={`HPBARS.${key}`} section={SECTION} option={key} visible={true} />
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
    background: var(--hpbars-bg-color, rgba(0, 0, 0, 0));
  }

  .load-error {
    max-width: 520px;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--hpbars-error-border-color, rgba(220, 38, 38, 0.4));
    background: var(--hpbars-error-bg-color, rgba(220, 38, 38, 0.15));
    color: var(--hpbars-error-text-color, #fecaca);
    font-size: 13px;
    line-height: 1.3;
    user-select: text;
    text-align: center;
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
