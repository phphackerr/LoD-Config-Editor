<script module>
  export const tabMetadata = {
    order: 3
  };
</script>

<script>
  // @ts-nocheck
  import Checkbox from './components/Checkbox.svelte';
  import Dropdown from './components/Dropdown.svelte';
  import ColorPicker from './components/ColorPicker.svelte';
  import Slider from './components/Slider.svelte';
  import { setContext } from 'svelte';

  const tabId = 'Visuals';
  setContext('tabId', tabId);

  const SECTION = 'VISUALS';

  const visualsKeys = [
    'UIManacostDisplay',
    'AlwaysDisplayRangeMarkers',
    'AlwaysDisplayNeutralMarkers',
    'AlwaysDisplayHPRegen',
    'SameSelectionCircleForEveryone',
    'CustomFPSInfo',
    'EscClearsChat',
    'EscClearsPlayersChat',
    'GoodMinimap',
    'ProperColorsForCreeps',
    'AlliesAlwaysGreen',
    'BetterFPS',
    'BetterFPS2',
    'DisableDefaultSpace',
    'DisableDefaultMouseWheel',
    'DisableDefaultTilde',
    'ShowItemsInMultiboard',
    'DisableAltTogglingHPBars',
    'IgnoreAllChat',
    'RepeatGameMessagesIntoChatLog',
    'AlwaysShowCourierButton',
    'HideMinimapSignals',
    'ColorblindMode',
    'AdvancedStatsIconDisabled',
    'CameraFlip',
    'SmoothFogReveal',
    'ClassicIngameTime',
    'DisplayAllyGoldOnSelection',
    'ShowFullWardRadiusObserver',
    'ShowFullWardRadiusSentry',
    'ShowFullWardRadiusPlague',
    'ShowFullWardRadiusNether',
    'ShowFullWardRadiusTombstone',
    'ShowFullWardRadiusIcySpirit'
  ];

  // Разделяем список пополам
  const midIndex = Math.ceil(visualsKeys.length / 2);
  const firstHalfKeys = visualsKeys.slice(0, midIndex);
  const secondHalfKeys = visualsKeys.slice(midIndex);
</script>

<div class="visuals-grid">
  <div class="options-container">
    <!-- Левая колонка -->
    <div class="options-column">
      {#each firstHalfKeys as key}
        <Checkbox
          label={`VISUALS.${key.toLowerCase()}`}
          section={SECTION}
          option={key}
          ttKey={`VISUALS.TOOLTIPS.${key.toLowerCase()}_tooltip`}
          ttPlace="right"
        />
      {/each}
    </div>

    <!-- Средняя колонка -->
    <div class="options-column">
      {#each secondHalfKeys as key}
        <Checkbox
          label={`VISUALS.${key.toLowerCase()}`}
          section={SECTION}
          option={key}
          ttKey={`VISUALS.TOOLTIPS.${key.toLowerCase()}_tooltip`}
          ttPlace="right"
        />
      {/each}
    </div>

    <!-- Правая колонка -->
    <div class="options-column">
      <Dropdown
        label="VISUALS.naturallighting"
        section={SECTION}
        option="NaturalLighting"
        options={['Dalaran', 'Lordaeron', 'Ashenvale', 'Felwood']}
        options_keys={['1', '2', '3', '4']}
      />

      <Dropdown
        label="VISUALS.weather"
        section={SECTION}
        option="Weather"
        options={['Off', 'Rain', 'Wind', 'Snow', 'Moonlight']}
        options_keys={['off', 'rain', 'wind', 'snow', 'moonlight']}
      />

      <ColorPicker label="VISUALS.watercolor" section={SECTION} option="WaterColor" />

      <Slider
        label="VISUALS.fogdensity"
        section={SECTION}
        option="FogDensity"
        min={0}
        max={255}
        step={1}
        defaultValue={90}
      />

      <Slider
        label="VISUALS.chatmessageduration"
        section={SECTION}
        option="ChatMessageDuration"
        min={1.0}
        max={60.0}
        step={0.5}
        defaultValue={8.0}
        valueType="float"
      />

      <Slider
        label="VISUALS.cameraheight"
        section={SECTION}
        option="CameraHeight"
        min={1.0}
        max={5.0}
        step={0.1}
        defaultValue={1.5}
      />

      <Slider
        label="VISUALS.cameraangle"
        section={SECTION}
        option="CameraAngle"
        min={10}
        max={90}
        step={1}
        defaultValue={56}
      />
    </div>
  </div>
</div>

<style>
  .visuals-grid {
    width: 100%;
    height: 100%;
    overflow-y: auto;
  }

  .options-container {
    height: 100%;
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
    flex: 1;
  }

  .options-column {
    flex: 1;
    height: 100%;
    min-width: 300px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    justify-content: space-between;
  }

  /* Медиа-запрос для маленьких экранов */
  @media (max-width: 768px) {
    .options-container {
      flex-direction: column;
    }

    .options-column {
      min-width: 100%;
    }
  }
</style>
