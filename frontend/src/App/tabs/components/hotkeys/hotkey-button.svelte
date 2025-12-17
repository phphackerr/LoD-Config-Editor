<script>
  // @ts-nocheck
  import { onMount } from 'svelte';
  import HotkeyCaptureModal from './HotkeyCaptureModal.svelte';
  import { getConfigValue, isConfigAvailable } from '../../../lib/store/config';
  import { normalizeKey } from './keyFuncs';
  import { t } from 'svelte-i18n';
  import Base from '../Base.svelte';

  export let size = ['70', '70'];
  export let color = '#FFFFFF';
  export let visible = true;
  export let imageSrc = '';
  export let alignment = 'center';
  export let section = '';
  export let option = '';
  export let disabled = false;
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = '';

  let hotkeyValue = '';
  let showModal = false;
  let prevConfigData = null;

  async function loadValue(configAvailable) {
    if (!configAvailable || option == 'ExtraSlot1') {
      hotkeyValue = '';
      return;
    }

    const savedValue = await getConfigValue(section, option);
    hotkeyValue = normalizeKey(savedValue);
  }

  function openModal(configAvailable) {
    if (!configAvailable || disabled) return;
    showModal = true;
  }

  function handleModalClose(configAvailable) {
    showModal = false;
    loadValue(configAvailable); // Перезагружаем значение после закрытия модалки
  }
</script>

<Base
  label=""
  {ttKey}
  {ttImage}
  {ttPlace}
  className="hotkey-button-wrapper"
  style={visible ? '' : 'display: none;'}
  let:configAvailable
  let:configData
>
  {#if configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(configAvailable), '')}
  {/if}

  <button
    class="hotkey-button"
    style="
            width: {size[0]}px;
            height: {size[1]}px;
            justify-content: {alignment};
        "
    on:click={() => openModal(configAvailable)}
    disabled={disabled || !configAvailable}
  >
    <div class="hotkey-content">
      <img src={imageSrc} alt="" class="hotkey-image" />
      <div role="button" tabindex="-1" class="hotkey-text" style="color: {color}">
        {hotkeyValue}
      </div>
    </div>
  </button>

  {#if showModal}
    <HotkeyCaptureModal {section} {option} onclose={() => handleModalClose(configAvailable)} />
  {/if}
</Base>

<style>
  :global(.hotkey-button-wrapper) {
    display: inline-block;
    vertical-align: top;
  }

  .hotkey-button {
    padding: 0;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 2px;
    overflow: hidden;
    position: relative;
    min-width: 50px;
    min-height: 50px;
    aspect-ratio: 1 / 1;
    display: block;
  }

  .hotkey-button:disabled {
    cursor: default;
    opacity: 0.5;
  }

  .hotkey-content {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    position: relative;
  }

  .hotkey-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    position: absolute;
    top: 0;
    left: 0;
  }

  .hotkey-text {
    position: relative;
    z-index: 1;
    text-shadow:
      1px 1px 0 #000,
      -1px -1px 0 #000,
      1px -1px 0 #000,
      -1px 1px 0 #000,
      0 0 4px rgba(0, 0, 0, 0.8);
    font-size: 12px;
    font-weight: 600;
    text-align: center;
  }
</style>
