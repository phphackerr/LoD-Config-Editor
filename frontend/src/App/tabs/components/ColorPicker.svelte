<script>
  // @ts-nocheck
  import { t } from 'svelte-i18n';
  import { configStore, getConfigValue, setConfigValue } from '../../lib/store/config';
  import ColorPicker, { ChromeVariant } from 'svelte-awesome-color-picker';
  import { isInternalChange } from '../../lib/store/internalChange';
  import Base from './Base.svelte';

  export let label = '';
  export let section = '';
  export let option = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = '';
  export let visible = true;
  export let onUpdate = null;

  let hex = '#FFFFFF';
  let prevConfigData = null;
  let configAvailable = false;

  $: {
    const storeValue = $configStore;
    configAvailable = !!storeValue?.path && storeValue.error === null;
  }

  async function loadValue() {
    if (!configAvailable) {
      hex = '#FFFFFF';
      return;
    }

    const value = await getConfigValue(section, option);
    hex =
      '#' +
      (section === 'DOTA2HPBARSOPTIONS' ? (value || 'FFFFFF').substring(2) : value || 'FFFFFF');
  }

  async function handleSave(newColor) {
    if (!configAvailable) return;

    try {
      let colorToSave = newColor.replace(/^#/, '').replace(/^FF/, '');
      if (section === 'DOTA2HPBARSOPTIONS') {
        colorToSave = 'FF' + colorToSave;
      }
      isInternalChange.mark();
      await setConfigValue(section, option, colorToSave);
      hex = newColor;
      onUpdate?.(newColor);
    } catch (err) {
      console.error('Ошибка сохранения значения цвета:', err);
    }
  }

  function handleInput(newColor) {
    hex = newColor;
  }

  function handleMouseUp() {
    handleSave(hex);
  }
</script>

<Base
  {label}
  {section}
  {option}
  {ttKey}
  {ttImage}
  {ttPlace}
  className="color-picker-wrapper {!configAvailable ? 'disabled' : ''}"
  style={visible ? '' : 'display: none;'}
  role="button"
  tabindex="-1"
  on:mouseup={handleMouseUp}
  let:configData
>
  {#if configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(), '')}
  {/if}

  <div class="cp-inner">
    {#if ChromeVariant}
      <ColorPicker
        bind:hex
        on:input={(e) => handleInput(e.hex || '#FFFFFF')}
        components={ChromeVariant}
        sliderDirection="horizontal"
        position="responsive"
        isAlpha={false}
        format="hex"
        disableTextInput={!configAvailable}
        label={$t(label)}
      />
    {:else}
      <span>Loading color picker...</span>
    {/if}
  </div>
</Base>

<style>
  :global(.color-picker-wrapper) {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 5px 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    width: fit-content;
  }

  :global(.color-picker-wrapper:hover) {
    background: var(--element-bg-hover-color);
  }

  /* Убираем стандартные стили ColorPicker */
  :global(.color-picker-wrapper > div) {
    background: transparent !important;
    border: none !important;
  }

  :global(.color-picker-wrapper.disabled) {
    pointer-events: none;
    cursor: not-allowed;
    opacity: 0.5;
    filter: grayscale(100%);
  }
</style>
