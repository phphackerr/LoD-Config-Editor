<script module>
  export const tabMetadata = {
    order: 5
  };
</script>

<script>
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { GetAppVersion, OpenURL, GetDiscordStats } from '/bindings/lce/backend/utils/utils';
  import { GetLanguages } from '/bindings/lce/backend/i18n/i18n';
  import { checkForUpdates, updaterStore } from '../../lib/store/updater';
  import { cubicOut } from 'svelte/easing';

  let appVersion = '...';
  let discordStats = null;
  let languages = [];
  let checked = false;

  onMount(async () => {
    try {
      appVersion = await GetAppVersion();
    } catch (e) {
      console.error('Failed to get app version:', e);
      appVersion = 'Unknown';
    }

    try {
      discordStats = await GetDiscordStats('d35eBUs8P5');
    } catch (e) {
      console.error('Failed to get discord stats:', e);
    }

    try {
      languages = await GetLanguages();
    } catch (e) {
      console.error('Failed to get languages:', e);
    }
  });

  const openLink = (url) => {
    OpenURL(url);
  };

  const handleCheckUpdate = async () => {
    checked = false;
    await checkForUpdates();
    checked = true;
    // Reset checked status after 1.5 seconds
    setTimeout(() => {
      checked = false;
    }, 1500);
  };

  function slideHorizontal(node, { delay = 0, duration = 400, easing = cubicOut }) {
    const style = getComputedStyle(node);
    const width = parseFloat(style.width);
    const marginLeft = parseFloat(style.marginLeft);

    return {
      delay,
      duration,
      easing,
      css: (t) => `
        overflow: hidden;
        width: ${t * width}px;
        margin-left: ${t * marginLeft}px;
        opacity: ${t};
        white-space: nowrap;
      `
    };
  }
</script>

<div class="about-container">
  <div class="header">
    <h1>{$t('SETTINGS.ABOUT.app_name')}</h1>
    <div class="version-wrapper">
      <span class="version">{$t('SETTINGS.ABOUT.version')} {appVersion}</span>
      <button
        class="refresh-btn"
        on:click={handleCheckUpdate}
        disabled={$updaterStore.checking}
        title={$t('SETTINGS.ABOUT.check_for_updates')}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class:spin={$updaterStore.checking}
        >
          <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
          <path d="M3 3v5h5" />
          <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
          <path d="M16 16h5v5" />
        </svg>
      </button>
      {#if checked && !$updaterStore.available}
        <span class="status-text" transition:slideHorizontal>{$t('SETTINGS.ABOUT.up_to_date')}</span
        >
      {/if}
    </div>
  </div>

  <p class="description">
    {$t('SETTINGS.ABOUT.description')}
  </p>

  <div class="links">
    <button class="link-btn discord" on:click={() => openLink('https://discord.gg/d35eBUs8P5')}>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="currentColor"
        ><path
          d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.085 2.157 2.419 0 1.334-.956 2.42-2.157 2.42zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.085 2.157 2.419 0 1.334-.946 2.42-2.157 2.42z"
        /></svg
      >
      <div class="btn-content">
        <span>{$t('SETTINGS.ABOUT.discord')}</span>
        {#if discordStats}
          <div class="stats-row">
            <span class="online-count"
              >● {discordStats.approximate_presence_count} {$t('SETTINGS.ABOUT.online')}</span
            >
            <span class="total-count"
              >/ {discordStats.approximate_member_count} {$t('SETTINGS.ABOUT.total')}</span
            >
          </div>
        {/if}
      </div>
    </button>

    <button
      class="link-btn github"
      on:click={() => openLink('https://github.com/phphackerr/LoD-Config-Editor')}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="currentColor"
        ><path
          d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
        /></svg
      >
      {$t('SETTINGS.ABOUT.github')}
    </button>
  </div>

  <div class="credits">
    <div class="credit-section">
      <h3>{$t('SETTINGS.ABOUT.developer')}</h3>
      <p>phphacker</p>
    </div>

    <div class="credit-section">
      <h3>{$t('SETTINGS.ABOUT.translators')}</h3>
      {#each languages as lang}
        {#if lang.author}
          <p>{lang.name}: {lang.author}</p>
        {/if}
      {/each}
    </div>

    <div class="credit-section">
      <h3>{$t('SETTINGS.ABOUT.thanks')}</h3>
      <p>{$t('SETTINGS.ABOUT.thanksD')}</p>
    </div>
  </div>
</div>

<style>
  .about-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 40px 20px;
    color: var(--text-color);
    text-align: center;
    gap: 30px;
    max-width: 800px;
    margin: 0 auto;
  }

  .header h1 {
    font-size: 2.5rem;
    margin-bottom: 5px;
    background: linear-gradient(
      45deg,
      #ff6b6b,
      #f06595,
      #cc5de8,
      #845ef7,
      #5c7cfa,
      #339af0,
      #22b8cf,
      #20c997,
      #51cf66,
      #94d82d,
      #fcc419,
      #ff922b
    );
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    background-size: 300% 300%;
    animation: rainbow 10s ease infinite;
  }

  @keyframes rainbow {
    0% {
      background-position: 0% 50%;
    }
    50% {
      background-position: 100% 50%;
    }
    100% {
      background-position: 0% 50%;
    }
  }

  .version {
    font-size: 1rem;
    opacity: 0.7;
    font-family: monospace;
    margin-right: 10px;
  }

  .description {
    font-size: 1.1rem;
    line-height: 1.6;
    max-width: 600px;
    opacity: 0.9;
  }

  .links {
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
    justify-content: center;
  }

  .link-btn {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition:
      transform 0.2s,
      box-shadow 0.2s;
    color: white;
  }

  .link-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
  }

  .link-btn:active {
    transform: translateY(0);
  }

  .link-btn.discord {
    background-color: #5865f2;
  }

  .link-btn.github {
    background-color: #333;
  }

  .version-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 5px;
  }

  .refresh-btn {
    background: transparent;
    border: none;
    color: var(--text-color);
    opacity: 0.5;
    cursor: pointer;
    padding: 4px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .refresh-btn:hover:not(:disabled) {
    opacity: 1;
    background: rgba(255, 255, 255, 0.1);
  }

  .refresh-btn:disabled {
    cursor: wait;
    opacity: 0.8;
  }

  .status-text {
    font-size: 0.9rem;
    color: #4caf50;
    margin-left: 10px;
  }

  .spin {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .btn-content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    line-height: 1.2;
  }

  .online-count {
    font-size: 0.75rem;
    opacity: 0.8;
    color: #09f189;
    font-weight: 700;
  }

  .total-count {
    font-size: 0.75rem;
    opacity: 0.6;
    margin-left: 4px;
  }

  .stats-row {
    display: flex;
    align-items: center;
  }

  .credits {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 30px;
    width: 100%;
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    text-align: left;
  }

  .credit-section h3 {
    font-size: 1.1rem;
    margin-bottom: 15px;
    color: var(--accent-color, #646cff);
  }

  .credit-section p {
    margin: 5px 0;
    font-size: 0.95rem;
    opacity: 0.8;
  }
</style>
