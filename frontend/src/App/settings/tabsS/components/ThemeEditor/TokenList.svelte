<script>
  import { t } from 'svelte-i18n';
  import { tt } from '../../../../lib/tooltip';
  import { draftTheme, selectedTokenId, selectToken } from '../../../../lib/store/themeEditor';

  let query = '';

  $: allTokens = Array.isArray($draftTheme?.tokens) ? $draftTheme.tokens : [];
  $: normalizedQuery = String(query || '')
    .trim()
    .toLowerCase();
  $: visibleTokens =
    normalizedQuery === ''
      ? allTokens
      : allTokens.filter((token) => {
          const id = String(token?.id || '').toLowerCase();
          const type = String(token?.type || '').toLowerCase();
          return id.includes(normalizedQuery) || type.includes(normalizedQuery);
        });
</script>

<div class="list">
  <div class="toolbar">
    <input type="text" bind:value={query} placeholder={$t('SETTINGS.THEME_EDITOR.search_token')} />
  </div>

  {#if visibleTokens.length > 0}
    {#each visibleTokens as token}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="row {token.id === $selectedTokenId ? 'active' : ''}"
        on:click={() => selectToken(token.id)}
        use:tt={{ content: token.id }}
      >
        <span class="token-id">{token.id}</span>
        <span class="token-type">{token.type}</span>
      </div>
    {/each}
  {:else}
    <div class="empty">{$t('SETTINGS.THEME_EDITOR.no_tokens_found')}</div>
  {/if}
</div>

<style>
  .list {
    min-height: 0;
    background: var(--theme-editor-sidebar-bg, var(--bg-color, #181818));
    border-right: 1px solid var(--theme-editor-card-border, var(--border-color, #333));
    overflow-y: auto;
    overflow-x: hidden;
    overscroll-behavior: contain;
  }

  .toolbar {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 10px;
    background: var(--theme-editor-sidebar-bg, var(--bg-color, #181818));
    border-bottom: 1px solid var(--theme-editor-card-border, var(--border-color, #333));
  }

  .toolbar input {
    width: 100%;
    padding: 8px 10px;
    border-radius: 6px;
    border: 1px solid var(--theme-editor-input-border, var(--border-color, #444));
    background: var(--theme-editor-input-bg, var(--surface-base, #222));
    color: var(--text-color-primary, #fff);
    font-size: 12px;
  }

  .toolbar input:focus {
    outline: none;
    border-color: var(--theme-editor-input-border-focus, var(--action-primary-bg, #3ba475));
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 8px 12px;
    cursor: pointer;
    font-family: monospace;
  }

  .row:hover {
    background: var(--theme-editor-sidebar-row-hover-bg, var(--bg-color-medium, #222));
  }

  .row.active {
    background: var(--theme-editor-sidebar-row-active-bg, var(--tab-active-background, #2d2d2d));
    color: var(--theme-editor-sidebar-row-active-fg, var(--primary-color, #3ba475));
  }

  .token-id {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
    flex: 1;
  }

  .token-type {
    font-size: 11px;
    opacity: 0.85;
    padding: 2px 6px;
    border-radius: 999px;
    border: 1px solid var(--theme-editor-card-border, var(--border-color, #444));
    background: var(--theme-editor-badge-bg, var(--surface-elevated, #2b2b2b));
    color: var(--theme-editor-badge-text, var(--text-color-secondary, #ccc));
    flex: 0 0 auto;
  }

  .empty {
    padding: 12px;
    color: var(--theme-editor-hint-text, var(--text-color-muted, #777));
    font-size: 12px;
  }
</style>
