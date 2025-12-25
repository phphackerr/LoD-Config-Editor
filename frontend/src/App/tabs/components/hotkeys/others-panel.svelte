<script>
  import { t } from 'svelte-i18n';
  import HotkeyButton from './hotkey-button.svelte';
  import Checkbox from '../checkbox.svelte';

  let rightPanelItems = $derived([
    {
      id: 'DisplayScoreboard',
      icon: 'BTNSkillSlot.png',
      label: 'HOTKEYS.display_scoreboard'
    },
    {
      id: 'SelectAllUnits',
      icon: 'BTNSkillSlot.png',
      label: 'HOTKEYS.select_all_units'
    },
    {
      id: 'SelectAllOtherUnits',
      icon: 'BTNSkillSlot.png',
      label: 'HOTKEYS.select_all_other_units'
    },
    {
      id: 'DisplayNeutralsSpawnAreaHotkey',
      icon: 'BTNSkillSlot.png',
      label: 'HOTKEYS.display_neutrals_spawn_area'
    },
    {
      id: 'DisplayTowerRangeHotkey',
      icon: 'BTNSkillSlot.png',
      label: 'HOTKEYS.display_tower_range'
    }
  ]);

  const utilityButtons = [
    {
      id: 'SelectBestCourier',
      icon: 'Cour.png'
    },
    {
      id: 'SelectCircleOfPower',
      icon: 'Circle.png'
    },
    {
      id: 'TeleportScrollHotkey',
      icon: 'Tp.png'
    },
    {
      id: 'ObserverWardHotkey',
      icon: 'ObserverWard.png'
    },
    {
      id: 'SentryWardHotkey',
      icon: 'SentryWard.png'
    },
    {
      id: 'DustHotkey',
      icon: 'Dust.png'
    },
    {
      id: 'SmokeHotkey',
      icon: 'Smoke.png'
    },
    {
      id: 'GlyphHotkey',
      icon: 'Glyph.png'
    },
    {
      id: 'ScanHotkey',
      icon: 'Scan.png'
    }
  ];
</script>

<div class="panel">
  <div class="header"><span>{$t('HOTKEYS.other_hotkeys')}</span></div>
  <div class="content">
    <div class="right-panel-content">
      <!-- Левая колонка с утилитами -->
      <div class="utility-column">
        {#each utilityButtons as button (button.id)}
          <HotkeyButton
            section="HOTKEYS"
            option={button.id}
            imageSrc={`htk_icons/${button.icon}`}
            size={['50', '50']}
            color="#AFFAFF"
          />
        {/each}
      </div>

      <!-- Правая колонка с кнопками и чекбоксами -->
      <div class="settings-column">
        <!-- Кнопки -->
        <div class="settings-buttons">
          {#each rightPanelItems as item (item.id)}
            <div class="right-panel-item">
              <span>{$t(item.label)}</span>
              <HotkeyButton
                section="HOTKEYS"
                option={item.id}
                imageSrc={`htk_icons/${item.icon}`}
                size={['50', '50']}
              />
            </div>
          {/each}
        </div>

        <!-- Чекбоксы -->
        <div class="checkbox-group">
          <Checkbox
            label="HOTKEYS.ShopsQWERTY"
            section="HOTKEYS"
            option="ShopsQWERTY"
            ttKey="HOTKEYS.TOOLTIPS.shops_qwer_tooltip"
            ttImage="/tooltips/shop-qwer.png"
          />
          <Checkbox
            label="HOTKEYS.UseDefaultCourierHotkeys"
            section="HOTKEYS"
            option="UseDefaultCourierHotkeys"
          />
          <Checkbox
            label="HOTKEYS.UseDefaultCircleHotkeys"
            section="HOTKEYS"
            option="UseDefaultCircleHotkeys"
          />
          <Checkbox
            label="HOTKEYS.DisableDefaultAltHotkeys"
            section="HOTKEYS"
            option="DisableDefaultAltHotkeys"
          />
          <Checkbox
            label="HOTKEYS.DisableAllDefaultHotkeys"
            section="HOTKEYS"
            option="DisableAllDefaultHotkeys"
          />
        </div>
      </div>
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
    /* overflow-y: auto; */
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
    padding: 10px;
  }

  .right-panel-content {
    display: flex;
    gap: 20px;
    justify-content: center; /* Центрируем колонки внутри right-panel-content */
    width: 100%; /* Убедимся, что контейнер занимает всю доступную ширину */
  }

  .utility-column {
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: center;
  }

  .settings-column {
    display: flex;
    flex-direction: column;
    gap: 20px; /* Увеличиваем отступ между кнопками и чекбоксами */
    flex: 1;
  }

  .settings-buttons {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .right-panel-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 5px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
  }

  .right-panel-item:hover {
    background: var(--element-bg-hover-color);
  }

  .right-panel-item span {
    flex-grow: 1; /* Позволяет тексту занимать оставшееся пространство */
  }

  .checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  @media (max-width: 1100px) {
    .content {
      transform: translateY(0%);
    }
  }
</style>
