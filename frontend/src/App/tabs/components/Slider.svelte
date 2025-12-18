<script>
  // @ts-nocheck
  import { tick } from 'svelte';
  import { t } from 'svelte-i18n';
  import { configStore, getConfigValue, setConfigValue } from '../../lib/store/config';
  import { isInternalChange } from '../../lib/store/internalChange';
  import { CheckIc, EditIc, LoopIc } from '../../lib/icons';
  import Base from './Base.svelte';

  export let label = '';
  export let section = '';
  export let option = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = '';
  export let visible = true;
  export let min = 0;
  export let max = 100;
  export let step = 1;
  export let defaultValue = 0;
  export let valueType = 'int'; // "int" или "float"

  let value = defaultValue;
  let prevConfigData = null;
  let configAvailable = false;

  $: {
    const storeValue = $configStore;
    configAvailable = !!storeValue?.path && storeValue.error === null;
  }

  let currentValueType = valueType;
  let currentMin = min;
  let currentMax = max;
  let currentStep = step;

  let rangeInput;

  let isEditing = false;
  let editValue = '';
  let editInputElement;

  async function loadValue() {
    if (!configAvailable) {
      value = defaultValue;
      return;
    }

    let configValue = await getConfigValue(section, option);
    if (!configValue) configValue = defaultValue.toString();

    if (option === 'CameraHeight') {
      const numValue = parseFloat(configValue);
      if (Number.isInteger(numValue)) {
        currentValueType = 'int';
        currentMin = 1650;
        currentMax = 8250;
        currentStep = 1;
      } else {
        currentValueType = 'float';
        currentMin = 1.0;
        currentMax = 5.0;
        currentStep = 0.1;
      }
    }

    value = currentValueType === 'int' ? parseInt(configValue) : parseFloat(configValue);
    value = Math.min(Math.max(value, currentMin), currentMax);
    if (currentValueType === 'float') value = Math.round(value * 10) / 10;
  }

  async function handleChange(event) {
    if (!configAvailable) return;

    let newValue = parseFloat(event.target.value);
    if (currentValueType === 'float') newValue = Math.round(newValue * 10) / 10;
    else newValue = Math.round(newValue);

    isInternalChange.mark();
    await setConfigValue(section, option, newValue.toString());
    value = newValue;
  }

  function toggleValueType() {
    if (option !== 'CameraHeight') return;

    if (currentValueType === 'int') {
      value = +(value / 1650).toFixed(1);
      currentValueType = 'float';
      currentMin = 1.0;
      currentMax = 5.0;
      currentStep = 0.1;
    } else {
      value = Math.round(value * 1650);
      currentValueType = 'int';
      currentMin = 1650;
      currentMax = 8250;
      currentStep = 50;
    }

    if (isEditing) editValue = value.toString();
    handleChange({ target: { value: value.toString() } });
  }

  async function startEditing() {
    isEditing = true;
    editValue = value.toString();
    await tick();
    editInputElement?.focus();
    editInputElement?.select();
  }

  function submitEdit() {
    const newValue = parseFloat(editValue);
    if (!isNaN(newValue)) {
      const boundedValue = Math.min(Math.max(newValue, currentMin), currentMax);
      handleChange({ target: { value: boundedValue.toString() } });
    }
    isEditing = false;
  }

  function handleKeydown(e) {
    if (!configAvailable) return;

    let newValue = value;

    if (e.key === 'ArrowRight' || e.key === 'ArrowUp') {
      newValue = Math.min(value + currentStep, currentMax);
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowDown') {
      newValue = Math.max(value - currentStep, currentMin);
    } else {
      return;
    }

    e.preventDefault();
    value = newValue;

    if (rangeInput) {
      rangeInput.value = value;
    }

    handleChange({ target: { value } });
  }
</script>

<Base
  {label}
  {ttKey}
  {ttImage}
  {ttPlace}
  className="slider-wrapper {!configAvailable ? 'disabled' : ''}"
  style={visible ? '' : 'display: none;'}
  tabindex="0"
  role="slider"
  aria-valuenow={value}
  on:keydown={handleKeydown}
  let:configData
