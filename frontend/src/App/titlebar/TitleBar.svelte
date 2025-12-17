<script>
  //@ts-nocheck
  import { Window, Application } from '@wailsio/runtime';
  import icon from '/favicon.png';
  import { SettingsIc, MinimizeIc, MaximizeIc, CloseIc } from '../lib/icons.js';
  import { tt } from '../lib/tooltip';
  import { t } from 'svelte-i18n';
  import Search from './Search.svelte';
  import { openSettings } from '../lib/store/settingsModal.js';
</script>

<div class="titlebar" on:dblclick={Window.ToggleMaximise} role="presentation">
  <div class="logo">
    <img src={icon} class="icon" alt="icon" />
    <span class="span-text">{$t('window_title')}</span>
  </div>
  <div class="titlebar-center">
    <Search />
  </div>
  <div class="buttons">
    <button
      type="button"
      class="titlebar-button"
      on:click={openSettings}
      use:tt={{ content: $t('settings_tooltip') }}
      aria-label="Settings"
    >
      <div class="button-icon">
        <SettingsIc />
      </div>
    </button>
    <button
      class="titlebar-button"
      id="titlebar-minimize"
      on:click={Window.Minimise}
      use:tt={{ content: $t('minimize_tooltip') }}
      aria-label="Minimize"
    >
      <div class="button-icon">
        <MinimizeIc />
      </div>
    </button>
    <button
      class="titlebar-button"
      id="titlebar-maximize"
      on:click={Window.ToggleMaximise}
      use:tt={{ content: $t('maximize_tooltip') }}
      aria-label="Maximize"
    >
      <div class="button-icon">
        <MaximizeIc />
      </div>
    </button>
    <button
      class="titlebar-button close"
      id="titlebar-close"
      on:click={Application.Quit}
      aria-label="Close"
      use:tt={{ content: $t('close_tooltip') }}
    >
      <div class="button-icon close-svg">
        <CloseIc />
      </div>
    </button>
  </div>
</div>

<style>
  .titlebar {
    --wails-draggable: drag;
    height: 30px;
    background: var(--titlebar-bg-color);
    user-select: none;
    display: flex;
    justify-content: space-between;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
  }

  .logo {
    display: flex;
    align-items: center;
    padding-left: 4px;
  }

  .icon {
    width: 26px;
    height: 26px;
  }

  .span-text {
    color: var(--titlebar-logo-text-color);
    margin-left: 10px;
    text-align: center;
    font-size: 16px;
    font-weight: 600 !important;
    white-space: nowrap;
  }

  .titlebar-center {
    height: 100%;
    display: flex;
    align-items: center;
    pointer-events: none; /* чтобы не мешать drag-region */
    width: 300px; /* или auto, если нужно */
    justify-content: center;
  }

  .titlebar-button {
    --wails-draggable: no-drag;
    display: inline-flex;
    justify-content: center;
    align-items: center;
    width: 40px;
    height: 30px;
    background: var(--titlebar-button-bg-color);
    border: none;
    transition: 0.3s;
  }

  .titlebar-button:focus-visible {
    background: var(--titlebar-button-bg-color-hover);
    outline: none;
    border: 0.5px solid black;
  }

  .titlebar-button:hover {
    background: var(--titlebar-button-bg-color-hover);
  }

  .titlebar-button:hover .button-icon {
    color: var(--titlebar-button-color-hover);
  }

  .close:focus {
    background: var(--titlebar-close-button-bg-color-hover);
  }

  .close:hover {
    background: var(--titlebar-close-button-bg-color-hover);
  }

  .close:hover .close-svg {
    color: var(--titlebar-close-button-color-hover);
  }

  .button-icon {
    color: var(--titlebar-button-color);
    width: 24px;
    height: 24px;
  }
</style>
