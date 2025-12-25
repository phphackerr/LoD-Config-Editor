<script>
  // @ts-nocheck
  import { createEventDispatcher } from 'svelte';
  import { t } from 'svelte-i18n';
  import { getConfigValue, setConfigValue } from '../../lib/store/config';
  import { isInternalChange } from '../../lib/store/internalChange';
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
  let prevConfigData = null;

  async function loadValue(configAvailable) {
    if (!configAvailable) {
      _value = '';
      return;
    }

    const val = await getConfigValue(section, option, '');
    _value = val;
  }

  async function handleChange(event, configAvailable) {
    if (!configAvailable) return;

    const newValue = event.target.value;

    if (value !== undefined) {
      // Controlled mode
      try {
        isInternalChange.mark();
        await setConfigValue(section, option, newValue);
        _value = newValue;
        onUpdate?.(newValue);
      } catch (err) {
        console.error('Ошибка сохранения dropdown:', err);
      }
      dispatch('change', { value: newValue });
      return;
    }

    // Uncontrolled mode
    try {
      isInternalChange.mark();
      await setConfigValue(section, option, newValue);
      _value = newValue;
      onUpdate?.(newValue);
    } catch (err) {
      console.error('Ошибка сохранения dropdown:', err);
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
    color: var(--color);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .dropdown select:hover {
    background-color: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .dropdown select:focus {
    outline: none;
    border-color: #ffd700;
    box-shadow: 0 0 0 2px rgba(255, 215, 0, 0.2);
  }

  .dropdown select option {
    background-color: #2a2a2a;
    color: #fff;
  }

  .dropdown select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    border-color: #666;
  }

  .dropdown.disabled {
    cursor: not-allowed;
    opacity: 0.5;
    background: rgba(255, 255, 255, 0.02);
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
