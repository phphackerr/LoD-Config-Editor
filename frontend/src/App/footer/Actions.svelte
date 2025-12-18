<script>
  //@ts-nocheck
  import { LaunchGameExe, OpenFolderInExplorer, OpenFile } from '/bindings/lce/backend/utils/utils';
  import { appSettings } from '../lib/store/appSettings';
  import { t } from 'svelte-i18n';
  import { tt } from '../lib/tooltip';

  let hovered = false;
  let focused = false;

  let configBtn, folderBtn, launchBtn;

  async function launchGame() {
    await LaunchGameExe($appSettings.game_path);
  }

  async function openFolder() {
    await OpenFolderInExplorer($appSettings.game_path);
  }

  async function openConfig() {
    await OpenFile($appSettings.game_path + '\\config.lod.ini');
  }

  function handleFocusOut(event) {
    // Скрываем кнопки, только если фокус уходит за пределы всей панели
    if (!event.currentTarget.contains(event.relatedTarget)) {
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
</script>

<div
  class="panel"
  on:mouseleave={() => (hovered = false)}
  on:focusout={handleFocusOut}
  on:keydown={handleKeydown}
  style="background-color: {hovered || focused
    ? 'rgba(0, 0, 0, 0.5)'
    : ''}; display: {$appSettings.game_path ? 'flex' : 'none'};"
  role="presentation"
>
  {#if $appSettings.game_path}
    <button
      class="btn hidden"
      class:visible={hovered || focused}
      on:click={openConfig}
      use:tt={{ content: $t('open_config'), placement: 'top' }}
      bind:this={configBtn}>🗎</button
    >
    <button
      class="btn hidden"
      class:visible={hovered || focused}
      on:click={openFolder}
      use:tt={{ content: $t('FOOTER.open_game_folder'), placement: 'top' }}
      bind:this={folderBtn}>🗁</button
    >
    <button
      class="btn"
      on:mouseenter={() => (hovered = true)}
      on:focusin={() => (focused = true)}
      on:click={launchGame}
      use:tt={{ content: $t('FOOTER.launch_game'), placement: 'top' }}
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

  .btn {
    width: 50px;
    height: 50px;
    background-color: rgba(59, 164, 117, 0.4);
    border: none;
    border-radius: 14px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
    color: white;
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
