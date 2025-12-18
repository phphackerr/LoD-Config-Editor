<script>
  import { onMount, onDestroy } from 'svelte';
  import { slide } from 'svelte/transition';
  import {
    updaterStore,
    checkForUpdates,
    doUpdate,
    updateComponent,
    restartApp
  } from './lib/store/updater';
  import { t } from 'svelte-i18n';

  let isDragging = false;
  let startX, startY;
  let initialLeft, initialTop;
  let notificationElement;
  let position = { top: '50%', left: '50%', transform: 'translate(-50%, -50%)' };

  onMount(() => {
    checkForUpdates();

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
  });

  onDestroy(() => {
    window.removeEventListener('mousemove', handleMouseMove);
    window.removeEventListener('mouseup', handleMouseUp);
  });

  function handleMouseDown(e) {
    if (e.target.closest('button') || e.target.closest('.changelog')) return;

    isDragging = true;
    startX = e.clientX;
    startY = e.clientY;

    const rect = notificationElement.getBoundingClientRect();
    initialLeft = rect.left;
    initialTop = rect.top;

    // Switch to absolute positioning for dragging
    position = {
      top: `${initialTop}px`,
      left: `${initialLeft}px`,
      transform: 'none'
    };
  }

  function handleMouseMove(e) {
    if (!isDragging) return;

    const dx = e.clientX - startX;
    const dy = e.clientY - startY;

    let newTop = initialTop + dy;
    let newLeft = initialLeft + dx;

    // Constrain to window bounds
    const rect = notificationElement.getBoundingClientRect();
    const windowWidth = window.innerWidth;
    const windowHeight = window.innerHeight;

    if (newTop < 0) newTop = 0;
    if (newLeft < 0) newLeft = 0;
    if (newTop + rect.height > windowHeight) newTop = windowHeight - rect.height;
    if (newLeft + rect.width > windowWidth) newLeft = windowWidth - rect.width;

    position = {
      top: `${newTop}px`,
      left: `${newLeft}px`,
      transform: 'none'
    };
  }

  function handleMouseUp() {
    isDragging = false;
  }

  function handleUpdate() {
    if ($updaterStore.version) {
      doUpdate($updaterStore.version);
    }
  }

  function handleRestart() {
    restartApp();
  }

  let visible = true;

  function handleClose() {
    visible = false;
  }

  function handleComponentUpdate(component) {
    updateComponent(component);
  }
</script>

