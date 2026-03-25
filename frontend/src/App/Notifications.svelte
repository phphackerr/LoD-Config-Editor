<script>
  import { fly } from 'svelte/transition';
  import { t } from 'svelte-i18n';
  import { dismissNotification, notifications } from './lib/store/notifications';

  function typeClass(type) {
    const normalized = String(type || '').toLowerCase();
    if (normalized === 'success') return 'success';
    if (normalized === 'error') return 'error';
    if (normalized === 'warning') return 'warning';
    return 'info';
  }
</script>

<div class="notifications" role="status" aria-live="polite" aria-atomic="false">
  {#each $notifications as item (item.id)}
    <div
      class="toast {typeClass(item.type)}"
      in:fly={{ y: 8, duration: 160 }}
      out:fly={{ y: -8, duration: 140 }}
    >
      <div class="head">
        {#if item.title}
          <strong class="title">{item.title}</strong>
        {/if}
        <button
          class="close"
          on:click={() => dismissNotification(item.id)}
          aria-label={$t('COMMON.dismiss_notification')}
        >
          ×
        </button>
      </div>
      <div class="message">{item.message}</div>
    </div>
  {/each}
</div>

<style>
  .notifications {
    position: fixed;
    top: 16px;
    right: 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    z-index: 12000;
    pointer-events: none;
    max-width: min(420px, calc(100vw - 24px));
  }

  .toast {
    pointer-events: auto;
    border-radius: 10px;
    border: 1px solid
      var(--notifications-border-color, var(--status-info-border, rgba(56, 189, 248, 0.35)));
    background: var(--notifications-bg-color, var(--surface-panel, rgba(20, 20, 20, 0.85)));
    box-shadow: var(--notifications-shadow, 0 8px 24px rgba(0, 0, 0, 0.35));
    color: var(--notifications-text-color, var(--text-color-secondary, #e2e8f0));
    padding: 10px 12px;
    backdrop-filter: blur(8px);
  }

  .toast.success {
    border-color: var(
      --notifications-success-border,
      var(--status-success-border, rgba(36, 147, 79, 0.38))
    );
    background: var(--notifications-success-bg, var(--status-success-bg, rgba(36, 147, 79, 0.18)));
    color: var(--notifications-success-text, var(--status-success-text, #9be5be));
  }

  .toast.error {
    border-color: var(
      --notifications-error-border,
      var(--status-error-border, rgba(220, 38, 38, 0.38))
    );
    background: var(--notifications-error-bg, var(--status-error-bg, rgba(220, 38, 38, 0.16)));
    color: var(--notifications-error-text, var(--status-error-text, #fecaca));
  }

  .toast.warning {
    border-color: var(
      --notifications-warning-border,
      var(--status-warning-border, rgba(245, 158, 11, 0.35))
    );
    background: var(--notifications-warning-bg, var(--status-warning-bg, rgba(245, 158, 11, 0.12)));
    color: var(--notifications-warning-text, var(--status-warning-text, #fde68a));
  }

  .toast.info {
    border-color: var(
      --notifications-info-border,
      var(--status-info-border, rgba(56, 189, 248, 0.35))
    );
    background: var(--notifications-info-bg, var(--status-info-bg, rgba(56, 189, 248, 0.14)));
    color: var(--notifications-info-text, var(--status-info-text, #bae6fd));
  }

  .head {
    display: flex;
    align-items: flex-start;
    gap: 8px;
  }

  .title {
    flex: 1;
    font-size: 12px;
    line-height: 1.3;
    color: var(--notifications-title-color, var(--text-color-primary, #f8fafc));
  }

  .close {
    border: none;
    background: transparent;
    color: var(--notifications-close-color, var(--text-color-secondary, #cbd5e1));
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    min-width: 16px;
    margin-top: -1px;
  }

  .close:hover {
    color: var(--notifications-close-hover-color, var(--text-color-primary, #fff));
  }

  .message {
    font-size: 13px;
    line-height: 1.35;
    word-wrap: break-word;
    margin-top: 4px;
  }
</style>
