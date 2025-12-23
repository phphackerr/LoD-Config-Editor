<script>
  import { t } from 'svelte-i18n';
  import Radio from './radio.svelte';
  import HotkeyButton from './hotkey-button.svelte';
  import { gridLayout, allButtons, skillRules, extraRules } from './buttons';

  let radioState = $state('cast'); // ← состояние радио

  const commnadRadioOptions = [
    { value: 'cast', label: 'HOTKEYS.cast' },
    { value: 'quickcast', label: 'HOTKEYS.quickcast' },
    { value: 'autocast', label: 'HOTKEYS.autocast' }
  ];

  function resolveButton(slotId, radio) {
    if (!slotId) return null;

    if (extraRules[slotId]) {
      const rule = extraRules[slotId];
      const targetId = rule.any || rule[radio];
      return allButtons[targetId] || null;
    }

    if (slotId.startsWith('SkillSlot')) {
      return skillRules[radio](slotId);
    }
    return allButtons[slotId] || null;
  }

  const commandButtonGrid = $derived(
    gridLayout.map((row) => row.map((slotId) => resolveButton(slotId, radioState)))
  );
</script>

<div class="panel">
  <div class="header"><span>{$t('HOTKEYS.command_panel')}</span></div>
  <div class="content">
    <div class="radio">
      <!-- привязываем радио к состоянию -->
      <Radio bind:group={radioState} options={commnadRadioOptions} />
    </div>
    <div class="button-groups">
      {#each commandButtonGrid as group, groupIndex}
        <div class="button-row">
          {#each group as button, buttonIndex (button ? button.id : `empty-${groupIndex}-${buttonIndex}`)}
            {#if button}
              <HotkeyButton
                section="HOTKEYS"
                option={button.id}
                imageSrc={`htk_icons/${button.icon}`}
                disabled={button.disabled || false}
                ttKey={button.tooltip}
              />
            {:else}
              <div class="empty-slot"></div>
            {/if}
          {/each}
        </div>
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
    gap: 10px;
    transform: translateY(-17%);
    padding: 10px;
  }

  .radio {
    margin-bottom: 10px;
  }

  .button-groups {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .button-row {
    display: flex;
    justify-content: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  @media (max-width: 1100px) {
    .content {
      transform: translateY(0%);
    }

    .radio {
      margin-bottom: 0;
    }
  }
</style>
