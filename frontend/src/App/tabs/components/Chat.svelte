<script>
  // @ts-nocheck
  import { t } from 'svelte-i18n';
  import { getConfigValue, setConfigValue } from '../../lib/store/config';
  import HotkeyButton from './hotkeys/hotkey-button.svelte';
  import { isInternalChange } from '../../lib/store/internalChange';
  import Base from './Base.svelte';

  export let label = '';
  export let section = '';
  export let option = '';
  export let hotkeyOption = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = '';
  export let visible = true;
  export let isOnStart = false;
  export let width = 480;

  let textValue = '';
  let prevConfigData = null;

  async function loadValue(configAvailable) {
    if (!configAvailable) {
      textValue = '';
      return;
    }

    const value = await getConfigValue(section, option);
    textValue = value ?? '';
  }

  async function handleChange(event, configAvailable) {
    if (!configAvailable) return;

    const newValue = event.target.value;
    try {
      isInternalChange.mark();
      await setConfigValue(section, option, newValue);
      textValue = newValue;
    } catch (err) {
      console.error('Ошибка сохранения значения текстового поля:', err);
    }
  }
</script>

<Base
  {label}
  {ttKey}
  {ttImage}
  {ttPlace}
  className="chat-control-wrapper"
  style={visible ? '' : 'display: none;'}
  let:configAvailable
  let:configData
>
  {#if configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(configAvailable), '')}
  {/if}

  {#if isOnStart}
    <input
      type="text"
      class="chat-input"
      bind:value={textValue}
      on:input={(e) => handleChange(e, configAvailable)}
      disabled={!configAvailable}
      style="flex: 1;"
      placeholder={$t(label)}
    />
  {:else}
    <div class="chat-control-row">
      <input
        type="text"
        class="chat-input"
        bind:value={textValue}
        on:input={(e) => handleChange(e, configAvailable)}
        disabled={!configAvailable}
        style="flex: 1;"
        placeholder={$t(label)}
      />
      {#if hotkeyOption}
        <HotkeyButton
          {section}
          option={hotkeyOption}
          imageSrc="htk_icons/BTNSkillSlot.png"
          size={['50', '50']}
          visible={true}
        />
      {/if}
    </div>
  {/if}
</Base>

<style>
  :global(.chat-control-wrapper) {
    display: inline-flex;
    width: 100%;
  }

  .chat-control-row {
    display: flex;
    gap: 10px;
    align-items: center;
    width: 100%;
  }

  .chat-input {
    background-color: var(--chat-input-bg-color);
    border: 1px solid var(--chat-input-border-color);
    border-radius: 4px;
    color: var(--color);
    padding: 10px 14px;
    font-size: 16px;
    transition: all 0.2s ease;
    min-height: 50px;
    box-sizing: border-box;
  }

  .chat-input:hover {
    background-color: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .chat-input:focus {
    outline: none;
    border-color: #ffd700;
    box-shadow: 0 0 0 2px rgba(255, 215, 0, 0.2);
  }

  .chat-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    border-color: #666;
  }
</style>
