<script>
  import { t } from 'svelte-i18n';
  import { getConfigValue } from '../../lib/store/config';
  import HotkeyButton from './hotkeys/hotkey-button.svelte';
  import { persistControlValue } from './lib/controlPipeline';
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
  let lastSavedValue = '';
  let prevConfigData = null;
  $: inputStyle = `flex: 1; max-width: ${width}px;`;

  async function loadValue(configAvailable) {
    if (!configAvailable) {
      textValue = '';
      return;
    }

    const value = String((await getConfigValue(section, option)) || '');
    textValue = value;
    lastSavedValue = value;
  }

  async function handleChange(event, configAvailable) {
    if (!configAvailable) return;

    const newValue = String(event?.target?.value || '');

    const result = await persistControlValue(
      section,
      option,
      newValue,
      $t('ERRORS.controls.save_chat')
    );
    if (!result.ok) {
      textValue = lastSavedValue;
      return;
    }

    textValue = newValue;
    lastSavedValue = newValue;
  }
</script>

<Base
  {label}
  {section}
  {option}
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
      style={inputStyle}
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
        style={inputStyle}
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
    color: var(--text-color-primary);
    padding: 10px 14px;
    font-size: 16px;
    transition: all 0.2s ease;
    min-height: 50px;
    box-sizing: border-box;
  }

  .chat-input:hover {
    background-color: var(--chat-input-hover-bg-color, rgba(255, 255, 255, 0.15));
    border-color: var(--chat-input-hover-border-color, rgba(255, 255, 255, 0.3));
  }

  .chat-input:focus {
    outline: none;
    border-color: var(--chat-input-focus-border-color, #ffd700);
    box-shadow: 0 0 0 2px var(--chat-input-focus-ring-color, rgba(255, 215, 0, 0.2));
  }

  .chat-input:disabled {
    opacity: var(--control-disabled-opacity, 0.5);
    cursor: not-allowed;
    border-color: var(--control-disabled-border-color, #666);
  }
</style>