{#if ($updaterStore.available || $updaterStore.componentUpdates.length > 0) && visible}
  <div
    class="updater-notification"
    transition:slide
    bind:this={notificationElement}
    on:mousedown={handleMouseDown}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    style="top: {position.top}; left: {position.left}; transform: {position.transform};"
  >
    <button class="close-btn" on:click={handleClose} aria-label="Close">×</button>
    <div class="content">
      {#if $updaterStore.available}
        <div class="info">
          <span class="title"
            >{$t('update_available', { default: 'Update Available' })}: {$updaterStore.version}</span
          >
          {#if $updaterStore.body}
            <div class="changelog">
              <div class="changelog-title">{$t('changelog', { default: 'Changelog' })}:</div>
              <pre class="changelog-text">{$updaterStore.body}</pre>
            </div>
          {/if}
          {#if $updaterStore.downloading}
            <div class="progress-bar">
              <div class="fill" style="width: {$updaterStore.progress}%"></div>
            </div>
            <span class="status">Downloading... {Math.round($updaterStore.progress)}%</span>
          {:else if $updaterStore.readyToRestart}
            <div class="success-msg">{$t('update_ready', { default: 'Download Complete!' })}</div>
            <button class="update-btn restart-btn" on:click={handleRestart}>
              {$t('restart_now', { default: 'Restart Now' })}
            </button>
          {:else}
            <button class="update-btn" on:click={handleUpdate}>
              {$t('update_now', { default: 'Update Now' })}
            </button>
          {/if}
        </div>
      {/if}

      {#if $updaterStore.componentUpdates.length > 0}
        <div class="components-list">
          <span class="title">{$t('component_updates', { default: 'Component Updates' })}</span>
          {#each $updaterStore.componentUpdates as comp}
            <div class="component-item">
              <div class="comp-info">
                <span class="comp-name">{comp.type}: {comp.name} ({comp.version})</span>
                {#if comp.changelog}
                  <span class="comp-changelog">{comp.changelog}</span>
                {/if}
              </div>
              <button class="btn-small" on:click={() => handleComponentUpdate(comp)}>
                {$t('update', { default: 'Update' })}
              </button>
            </div>
          {/each}
        </div>
      {/if}

      {#if $updaterStore.error}
        <div class="error">{$updaterStore.error}</div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .updater-notification {
    position: fixed;
    background: rgba(20, 20, 20, 0.85);
    border: 1px solid rgba(255, 215, 0, 0.5);
    border-radius: 12px;
    padding: 20px;
    z-index: 9999;
    box-shadow:
      0 20px 60px 0 rgba(0, 0, 0, 0.95),
      0 0 30px rgba(255, 215, 0, 0.4);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    width: 600px;
    max-width: 90vw;
    font-family: 'Segoe UI', sans-serif;
    color: #fff;
    cursor: move;
    user-select: none;
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .title {
    color: #ffd700;
    font-size: 16px;
    font-weight: 600;
    display: block;
    margin-bottom: 8px;
    text-shadow: 0 0 10px rgba(255, 215, 0, 0.3);
  }

  .changelog {
    cursor: default;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 6px;
    padding: 8px;
    margin-bottom: 10px;
    max-height: 150px;
    overflow-y: auto;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .changelog-title {
    font-size: 12px;
    color: #aaa;
    margin-bottom: 4px;
    font-weight: 600;
  }

  .changelog-text {
    font-size: 12px;
    color: #ddd;
    white-space: pre-wrap;
    margin: 0;
    font-family: inherit;
    line-height: 1.4;
  }

  .update-btn {
    background: linear-gradient(135deg, #e6c200 0%, #e6a800 100%);
    color: #1a1a1a;
    border: none;
    padding: 10px 20px;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 600;
    font-size: 13px;
    transition: all 0.2s ease;
    width: 100%;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
  }

  .update-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(255, 215, 0, 0.3);
    filter: brightness(1.1);
  }

  .update-btn:active {
    transform: translateY(0);
  }

  .restart-btn {
    background: linear-gradient(135deg, #4caf50 0%, #45a049 100%);
    color: white;
  }

  .restart-btn:hover {
    box-shadow: 0 4px 8px rgba(76, 175, 80, 0.3);
  }

  .success-msg {
    color: #4caf50;
    font-size: 12px;
    font-weight: bold;
    margin-bottom: 5px;
  }

  .progress-bar {
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 5px;
  }

  .fill {
    height: 100%;
    background: #ffd700;
    transition: width 0.3s ease;
  }

  .status {
    font-size: 12px;
    color: #aaa;
  }

  .error {
    color: #ff4444;
    font-size: 12px;
    margin-top: 5px;
  }

  .components-list {
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    padding-top: 10px;
    margin-top: 5px;
  }

  .component-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    margin-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    padding-bottom: 5px;
  }

  .comp-info {
    display: flex;
    flex-direction: column;
  }

  .comp-name {
    font-weight: bold;
  }

  .comp-changelog {
    font-size: 10px;
    color: #aaa;
    margin-top: 2px;
    font-style: italic;
  }

  .btn-small {
    background: #444;
    color: #fff;
    border: 1px solid #666;
    padding: 2px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 10px;
  }

  .btn-small:hover {
    background: #555;
  }

  .close-btn {
    position: absolute;
    top: 10px;
    right: 10px;
    background: transparent;
    border: none;
    color: #aaa;
    font-size: 24px;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
    z-index: 10;
  }

  .close-btn:hover {
    color: #fff;
  }
</style>