>
  <!-- Note: handleKeydown needs configAvailable, but we can't easily pass it from on:keydown unless we wrap it. 
         However, Base forwards on:keydown. We can bind to the div inside Base? 
         Actually, Base forwards the event to the parent. 
         Wait, Base `on:keydown` forwards to `Slider`'s usage of `Base`.
         So `<Base on:keydown={...} />` works.
         But we need `configAvailable` inside `handleKeydown`.
         We can get it from the let: directive? No, that's for the slot.
         
         Workaround: We can use a reactive statement to update a local `configAvailable` variable 
         or just pass it to the handler in the template if possible.
         
         Actually, `handleKeydown` is called when the `div` (Base) is focused and key is pressed.
         We can just use `configAvailable` from the `let:` if we put the handler on an element inside the slot?
         No, the wrapper needs to handle keydown for accessibility (role="slider").
         
         Let's use a reactive statement to sync configAvailable from the store directly or just rely on the store import.
         Since we import `configStore` in `Base`, we can also import it here? 
         Yes, we can just use `$configStore` here too for the logic, but `Base` handles the "availability" logic.
         
         Let's just use `$configStore` here for `handleKeydown` check, duplicating the check slightly but keeping it safe.
         Or better: `Base` exposes `configAvailable` to the slot. 
         But `on:keydown` is on `Base`.
         
         Let's just use the store here.
    -->

  {#if configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(), '')}
  {/if}

  <div class="slider-header">
    <span class="label">
      <span>{$t(label)}</span>:
      <span class="value">{value}</span>
    </span>

    <div class="header-buttons">
      {#if option === 'CameraHeight'}
        <button
          class="icon-button"
          on:click={() => toggleValueType()}
          aria-label="Переключить тип значения"
        >
          <div class="icon">
            <LoopIc />
          </div>
        </button>
      {/if}

      {#if isEditing}
        <div class="edit-container">
          <input
            type="number"
            bind:this={editInputElement}
            bind:value={editValue}
            on:keydown={(e) => {
              if (e.key === 'Enter') submitEdit();
            }}
            on:blur={() => submitEdit()}
            step={currentStep}
            min={currentMin}
            max={currentMax}
          />
          <button class="icon-button" on:click={() => submitEdit()} aria-label="Подтвердить">
            <div class="icon">
              <CheckIc />
            </div>
          </button>
        </div>
      {:else}
        <button class="icon-button" on:click={startEditing} aria-label="Редактировать">
          <div class="icon">
            <EditIc />
          </div>
        </button>
      {/if}
    </div>
  </div>

  <input
    type="range"
    class="slider"
    bind:value
    bind:this={rangeInput}
    min={currentMin}
    max={currentMax}
    step={currentStep}
    on:change={(e) => handleChange(e)}
    disabled={!configAvailable}
    tabindex="-1"
  />
</Base>

<style>
  :global(.slider-wrapper) {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
  }

  .slider-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .label {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .value {
    color: #ffd700;
  }

  .header-buttons {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .icon {
    width: 28px;
    height: 28px;
  }

  .edit-container {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .edit-container input {
    width: 80px;
    color: var(--color);
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 4px;
    padding: 4px 8px;
  }

  .edit-container input:focus {
    outline: none;
    border-color: #ffd700;
  }

  .slider {
    appearance: none;
    -webkit-appearance: none;
    width: 100%;
    height: 4px;
    background: var(--slider-line-color);
    border-radius: 2px;
    outline: none;
  }

  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 16px;
    height: 16px;
    background: #ffd700;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
  }

  .slider::-moz-range-thumb {
    width: 16px;
    height: 16px;
    background: #ffd700;
    border-radius: 50%;
    cursor: pointer;
    border: none;
  }

  .slider::-webkit-slider-thumb:hover {
    transform: scale(1.1);
  }
  .slider::-moz-range-thumb:hover {
    transform: scale(1.1);
  }

  .slider:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .icon-button {
    background: none;
    border: none;
    color: var(--color);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .icon-button:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  :global(.slider-wrapper.disabled) {
    pointer-events: none;
    cursor: not-allowed;
    opacity: 0.5;
    filter: grayscale(100%);
  }
</style>
