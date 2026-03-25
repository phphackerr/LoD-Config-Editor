<script>
  import { createEventDispatcher } from 'svelte';
  import { t } from 'svelte-i18n';
  import { getConfigValue } from '../../lib/store/config';
  import { persistControlValue } from './lib/controlPipeline';
  import Base from './Base.svelte';

  const dispatch = createEventDispatcher();

  export let label = '';
  export let section = '';
  export let option = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = 'auto';
  export let visible = true;
  export let onUpdate = null;
  export let width = 200;
  export let height = 40;
  export let padding = 0;
  export let text_size = 15;
  export let options = [];
  export let options_keys = [];

  // Controlled mode → если родитель передал bind:value
  export let value = undefined;

  // Uncontrolled mode → внутреннее состояние
  let _value = '';
  let lastSavedValue = '';
  let prevConfigData = null;

  async function loadValue(configAvailable) {
    if (!configAvailable) {
      _value = '';
      return;
    }

    const val = String((await getConfigValue(section, option)) || '');
    _value = val;
    lastSavedValue = val;
  }

  async function handleChange(event, configAvailable) {
    if (!configAvailable) return;

    const newValue = String(event?.target?.value || '');
    const previousValue =
      value !== undefined ? String(value ?? lastSavedValue) : String(lastSavedValue || _value);

    const result = await persistControlValue(
      section,
      option,
      newValue,
      $t('ERRORS.controls.save_dropdown')
    );

    if (value !== undefined) {
      const nextValue = result.ok ? newValue : previousValue;
      value = nextValue;
      _value = nextValue;

      if (result.ok) {
        _value = newValue;
        lastSavedValue = newValue;
        onUpdate?.(newValue);
      }
      dispatch('change', { value: nextValue, error: result.error });
      return;
    }

    if (result.ok) {
      _value = newValue;
      lastSavedValue = newValue;
      onUpdate?.(newValue);
      return;
    }

    _value = previousValue;
  }
</script>

<Base
  {label}
  {section}
  {option}
  {ttKey}
  {ttImage}
  {ttPlace}
  className="dropdown-wrapper"
  style={visible ? '' : 'display: none;'}
  tabindex="0"
  role="button"
  let:configAvailable
  let:configData
>
  {#if value === undefined && configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(configAvailable), '')}
  {/if}

  <label class="dropdown" class:disabled={!configAvailable}>
    <span>{$t(label)}</span>

    {#if value !== undefined}
      <!-- Controlled mode -->
      <select
        bind:value
        on:change={(e) => handleChange(e, configAvailable)}
        disabled={!configAvailable}
        style="
                    width: {width}px;
                    height: {height}px;
                    padding: {padding}px;
                    font-size: {text_size}px;
                "
      >
        {#each options as optionLabel, i}
          <option value={options_keys[i]}>
            {optionLabel}
          </option>
        {/each}
      </select>
    {:else}
      <!-- Uncontrolled mode -->
      <select
        bind:value={_value}
        on:change={(e) => handleChange(e, configAvailable)}
        disabled={!configAvailable}
        style="
                    width: {width}px;
                    height: {height}px;
                    padding: {padding}px;
                    font-size: {text_size}px;
                "
      >
        {#each options as optionLabel, i}
          <option value={options_keys[i]}>
            {optionLabel}
          </option>
        {/each}
      </select>
    {/if}
  </label>
</Base>

<style>
  :global(.dropdown-wrapper) {
    display: inline-flex;
    width: fit-content;
  }

  .dropdown {
    display: flex;
    gap: 10px;
    align-items: center;
    cursor: pointer;
    user-select: none;
    padding: 5px 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    width: fit-content;
  }

  .dropdown:hover {
    background: var(--element-bg-hover-color);
  }

  .dropdown select {
    background-color: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--dd-select-border-color);
    border-radius: 4px;
    color: var(--text-color-primary);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .dropdown select:hover {
    background-color: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .dropdown select:focus {
    outline: none;
    border-color: var(--control-focus-border-color, #ffd700);
    box-shadow: 0 0 0 2px var(--control-focus-ring-color, rgba(255, 215, 0, 0.2));
  }

  .dropdown select option {
    background-color: var(--surface-elevated, #2a2a2a);
    color: var(--text-color-primary, #fff);
  }

  .dropdown select:disabled {
    opacity: var(--control-disabled-opacity, 0.5);
    cursor: not-allowed;
    border-color: var(--control-disabled-border-color, #666);
  }

  .dropdown.disabled {
    cursor: not-allowed;
    opacity: var(--control-disabled-opacity, 0.5);
    background: var(--control-disabled-bg-color, rgba(255, 255, 255, 0.02));
  }

  /* Стилизация скроллбара */
  .dropdown select::-webkit-scrollbar {
    width: 8px;
  }

  .dropdown select::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }

  .dropdown select::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 4px;
  }

  .dropdown select::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.3);
  }
</style>
