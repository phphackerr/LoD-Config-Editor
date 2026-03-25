<script>
  import { tick } from 'svelte';
  import { t } from 'svelte-i18n';
  import { configStore, getConfigValue } from '../../lib/store/config';
  import { CheckIc, EditIc, LoopIc } from '../../lib/icons';
  import { persistControlValue } from './lib/controlPipeline';
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
    if (!configAvailable) return false;

    let newValue = parseFloat(event?.target?.value);
    if (Number.isNaN(newValue)) return false;
    if (currentValueType === 'float') newValue = Math.round(newValue * 10) / 10;
    else newValue = Math.round(newValue);

    const result = await persistControlValue(
      section,
      option,
      newValue.toString(),
      $t('ERRORS.controls.save_slider')
    );
    if (!result.ok) return false;

    value = newValue;
    return true;
  }

  async function toggleValueType() {
    if (option !== 'CameraHeight') return;

    const previousState = {
      value,
      currentValueType,
      currentMin,
      currentMax,
      currentStep
    };

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

    const saved = await handleChange({ target: { value: value.toString() } });
    if (!saved) {
      value = previousState.value;
      currentValueType = previousState.currentValueType;
      currentMin = previousState.currentMin;
      currentMax = previousState.currentMax;
      currentStep = previousState.currentStep;
      if (isEditing) editValue = value.toString();
      if (rangeInput) rangeInput.value = String(value);
    }
  }

  async function startEditing() {
    isEditing = true;
    editValue = value.toString();
    await tick();
    editInputElement?.focus();
    editInputElement?.select();
  }

  async function submitEdit() {
    const newValue = parseFloat(editValue);
    if (!isNaN(newValue)) {
      const boundedValue = Math.min(Math.max(newValue, currentMin), currentMax);
      const saved = await handleChange({ target: { value: boundedValue.toString() } });
      if (!saved) {
        editValue = value.toString();
      }
    }
    isEditing = false;
  }

  async function handleKeydown(e) {
    if (!configAvailable) return;

    const previousValue = value;
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
      rangeInput.value = String(value);
    }

    const saved = await handleChange({ target: { value } });
    if (!saved) {
      value = previousValue;
      if (rangeInput) {
        rangeInput.value = String(previousValue);
      }
    }
  }
</script>

<Base
  {label}
  {section}
  {option}
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
          aria-label={$t('CONTROLS.toggle_value_type')}
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
              if (e.key === 'Enter') {
                submitEdit();
              }
            }}
            on:blur={() => submitEdit()}
            step={currentStep}
            min={currentMin}
            max={currentMax}
          />
          <button
            class="icon-button"
            on:click={() => submitEdit()}
            aria-label={$t('COMMON.confirm')}
          >
            <div class="icon">
              <CheckIc />
            </div>
          </button>
        </div>
      {:else}
        <button class="icon-button" on:click={startEditing} aria-label={$t('COMMON.edit')}>
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
    on:change={handleChange}
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
    color: var(--accent-color, #ffd700);
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
    color: var(--text-color-primary);
    background: var(--element-bg-color, rgba(255, 255, 255, 0.1));
    border: 1px solid var(--element-bg-hover-color, rgba(255, 255, 255, 0.2));
    border-radius: 4px;
    padding: 4px 8px;
  }

  .edit-container input:focus {
    outline: none;
    border-color: var(--control-focus-border-color, var(--accent-color, #ffd700));
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
    background: var(--accent-color, #ffd700);
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
  }

  .slider::-moz-range-thumb {
    width: 16px;
    height: 16px;
    background: var(--accent-color, #ffd700);
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
    color: var(--text-color-primary);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .icon-button:hover {
    background: var(--control-icon-hover-bg, rgba(255, 255, 255, 0.1));
  }

  :global(.slider-wrapper.disabled) {
    pointer-events: none;
    cursor: not-allowed;
    opacity: var(--control-disabled-opacity, 0.5);
    filter: grayscale(100%);
  }
</style>
