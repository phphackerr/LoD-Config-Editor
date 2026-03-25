<script>
  import { createEventDispatcher } from 'svelte';
  import { t } from 'svelte-i18n';

  export let number;

  const dispatch = createEventDispatcher();
  const units = ['', 'px', 'rem', 'em', '%', 'vw', 'vh'];

  let value = 0;
  let unit = 'px';

  $: if (number) {
    value = Number.isFinite(number.value) ? number.value : 0;
    unit = typeof number.unit === 'string' ? number.unit : 'px';
  }

  function emitChange() {
    dispatch('change', {
      value: Number.isFinite(Number(value)) ? Number(value) : 0,
      unit
    });
  }

  function handleValueInput(event) {
    value = Number(event.currentTarget?.value ?? 0);
    emitChange();
  }

  function handleUnitChange(event) {
    unit = String(event.currentTarget?.value ?? '');
    emitChange();
  }
</script>

<div class="number-editor">
  <label>
    {$t('COMMON.value')}
    <input type="number" {value} step="0.1" on:input={handleValueInput} />
  </label>

  <label>
    {$t('SETTINGS.THEME_EDITOR.unit')}
    <select value={unit} on:change={handleUnitChange}>
      {#each units as u}
        <option value={u}>{u || $t('SETTINGS.THEME_EDITOR.none')}</option>
      {/each}
    </select>
  </label>
</div>

<style>
  .number-editor {
    display: grid;
    grid-template-columns: 1fr 160px;
    gap: 12px;
    max-width: 460px;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    color: var(--theme-editor-hint-text, var(--text-color-muted, #999));
  }

  input,
  select {
    width: 100%;
    padding: 8px 10px;
    border-radius: 6px;
    border: 1px solid var(--theme-editor-input-border, var(--border-color, #444));
    background: var(--theme-editor-input-bg, var(--surface-base, #222));
    color: var(--text-color-primary, #fff);
    font-size: 13px;
  }

  input:focus,
  select:focus {
    outline: none;
    border-color: var(--theme-editor-input-border-focus, var(--action-primary, #3ba475));
  }
</style>
