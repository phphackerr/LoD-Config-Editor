<script module>
  export const tabMetadata = {
    order: 1
  };
</script>

<script>
  import { t } from 'svelte-i18n';
  import { appSettings, appSettingsState, updateWindowedMode } from '../../lib/store/appSettings';

  async function handleWindowedModeChange(event) {
    const checked = Boolean(event?.currentTarget?.checked);
    await updateWindowedMode(checked);
  }
</script>

<div class="general-container">
  {#if $appSettingsState.error && $appSettingsState.operation === 'appSettings:update'}
    <div class="error-message">{$appSettingsState.error}</div>
  {/if}

  <div class="setting-item">
    <label class="checkbox">
      <input
        type="checkbox"
        checked={$appSettings.windowed_mode}
        on:change={handleWindowedModeChange}
      />
      <span>{$t('SETTINGS.GENERAL.windowed_mode')}</span>
    </label>
  </div>
</div>

<style>
  .general-container {
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 20px;
    color: var(--text-color);
  }

  .setting-item {
    display: flex;
    align-items: center;
  }

  .error-message {
    color: var(--status-error-text, #fecaca);
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    border-radius: 8px;
    padding: 10px 12px;
    font-size: 14px;
  }

  .checkbox {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    user-select: none;
  }

  .checkbox input[type='checkbox'] {
    appearance: none;
    -webkit-appearance: none;
    width: 18px;
    height: 18px;
    border: 2px solid var(--text-color);
    border-radius: 4px;
    position: relative;
    cursor: pointer;
    transition: all 0.2s;
    background: transparent;
  }

  .checkbox input[type='checkbox']:checked {
    background: var(--action-primary-bg, #3ba475);
    border-color: var(--action-primary-bg, #3ba475);
  }

  .checkbox input[type='checkbox']:checked::after {
    content: '✓';
    position: absolute;
    color: var(--text-color-primary, white);
    font-size: 14px;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  .checkbox input[type='checkbox']:hover {
    border-color: var(--action-primary-bg, #3ba475);
  }
</style>
