<script module>
  export const tabMetadata = {
    order: 2
  };
</script>

<script>
  import Checkbox from './components/Checkbox.svelte';
  import Dropdown from './components/Dropdown.svelte';
  import ColorPicker from './components/ColorPicker.svelte';
  import { setContext } from 'svelte';

  const tabId = 'GameOptions'; // <-- ВАЖНО: Замените это на ID для каждого файла!
  setContext('tabId', tabId);

  const SECTION = 'GAMEOPTIONS';

  const gameoptionsKeys = [
    'WideScreen',
    'AutoFPSLimit',
    'LockMouseAtWindow',
    'AutoselectHero',
    'TeleportationCanOnlyBeStopped',
    'CloseWC3EveryGame',
    'AutoattackEnabledHeroes',
    'AutoattackEnabledUnits',
    'AutoattackDisabledByStopOnlyHeroes',
    'AutoattackDisabledByStopOnlyUnits',
    'SmartAttackEnabled',
    'RightClickDeny',
    'SelectionHelperEnabled',
    'DoubleClickHelperDisabled',
    'IDontWantToVisitSite',
    'IAmShy',
    'DisplayFPSCounter'
  ];

  // Разделяем список пополам
  const midIndex = Math.ceil(gameoptionsKeys.length / 2);
  const firstHalfKeys = gameoptionsKeys.slice(0, midIndex);
  const secondHalfKeys = gameoptionsKeys.slice(midIndex);
</script>

<div class="options-container">
  <!-- Левая колонка -->
  <div class="options-column">
    <div class="centered-content-wrapper">
      {#each firstHalfKeys as key}
        <Checkbox
          label={`GAMEOPTIONS.${key.toLowerCase()}`}
          section={SECTION}
          option={key}
          ttKey={`GAMEOPTIONS.TOOLTIPS.${key.toLowerCase()}_tooltip`}
          ttPlace="left"
        />
      {/each}

      <Dropdown
        label="GAMEOPTIONS.announcer"
        section="HEROSETS"
        option="Announcer"
        options={['Default', 'Sexy', 'Anime']}
        options_keys={['default', 'sexy', 'anime']}
      />

      <Dropdown
        label="GAMEOPTIONS.blinkeffect"
        section="HEROSETS"
        option="BlinkEffect"
        options={['Default', 'SF', 'Tinker']}
        options_keys={['default', 'sf', 'tinker']}
      />

      <Dropdown
        label="GAMEOPTIONS.membershipeffect"
        section="HEROSETS"
        option="MembershipEffect"
        options={['Off', 'Amethyst', 'Silver', 'Gold']}
        options_keys={['off', 'amethyst', 'silver', 'gold']}
        ttKey={'GAMEOPTIONS.TOOLTIPS.membershipeffect_tooltip'}
      />
    </div>
  </div>

  <!-- Правая колонка -->
  <div class="options-column">
    <div class="centered-content-wrapper">
      {#each secondHalfKeys as key}
        <Checkbox
          label={`GAMEOPTIONS.${key.toLowerCase()}`}
          section={SECTION}
          option={key}
          ttKey={`GAMEOPTIONS.TOOLTIPS.${key.toLowerCase()}_tooltip`}
          ttPlace="left"
          reverted={key === 'DoubleClickHelperDisabled'}
        />
      {/each}

      <ColorPicker
        label="GAMEOPTIONS.customchatmessagescolor"
        section="HEROSETS"
        option="CustomChatMessagesColor"
      />
    </div>
  </div>
</div>

<style>
  .options-container {
    width: 100%;
    height: 100%;
    display: flex;
    gap: 20px;
    overflow-y: auto;
    background: var(--gameoptions-bg-color, rgba(0, 0, 0, 0));
  }

  .options-column {
    flex: 1;
    min-width: 300px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    justify-content: space-between;
    align-items: center;
  }

  .centered-content-wrapper {
    width: fit-content;
    height: 100%;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: inherit;
    justify-content: inherit;
    align-items: flex-start;
  }

  @media (max-width: 768px) {
    .options-container {
      flex-direction: column;
      gap: 10px;
    }

    .options-column {
      min-width: 100%;
    }

    .options-column .centered-content-wrapper {
      margin: 0 auto;
      align-items: flex-start;
    }
  }
</style>
