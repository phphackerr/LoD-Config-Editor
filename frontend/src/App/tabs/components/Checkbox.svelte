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
  export let reverted = false;

  // Если родитель передал bind:checked → controlled mode
  export let checked = undefined;

  // Внутренний стейт для uncontrolled mode
  let _checked = false;
  let prevConfigData = null;
  let inputElement;

  async function loadValue(configAvailable) {
    if (!configAvailable) {
      _checked = false;
      return;
    }

    const value = String((await getConfigValue(section, option)) || '').toLowerCase();
    const isTrue = value === 'true';
    _checked = reverted ? !isTrue : isTrue;
  }

  async function handleChange(event, configAvailable) {
    if (!configAvailable) return;

    const newValue = Boolean(event?.target?.checked);
    const previousValue = checked !== undefined ? Boolean(checked) : Boolean(_checked);
    const persistedValue = (reverted ? !newValue : newValue).toString();

    const result = await persistControlValue(
      section,
      option,
      persistedValue,
      $t('ERRORS.controls.save_checkbox')
    );

    const nextValue = result.ok ? newValue : previousValue;
    _checked = nextValue;
    if (checked !== undefined) {
      checked = nextValue;
      dispatch('checked', { checked: nextValue, error: result.error });
    }

    if (result.ok) {
      onUpdate?.(nextValue);
    }
  }

  function handleKeyDown(event) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      inputElement?.click();
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
  className="checkbox-wrapper"
  style={visible ? '' : 'display: none;'}
  role="button"
  tabindex="0"
  on:keydown={handleKeyDown}
  let:configAvailable
  let:configData
>
  <!-- Реакция на изменение конфига -->
  {#if checked === undefined && configData !== prevConfigData}
    {((prevConfigData = configData), loadValue(configAvailable), '')}
  {/if}

  <label class="checkbox" class:disabled={!configAvailable}>
    {#if checked !== undefined}
      <!-- Controlled mode -->
      <input
        type="checkbox"
        bind:checked
        on:change={(e) => handleChange(e, configAvailable)}
        disabled={!configAvailable}
        tabindex="-1"
        bind:this={inputElement}
      />
    {:else}
      <!-- Uncontrolled mode -->
      <input
        type="checkbox"
        bind:checked={_checked}
        on:change={(e) => handleChange(e, configAvailable)}
        disabled={!configAvailable}
        tabindex="-1"
        bind:this={inputElement}
      />
    {/if}

    <span>{$t(label)}</span>
  </label>
</Base>

<style>
  /* .checkbox-wrapper теперь стилизуется через className в Base, но стили должны быть глобальными или передаваться */
  /* Svelte scoped styles не применяются к элементам внутри дочернего компонента, если они не переданы как класс */
  /* Но Base рендерит div с классом checkbox-wrapper. Чтобы стилизовать его отсюда, нужно :global */

  :global(.checkbox-wrapper) {
    display: inline-flex;
    width: fit-content;
  }

  .checkbox {
    display: flex;
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

  .checkbox:hover {
    background: var(--element-bg-hover-color);
  }

  .checkbox input[type='checkbox'] {
    appearance: none;
    -webkit-appearance: none;
    width: 16px;
    height: 16px;
    border: 2px solid var(--text-color-primary);
    border-radius: 4px;
    position: relative;
    cursor: pointer;
    transition: all 0.2s;
  }

  .checkbox input[type='checkbox']:checked {
    background: var(--control-active-bg-color, #ffd700);
    border-color: var(--control-active-bg-color, #ffd700);
  }

  .checkbox input[type='checkbox']:checked::after {
    content: '✓';
    position: absolute;
    color: var(--control-active-text-color, #000);
    font-size: 12px;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  .checkbox input[type='checkbox']:hover {
    border-color: var(--control-focus-border-color, #ffd700);
  }

  .checkbox input[type='checkbox']:disabled {
    opacity: var(--control-disabled-opacity, 0.5);
    cursor: not-allowed;
    border-color: var(--control-disabled-border-color, #666);
  }

  .checkbox.disabled {
    cursor: not-allowed;
    opacity: var(--control-disabled-opacity, 0.5);
    background: var(--control-disabled-bg-color, rgba(255, 255, 255, 0.02));
  }
</style>
