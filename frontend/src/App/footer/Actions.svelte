<script>
  import { onDestroy } from 'svelte';
  import { LaunchGameExe, OpenFolderInExplorer, OpenFile } from '/bindings/lce/backend/utils/utils';
  import { appSettings } from '../lib/store/appSettings';
  import { notifyError } from '../lib/store/notifications';
  import { t } from 'svelte-i18n';
  import { tt } from '../lib/tooltip';
  import { toErrorMessage } from '../lib/store/storeUtils';

  let hovered = false;
  let focused = false;
  let actionError = '';
  let clearErrorTimeout = null;

  let configBtn = null;
  let folderBtn = null;
  let launchBtn = null;

  function setActionError(message) {
    actionError = String(message || '').trim();

    if (clearErrorTimeout) {
      clearTimeout(clearErrorTimeout);
      clearErrorTimeout = null;
    }

    if (actionError) {
      clearErrorTimeout = setTimeout(() => {
        actionError = '';
        clearErrorTimeout = null;
      }, 6000);
    }
  }

  async function runAction(action, fallbackMessage) {
    try {
      actionError = '';
      await action();
      return true;
    } catch (error) {
      const message = toErrorMessage(error, fallbackMessage);
      setActionError(message);
      notifyError(message);
      return false;
    }
  }

  async function launchGame() {
    const gamePath = String($appSettings.game_path || '').trim();
    if (!gamePath) return;

    const args = $appSettings.windowed_mode ? ['-window'] : [];
    await runAction(() => LaunchGameExe(gamePath, ...args), $t('ERRORS.actions.launch_game'));
  }

  async function openFolder() {
    const gamePath = String($appSettings.game_path || '').trim();
    if (!gamePath) return;

    await runAction(() => OpenFolderInExplorer(gamePath), $t('ERRORS.actions.open_game_folder'));
  }

  async function openConfig() {
    const gamePath = String($appSettings.game_path || '').trim();
    if (!gamePath) return;

    await runAction(
      () => OpenFile(`${gamePath}\\config.lod.ini`),
      $t('ERRORS.actions.open_config_file')
    );
  }

  function handleFocusOut(event) {
    // Скрываем кнопки, только если фокус уходит за пределы всей панели
    if (!event?.currentTarget?.contains(event.relatedTarget)) {
      focused = false;
    }
  }

  function handleKeydown(event) {
    // Нас интересуют только стрелки влево и вправо
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') {
      return;
    }

    // Предотвращаем стандартное поведение браузера (например, прокрутку страницы)
    event.preventDefault();

    // Собираем видимые кнопки в массив для удобной навигации
    const buttons = [configBtn, folderBtn, launchBtn].filter(Boolean);
    if (buttons.length <= 1) return; // Нечего переключать

    const currentIndex = buttons.indexOf(document.activeElement);
    if (currentIndex === -1) return; // Фокус не на одной из наших кнопок

    let nextIndex;
    if (event.key === 'ArrowRight') {
      // Переход вправо, с зацикливанием в начало
      nextIndex = (currentIndex + 1) % buttons.length;
    } else {
      // Переход влево, с зацикливанием в конец
      nextIndex = (currentIndex - 1 + buttons.length) % buttons.length;
    }

    // Устанавливаем фокус на следующую кнопку
    buttons[nextIndex]?.focus();
  }

  onDestroy(() => {
    if (clearErrorTimeout) {
      clearTimeout(clearErrorTimeout);
      clearErrorTimeout = null;
    }
  });
</script>

<div
  class="panel"
  on:mouseleave={() => (hovered = false)}
  on:focusout={handleFocusOut}
  on:keydown={handleKeydown}
  style="background-color: {hovered || focused
    ? 'var(--surface-overlay, rgba(0, 0, 0, 0.5))'
    : ''}; display: {$appSettings.game_path ? 'flex' : 'none'};"
  role="presentation"
>
  {#if $appSettings.game_path}
    {#if actionError}
      <div class="action-error" role="status" aria-live="polite">{actionError}</div>
    {/if}

    <button
      class="btn hidden"
      class:visible={hovered || focused}
      on:click={openConfig}
      use:tt={{ content: $t('ACTIONS.open_config'), placement: 'top' }}
      bind:this={configBtn}>🗎</button
    >
    <button
      class="btn hidden"
      class:visible={hovered || focused}
      on:click={openFolder}
      use:tt={{ content: $t('ACTIONS.open_game_folder'), placement: 'top' }}
      bind:this={folderBtn}>🗁</button
    >
    <button
      class="btn"
      on:mouseenter={() => (hovered = true)}
      on:focusin={() => (focused = true)}
      on:click={launchGame}
      use:tt={{ content: $t('ACTIONS.launch_game'), placement: 'top' }}
      bind:this={launchBtn}>▷</button
    >
  {/if}
</div>

<style>
  .panel {
    position: fixed;
    bottom: 15px;
    right: 15px;
    display: flex;
    gap: 15px;
    padding: 10px;
    border-radius: 10px;
    transition: all 0.3s ease;
    z-index: 999;
  }

  .action-error {
    position: absolute;
    right: 0;
    bottom: calc(100% + 8px);
    max-width: 320px;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    color: var(--status-error-text, #fecaca);
    font-size: 13px;
    line-height: 1.3;
    user-select: text;
    pointer-events: none;
  }

  .btn {
    width: 50px;
    height: 50px;
    background-color: var(--action-primary-bg, #3ba475);
    border: none;
    border-radius: 14px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
    color: var(--text-color-primary, #fff);
    font-size: 24px;
    text-align: center;
    cursor: pointer;
  }

  .hidden {
    opacity: 0;
    display: none;
    visibility: hidden;
    pointer-events: none;
    transition: all 0.3s ease;
  }

  .hidden.visible {
    opacity: 1;
    display: block;
    visibility: visible;
    pointer-events: all;
  }
</style>
