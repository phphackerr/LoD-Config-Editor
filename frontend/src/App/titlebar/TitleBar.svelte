<script>
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
    <span class="span-text">{$t('TITLE.window_title')}</span>
  </div>
  <div class="titlebar-center">
    <Search />
  </div>
  <div class="buttons">
    <button
      type="button"
      class="titlebar-button"
      on:click={openSettings}
      use:tt={{ content: $t('TITLE.settings_tooltip') }}
      aria-label={$t('TITLE.settings_tooltip')}
    >
      <div class="button-icon">
        <SettingsIc />
      </div>
    </button>
    <button
      class="titlebar-button"
      id="titlebar-minimize"
      on:click={Window.Minimise}
      use:tt={{ content: $t('TITLE.minimize_tooltip') }}
      aria-label={$t('TITLE.minimize_tooltip')}
    >
      <div class="button-icon">
        <MinimizeIc />
      </div>
    </button>
    <button
      class="titlebar-button"
      id="titlebar-maximize"
      on:click={Window.ToggleMaximise}
      use:tt={{ content: $t('TITLE.maximize_tooltip') }}
      aria-label={$t('TITLE.maximize_tooltip')}
    >
      <div class="button-icon">
        <MaximizeIc />
      </div>
    </button>
    <button
      class="titlebar-button close"
      id="titlebar-close"
      on:click={Application.Quit}
      aria-label={$t('TITLE.close_tooltip')}
      use:tt={{ content: $t('TITLE.close_tooltip') }}
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
    user-select: none;
    display: flex;
    justify-content: space-between;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;

    background: var(--titlebar-bg, rgb(59, 164, 117));
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
    margin-left: 10px;
    text-align: center;
    white-space: nowrap;

    color: var(--titlebar-text, rgb(33, 33, 33));
    font-size: var(--titlebar-text-size, 16px);
    font-weight: var(--titlebar-text-weight, 600) !important;
  }

  .titlebar-center {
    height: 100%;
    display: flex;
    align-items: center;
    pointer-events: none; /* чтобы не мешать drag-region */
    width: 300px;
    justify-content: center;
  }

  .titlebar-button {
    --wails-draggable: no-drag;
    display: inline-flex;
    justify-content: center;
    align-items: center;
    width: 40px;
    height: 30px;
    border: none;
    transition: 0.3s;

    background: var(--titlebar-button-bg, rgba(0, 0, 0, 0));
  }

  .titlebar-button:focus-visible {
    background: var(--titlebar-button-bg--hover, rgb(91, 190, 195));
    outline: none;
    border: 0.5px solid black;
  }

  .titlebar-button:hover {
    background: var(--titlebar-button-bg--hover, rgb(91, 190, 195));
  }

  .titlebar-button:hover .button-icon {
    color: var(--titlebar-button-icon-color--hover, rgb(33, 33, 33));
  }

  .close:focus {
    background: var(--titlebar-close-bg--hover, rgb(202, 51, 51));
  }

  .close:hover {
    background: var(--titlebar-close-bg--hover, rgb(202, 51, 51));
  }

  .close:hover .close-svg {
    color: var(--titlebar-close-icon-color--hover, rgb(33, 33, 33));
  }

  .button-icon {
    width: 24px;
    height: 24px;

    color: var(--titlebar-button-icon-color, rgb(33, 33, 33));
  }
</style>
