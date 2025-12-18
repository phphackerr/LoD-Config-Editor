<script>
  //@ts-nochek
  import { onMount, onDestroy } from 'svelte';
  import { appSettings } from '../lib/store/appSettings';
  import { startWatcher, stopWatcher, onConfigChanged } from '../lib/watcher';
  import { isInternalChange } from '../lib/store/internalChange';
  import {
    CheckConfigDiff,
    ApplyChangesToConfig,
    DiscardExternalChanges
  } from '/bindings/lce/backend/config_editor/configeditor';
  import { get } from 'svelte/store';
  import { t } from 'svelte-i18n';

  let diff = {};
  let currentPath = null;
  let changedFile = '';
  let unsubscribe;
  let configListener;
  let isHovered = false;

  onMount(() => {
    unsubscribe = appSettings.subscribe(async (settings) => {
      const newPath = settings.game_path ? settings.game_path + '\\config.lod.ini' : null;
      if (newPath !== currentPath) {
        if (newPath) {
          try {
            await startWatcher(newPath);
            currentPath = newPath;
          } catch (error) {
            console.error('Ошибка при перезапуске вотчера:', error);
            // Опционально: можно добавить логику для уведомления пользователя
          }
        } else if (currentPath) {
          try {
            await stopWatcher();
            currentPath = newPath;
          } catch (error) {
            console.error('Ошибка при остановке вотчера:', error);
          }
        }
      }
    });

    configListener = onConfigChanged(async (filePath) => {
      if (!get(isInternalChange)) {
        changedFile = filePath;
        diff = await CheckConfigDiff();
      }
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
    if (configListener) configListener.off();
  });

  async function acceptChanges() {
    isInternalChange.mark();
    try {
      await ApplyChangesToConfig(diff);
    } catch (error) {
      console.error('[Watcher.svelte] Ошибка при применении изменений:', error);
    }
    diff = {};
    changedFile = '';
  }

  async function disacrdChanges() {
    isInternalChange.mark();
    try {
      await DiscardExternalChanges();
    } catch (error) {
      console.error('[Watcher.svelte] Ошибка при отмене изменений:', error);
    }
    diff = {};
    changedFile = '';
  }

  function handleMouseEnter() {
    isHovered = true;
  }

  function handleMouseLeave() {
    isHovered = false;
  }
</script>

<div class="floating-panel" style="display: {$appSettings.game_path ? 'block' : 'none'};">
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="panel"
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    style="
      height: {isHovered ? (Object.keys(diff).length ? '300px' : '60px') : '60px'};
      width: {isHovered ? (Object.keys(diff).length ? '400px' : '280px') : '60px'};
      border-radius: {isHovered ? '5px' : '50%'};"
  >
    <div
      class="indicator {Object.keys(diff).length ? 'alert' : 'ok'} {isHovered ? 'paused' : ''}"
      style="
      background-color: {Object.keys(diff).length ? 'orange' : 'rgba(59, 164, 117, 0.4)'};
      height: {Object.keys(diff).length ? (isHovered ? '93%' : '40px') : '40px'};
      border-radius: {Object.keys(diff).length ? (isHovered ? '5px' : '50%') : '50%'};"
    >
      <div class="icon">
        {Object.keys(diff).length ? '⚠' : '✔'}
      </div>
    </div>
    <div class="content {Object.keys(diff).length ? '' : 'no-changes'}" class:visible={isHovered}>
      {#if Object.keys(diff).length}
        <div class="headline">{$t('WATCHER.external_changes')}:</div>
        <hr />
        <div class="diff-text">
          {#each Object.entries(diff) as [section, keys]}
            <div class="section">
              <strong>{section}</strong>
              <ul>
                {#each Object.entries(keys) as [key, info]}
                  <li>
                    <span class="key-text">{key}</span>: {info.old ?? '<deleted>'} → {info.new ??
                      '<deleted>'}
                    <span
                      class="status {info.status == 'added'
                        ? 'added'
                        : info.status == 'deleted'
                          ? 'deleted'
                          : 'modified'}">({info.status})</span
                    >
                  </li>
                {/each}
              </ul>
            </div>
          {/each}
        </div>
        <div class="footer">
          <button class="button" on:click={acceptChanges}>
            <span class="button-text">{$t('WATCHER.accept_changes')}</span>
          </button>
          <button class="button" on:click={disacrdChanges}>
            <span class="button-text">{$t('WATCHER.discard_changes')}</span>
          </button>
        </div>
      {:else}
        <div class="eyes"></div>
        <span>{$t('WATCHER.watcher_working')}</span>
      {/if}
    </div>
  </div>
</div>

<style>
  .floating-panel {
    height: fit-content;
    max-height: 300px;
    max-width: 400px;
    padding: 10px;
    border-radius: 5px;
    position: fixed;
    bottom: 5px;
    left: 5px;
    z-index: 999;
  }

  .panel {
    width: 60px;
    height: 60px;
    border-radius: 5px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
    display: flex;
    background-color: rgba(0, 0, 0, 0.5);
    flex-direction: row;
    gap: 10px;
    transition: all 0.3s ease;
  }

  .indicator {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    align-self: flex-end;
    border-radius: 50%;
    margin-left: 10px;
    margin-bottom: 10px;
    transition: all 0.3s ease;
  }

  .icon {
    font-size: 22px;
  }

  .content {
    opacity: 0;
    visibility: hidden;
    max-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    align-items: center;
    flex-grow: 1;
    transition:
      opacity 0.3s ease,
      visibility 0.3s ease,
      max-height 0.3s ease,
      border-color 0.3s ease;
  }

  .content.visible {
    opacity: 1;
    visibility: visible;
    max-height: 100%;
    padding: 5px;
  }

  .diff-text {
    overflow-y: auto;
  }

  .content hr {
    width: 90%;
  }

  .section {
    margin-bottom: 10px;
    overflow-y: auto;
  }

  .section strong {
    display: block;
    margin-bottom: 3px;
    color: #ffd700; /* Золотой цвет для заголовков секций */
    font-size: 1.1em;
    padding-left: 5px; /* Небольшой отступ слева */
    border-left: 3px solid #ffd700; /* Декоративная полоса слева */
  }

  .section ul {
    list-style: none;
    padding-left: 10px;
    margin: 0;
  }

  .section li {
    font-size: 0.85em;
    margin-bottom: 2px;
    background-color: rgba(255, 255, 255, 0.05); /* Легкий фон для каждой строки */
    padding: 3px 5px;
    border-radius: 3px;
    margin-bottom: 5px;
  }

  .key-text {
    color: #add8e6; /* Светло-голубой цвет для ключей */
  }

  .added {
    color: rgb(31, 194, 31);
  }

  .deleted {
    color: rgb(194, 31, 31);
  }

  .modified {
    color: rgb(194, 194, 31);
  }

  .footer {
    display: flex;
    justify-content: space-around; /* Распределяем кнопки равномерно */
    width: 100%;
    padding-top: 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.2); /* Небольшая разделительная линия */
    margin-top: auto;
  }

  .button {
    background-color: rgba(255, 255, 255, 0.1); /* Полупрозрачный фон */
    color: white;
    padding: 8px 15px;
    border-radius: 5px;
    border: none;
    cursor: pointer;
    font-size: 0.9em;
    transition:
      background-color 0.2s ease,
      transform 0.1s ease; /* Плавный переход для hover */
    text-align: center; /* Центрируем текст внутри кнопки */
  }

  .button:hover {
    background-color: rgba(255, 255, 255, 0.2);
    transform: translateY(-1px); /* Небольшое поднятие кнопки */
  }

  .button:active {
    transform: translateY(0); /* Возвращаем на место при клике */
  }

  .indicator.alert {
    animation: pulse 1s infinite;
  }

  .indicator.paused {
    animation: none;
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.1);
    }
  }

  .no-changes {
    flex-direction: row;
    gap: 10px;
    align-items: center;
    justify-content: center;
  }

  .eyes {
    height: 30px;
    aspect-ratio: 2;
    display: grid;
    background:
      radial-gradient(farthest-side, #000 15%, #0000 18%) 0 0/50% 100%,
      radial-gradient(50% 100% at 50% 160%, #fff 95%, #0000) 0 0 /50% 50%,
      radial-gradient(50% 100% at 50% -60%, #fff 95%, #0000) 0 100%/50% 50%;
    background-repeat: repeat-x;
    animation: l2 1.5s infinite linear;
  }
  @keyframes l2 {
    0%,
    15% {
      background-position:
        0 0,
        0 0,
        0 100%;
    }
    20%,
    40% {
      background-position:
        5px 0,
        0 0,
        0 100%;
    }
    45%,
    55% {
      background-position:
        0 0,
        0 0,
        0 100%;
    }
    60%,
    80% {
      background-position:
        -5px 0,
        0 0,
        0 100%;
    }
    85%,
    100% {
      background-position:
        0 0,
        0 0,
        0 100%;
    }
  }
</style>
