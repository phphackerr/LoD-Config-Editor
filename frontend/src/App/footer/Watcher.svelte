<script>
  import { onMount, onDestroy } from 'svelte';
  import { appSettings } from '../lib/store/appSettings';
  import { startWatcher, stopWatcher, onConfigChanged } from '../lib/watcher';
  import { getCurrentConfigPath, loadConfig } from '../lib/store/config';
  import { isInternalChange } from '../lib/store/internalChange';
  import {
    CheckConfigDiff,
    ApplyChangesToConfig,
    DiscardExternalChanges
  } from '/bindings/lce/backend/config_editor/configeditor';
  import { get } from 'svelte/store';
  import { t } from 'svelte-i18n';
  import { notifyError, notifyWarning } from '../lib/store/notifications';
  import { samePath, toErrorMessage } from '../lib/store/storeUtils';

  let diff = [];
  let currentPath = null;
  let unsubscribe;
  let configListener;
  let isHovered = false;
  $: hasDiff = diff.length > 0;

  onMount(() => {
    unsubscribe = appSettings.subscribe(async (settings) => {
      const newPath = settings.game_path ? settings.game_path + '\\config.lod.ini' : null;
      if (!samePath(newPath, currentPath)) {
        if (newPath) {
          const result = await startWatcher(newPath);
          if (result.ok) {
            currentPath = newPath;
            if (Array.isArray(result.warnings)) {
              for (const warning of result.warnings) {
                notifyWarning(warning);
              }
            }
          } else {
            notifyError(result.error || $t('ERRORS.watcher.restart'));
          }
        } else if (currentPath) {
          const result = await stopWatcher();
          currentPath = newPath;
          if (!result.ok) {
            notifyError(result.error || $t('ERRORS.watcher.stop'));
          }
        }
      }
    });

    configListener = onConfigChanged(async () => {
      if (!get(isInternalChange)) {
        try {
          const nextDiff = await CheckConfigDiff();
          diff = Array.isArray(nextDiff) ? nextDiff : [];
        } catch (error) {
          notifyError(toErrorMessage(error, $t('ERRORS.watcher.check_external_changes')));
        }
      }
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
    if (configListener) configListener.off();
    if (currentPath) {
      stopWatcher();
      currentPath = null;
    }
  });

  async function acceptChanges() {
    isInternalChange.mark();
    try {
      await ApplyChangesToConfig(diff);
      const path = getCurrentConfigPath() || currentPath;
      if (path) {
        const reloadResult = await loadConfig(path);
        if (!reloadResult.ok) {
          notifyError(reloadResult.error || $t('ERRORS.watcher.reload_after_apply'));
        }
      }
    } catch (error) {
      notifyError(toErrorMessage(error, $t('ERRORS.watcher.apply_external_changes')));
    }
    diff = [];
  }

  async function discardChanges() {
    isInternalChange.mark();
    try {
      await DiscardExternalChanges();
      const path = getCurrentConfigPath() || currentPath;
      if (path) {
        const reloadResult = await loadConfig(path);
        if (!reloadResult.ok) {
          notifyError(reloadResult.error || $t('ERRORS.watcher.reload_after_discard'));
        }
      }
    } catch (error) {
      notifyError(toErrorMessage(error, $t('ERRORS.watcher.discard_external_changes')));
    }
    diff = [];
  }

  function handleMouseEnter() {
    isHovered = true;
  }

  function handleMouseLeave() {
    isHovered = false;
  }

  function renderValue(value, status, isNewValue) {
    if (status === 'added' && !isNewValue) return `<${$t('WATCHER.value_missing')}>`;
    if (status === 'deleted' && isNewValue) return `<${$t('WATCHER.value_deleted')}>`;
    if (value === '') return `<${$t('WATCHER.value_empty')}>`;
    return value;
  }
</script>

<div class="floating-panel" style="display: {$appSettings.game_path ? 'block' : 'none'};">
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="panel"
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    style="
      height: {isHovered ? (hasDiff ? '300px' : '60px') : '60px'};
      width: {isHovered ? (hasDiff ? '400px' : '280px') : '60px'};
      border-radius: {isHovered ? '5px' : '50%'};"
  >
    <div
      class="indicator {hasDiff ? 'alert' : 'ok'} {isHovered ? 'paused' : ''}"
      style="
      background-color: {hasDiff
        ? 'var(--status-warning-border, rgba(245, 158, 11, 0.35))'
        : 'var(--action-primary-bg, #3ba475)'};
      height: {hasDiff ? (isHovered ? '93%' : '40px') : '40px'};
      border-radius: {hasDiff ? (isHovered ? '5px' : '50%') : '50%'};"
    >
      <div class="icon">
        {hasDiff ? '⚠' : '✔'}
      </div>
    </div>
    <div class="content {hasDiff ? '' : 'no-changes'}" class:visible={isHovered}>
      {#if hasDiff}
        <div class="headline">{$t('WATCHER.external_changes')}:</div>
        <hr />
        <div class="diff-text">
          {#each diff as sectionDiff}
            <div class="section">
              <strong>{sectionDiff.section}</strong>
              {#if sectionDiff.status !== 'modified'}
                <div class="status {sectionDiff.status}">
                  ({$t('WATCHER.section_status')}: {sectionDiff.status})
                </div>
              {/if}
              {#if sectionDiff.keys?.length}
                <ul>
                  {#each sectionDiff.keys as info}
                    <li>
                      <span class="key-text">{info.key}</span>: {renderValue(
                        info.old,
                        info.status,
                        false
                      )} → {renderValue(info.new, info.status, true)}
                      <span class="status {info.status}">({info.status})</span>
                    </li>
                  {/each}
                </ul>
              {/if}
            </div>
          {/each}
        </div>
        <div class="footer">
          <button class="button" on:click={acceptChanges}>
            <span class="button-text">{$t('WATCHER.accept_changes')}</span>
          </button>
          <button class="button" on:click={discardChanges}>
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
    border-radius: 50%;
    box-shadow: var(--footer-watcher-panel-shadow, 0 2px 10px rgba(0, 0, 0, 0.3));
    display: flex;
    background-color: var(--footer-watcher-panel-bg, var(--surface-overlay, rgba(0, 0, 0, 0.5)));
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
    color: var(
      --footer-watcher-section-title,
      var(--accent-color, #ffd700)
    ); /* Золотой цвет для заголовков секций */
    font-size: 1.1em;
    padding-left: 5px; /* Небольшой отступ слева */
    border-left: 3px solid var(--footer-watcher-section-title-border, var(--accent-color, #ffd700)); /* Декоративная полоса слева */
  }

  .section ul {
    list-style: none;
    padding-left: 10px;
    margin: 0;
  }

  .section li {
    font-size: 0.85em;
    margin-bottom: 2px;
    background-color: var(
      --footer-watcher-section-bg,
      rgba(255, 255, 255, 0.05)
    ); /* Легкий фон для каждой строки */
    padding: 3px 5px;
    border-radius: 3px;
    margin-bottom: 5px;
  }

  .key-text {
    color: var(
      --footer-watcher-key,
      var(--status-info-text, #bae6fd)
    ); /* Светло-голубой цвет для ключей */
  }

  .added {
    color: var(--status-success-text, #9be5be);
  }

  .deleted {
    color: var(--status-error-text, #fecaca);
  }

  .modified {
    color: var(--status-warning-text, #fde68a);
  }

  .footer {
    display: flex;
    justify-content: space-around; /* Распределяем кнопки равномерно */
    width: 100%;
    padding-top: 10px;
    border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.2)); /* Небольшая разделительная линия */
    margin-top: auto;
  }

  .button {
    background-color: var(--action-secondary-bg, #444); /* Полупрозрачный фон */
    color: var(--text-color-primary, #fff);
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
    background-color: var(--action-secondary-bg-hover, #555);
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
      radial-gradient(farthest-side, var(--footer-watcher-eye-dark, #000) 15%, #0000 18%) 0 0/50%
        100%,
      radial-gradient(50% 100% at 50% 160%, var(--footer-watcher-eye-light, #fff) 95%, #0000) 0
        0 /50% 50%,
      radial-gradient(50% 100% at 50% -60%, var(--footer-watcher-eye-light, #fff) 95%, #0000) 0
        100%/50% 50%;
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
