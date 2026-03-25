<script module>
  export const tabMetadata = {
    order: 6
  };
</script>

<script>
  import {
    FetchMapInfo,
    GetChangelog,
    DownloadMap,
    PauseDownload,
    ResumeDownload,
    StopDownload
  } from '/bindings/lce/backend/map_downloader/mapdownloader';
  import {
    SetTaskbarProgress,
    SetTaskbarError,
    SetTaskbarCompleteAndFlash
  } from '/bindings/lce/backend/taskbar/taskbarutils';
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { t } from 'svelte-i18n';
  import { tt } from '../lib/tooltip';
  import { appSettings } from '../lib/store/appSettings';
  import { toErrorMessage } from '../lib/store/storeUtils';

  let mapInfo = null;
  let loading = false;
  let error = null;
  let downloadProgress = 0;
  let downloadSpeed = 0;
  let downloadedBytes = 0;
  let totalBytes = 0;
  let isDownloading = false;
  let isPaused = false;
  let isStopping = false;
  let stopRequested = false;
  let changelogData = [];
  let lastLoadedGamePath = '';
  let activeLoadId = 0;

  // Реактивно загружаем информацию о карте при изменении пути к игре
  $: {
    const gamePath = $appSettings.game_path || '';
    if (!gamePath) {
      lastLoadedGamePath = '';
      mapInfo = null;
      changelogData = [];
    } else if (gamePath !== lastLoadedGamePath) {
      lastLoadedGamePath = gamePath;
      loadMapInfo();
    }
  }

  async function loadMapInfo() {
    const loadID = ++activeLoadId;
    try {
      loading = true;
      error = null;
      // Используем локальную переменную для проверки актуальности
      const currentPath = $appSettings.game_path;

      const result = await FetchMapInfo();

      // Если путь изменился или стал пустым пока мы грузили - игнорируем результат
      if (
        loadID !== activeLoadId ||
        $appSettings.game_path !== currentPath ||
        !$appSettings.game_path
      ) {
        return;
      }

      mapInfo = result;
      if (mapInfo) {
        const rawChangelogHtml = await GetChangelog(mapInfo.version);
        changelogData = parseHTMLToData(rawChangelogHtml);
      }
    } catch (e) {
      // Тоже проверяем актуальность
      if (loadID !== activeLoadId || !$appSettings.game_path) return;
      error = toErrorMessage(e, $t('ERRORS.useful.load_map_info'));
      mapInfo = null;
    } finally {
      // Если путь пустой, loading должен быть false, но mapInfo null (уже обработано в else)
      if ($appSettings.game_path) {
        loading = false;
      } else {
        loading = false; // Все равно сбрасываем loading
      }
    }
  }

  function normalizeChangelogText(value) {
    return String(value || '')
      .replace(/\u00a0/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }

  function parseSkillBlocks(skillBlocks) {
    const parseSingleSkillBlock = (skillBlock) => {
      const parsed = [];

      const pushParagraph = (text) => {
        if (!text) return;
        parsed.push({ type: 'paragraph', text });
      };

      const pushSpacer = () => {
        if (parsed.length === 0) return;
        const last = parsed[parsed.length - 1];
        if (last?.type === 'spacer') return;
        parsed.push({ type: 'spacer' });
      };

      const pushList = (listElement) => {
        const items = Array.from(listElement.querySelectorAll(':scope > li'))
          .map((li) => normalizeChangelogText(li.textContent))
          .filter(Boolean);
        if (items.length > 0) {
          parsed.push({ type: 'list', items });
        }
      };

      const children = Array.from(skillBlock.children);

      if (children.length === 0) {
        const text = normalizeChangelogText(skillBlock.textContent);
        if (text) {
          pushParagraph(text);
        } else if (skillBlock.querySelector('br')) {
          pushSpacer();
        }
      } else {
        for (const child of children) {
          const tag = child.tagName;

          if (tag === 'P') {
            const text = normalizeChangelogText(child.textContent);
            if (text) {
              pushParagraph(text);
            } else if (child.querySelector('br')) {
              pushSpacer();
            }
            continue;
          }

          if (tag === 'UL' || tag === 'OL') {
            pushList(child);
            continue;
          }

          if (tag === 'PRE') {
            const code = normalizeChangelogText(child.textContent);
            if (code) parsed.push({ type: 'code', text: code });
            continue;
          }

          if (child.classList.contains('heroSkillIconBlock')) {
            const titleNode = child.querySelector('h1,h2,h3,h4,h5,h6');
            const subtitle = normalizeChangelogText(titleNode?.textContent || child.textContent);
            if (subtitle) parsed.push({ type: 'subtitle', text: subtitle });
            continue;
          }

          if (/^H[1-6]$/.test(tag)) {
            const subtitle = normalizeChangelogText(child.textContent);
            if (subtitle) parsed.push({ type: 'subtitle', text: subtitle });
            continue;
          }

          const text = normalizeChangelogText(child.textContent);
          if (text) {
            pushParagraph(text);
          } else if (child.querySelector('br')) {
            pushSpacer();
          }
        }
      }

      while (parsed.length > 0 && parsed[parsed.length - 1]?.type === 'spacer') {
        parsed.pop();
      }

      return parsed;
    };

    const blocks = [];
    for (const skillBlock of skillBlocks) {
      if (!(skillBlock instanceof Element)) continue;

      const parsedBlock = parseSingleSkillBlock(skillBlock);
      if (parsedBlock.length === 0) continue;

      if (blocks.length > 0 && blocks[blocks.length - 1]?.type !== 'spacer') {
        blocks.push({ type: 'spacer' });
      }

      blocks.push(...parsedBlock);
    }

    while (blocks.length > 0 && blocks[blocks.length - 1]?.type === 'spacer') {
      blocks.pop();
    }

    return blocks;
  }

  function parseVersionSections(container) {
    const sections = [];
    let currentSection = null;

    const heroBlocks = Array.from(container.querySelectorAll(':scope > .heroBlock'));
    for (const block of heroBlocks) {
      const title = normalizeChangelogText(
        block.querySelector(':scope > .heroIconBlock h3')?.textContent
      );
      const skillBlocks = Array.from(block.querySelectorAll(':scope > .heroSkillBlock'));

      if (block.classList.contains('generalBlock') && title) {
        currentSection = { title, groups: [] };
        sections.push(currentSection);
        continue;
      }

      const group = {
        title,
        blocks: parseSkillBlocks(skillBlocks)
      };

      if (!currentSection) {
        currentSection = { title: 'General', groups: [] };
        sections.push(currentSection);
      }

      if (group.title || group.blocks.length > 0) {
        currentSection.groups.push(group);
      }
    }

    return sections;
  }

  function parseModernChangelog(doc) {
    const textRoot = doc.querySelector('div.text') || doc.body;
    if (!textRoot) return [];

    const articles = Array.from(textRoot.querySelectorAll(':scope > article'));
    if (articles.length === 0) return [];

    return articles
      .map((article, index) => {
        const version = normalizeChangelogText(article.querySelector(':scope > h2')?.textContent);
        if (!version) return null;

        const contentRoot = article.querySelector(':scope > .version-content') || article;
        return {
          version,
          isExpanded: index === 0,
          sections: parseVersionSections(contentRoot),
          rawHTML: ''
        };
      })
      .filter(Boolean);
  }

  function parseLegacyChangelog(doc) {
    const sections = Array.from(doc.body.children);
    const versions = [];
    let currentVersion = null;
    let currentContent = [];

    sections.forEach((section) => {
      if (
        section.tagName === 'P' &&
        section.getAttribute('style')?.includes('text-align: center')
      ) {
        const versionText = normalizeChangelogText(section.textContent);
        if (versionText) {
          if (currentVersion) {
            versions.push({
              ...currentVersion,
              rawHTML: currentContent.join('').trim(),
              sections: []
            });
          }
          currentVersion = {
            version: versionText,
            isExpanded: versions.length === 0
          };
          currentContent = [];
        }
      } else if (currentVersion) {
        currentContent.push(section.outerHTML);
      }
    });

    if (currentVersion) {
      versions.push({
        ...currentVersion,
        rawHTML: currentContent.join('').trim(),
        sections: []
      });
    }

    return versions;
  }

  function parseHTMLToData(html) {
    if (!html) return [];

    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');

    const modern = parseModernChangelog(doc);
    if (modern.length > 0) {
      return modern;
    }

    return parseLegacyChangelog(doc);
  }

  function handleVersionClick(version) {
    changelogData = changelogData.map((item) => {
      if (item.version === version) {
        return { ...item, isExpanded: !item.isExpanded };
      }
      return item;
    });
  }

  async function downloadMap() {
    if (!mapInfo || isDownloading) return;
    let unlisten = null;
    try {
      loading = true;
      isDownloading = true;
      isPaused = false;
      isStopping = false;
      stopRequested = false;
      error = null;
      downloadProgress = 0;
      downloadSpeed = 0;
      downloadedBytes = 0;
      totalBytes = mapInfo.size || 0;
      const mapToDownload = mapInfo;

      unlisten = await Events.On('download-progress', (event) => {
        downloadProgress = event.data.progress;
        downloadedBytes = event.data.downloaded;
        totalBytes = event.data.total;
        downloadSpeed = event.data.speed;
        SetTaskbarProgress(downloadedBytes, totalBytes);
      });

      const result = await DownloadMap(mapToDownload);
      mapInfo = { ...mapToDownload, ...result };

      SetTaskbarCompleteAndFlash();
    } catch (e) {
      const message = toErrorMessage(e, $t('ERRORS.useful.download_map'));
      if (
        stopRequested ||
        message.toLowerCase().includes('остановлена пользователем') ||
        message.toLowerCase().includes('stopped by user')
      ) {
        error = null;
      } else {
        error = message;
        SetTaskbarError();
      }
      downloadProgress = 0;
      downloadSpeed = 0;
    } finally {
      if (typeof unlisten === 'function') {
        unlisten();
      }
      loading = false;
      isDownloading = false;
      isPaused = false;
      isStopping = false;
      stopRequested = false;
      SetTaskbarProgress(0, 0);
    }
  }

  async function togglePauseDownload() {
    if (!isDownloading || isStopping) return;

    try {
      if (isPaused) {
        await ResumeDownload();
        isPaused = false;
      } else {
        await PauseDownload();
        isPaused = true;
      }
    } catch (e) {
      error = toErrorMessage(e, $t('ERRORS.useful.download_map'));
    }
  }

  async function stopDownload() {
    if (!isDownloading || isStopping) return;

    isStopping = true;
    stopRequested = true;
    isPaused = false;

    try {
      await StopDownload();
    } catch (e) {
      error = toErrorMessage(e, $t('ERRORS.useful.download_map'));
      isStopping = false;
      stopRequested = false;
    }
  }

  function formatSize(bytes) {
    if (bytes === 0 || bytes === undefined) return $t('USEFUL.unknown');
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  }

  function formatSpeed(bytesPerSecond) {
    if (bytesPerSecond === 0) return '0 B/s';
    const k = 1024;
    const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    const i = Math.floor(Math.log(bytesPerSecond) / Math.log(k));
    return `${parseFloat((bytesPerSecond / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  }

  function initTooltip(node) {
    if (node && mapInfo?.save_path) {
      tt(node, {
        content: `<div style="white-space: nowrap;">${mapInfo.save_path}</div>`,
        placement: 'top',
        maxWidth: 500
      });
    }
  }

  function formatVersion(version) {
    // Извлекаем только цифры и точки из версии
    const match = version.match(/\d+\.\d+/);
    return match ? match[0] : version;
  }

  onDestroy(() => {
    activeLoadId += 1;
  });
</script>

<div class="tab-page">
  <div class="content-wrapper">
    <!-- Левая колонка - загрузка карты -->
    <div class="map-downloader">
      <h3>{$t('USEFUL.download_latest')}</h3>

      {#if error}
        <div class="error">{$t('USEFUL.error')}: {error}</div>
      {/if}

      {#if mapInfo}
        <div class="map-info">
          <div class="info-row">
            <span class="label">{$t('USEFUL.save_path')}:</span>
            <span class="value save-path" use:initTooltip>
              {mapInfo.save_path || $t('USEFUL.unknown')}
            </span>
          </div>
          <div class="info-row">
            <span class="label">{$t('USEFUL.map_version')}:</span>
            <span class="value version">{mapInfo.version}</span>
          </div>
          <div class="info-row">
            <span class="label">{$t('USEFUL.map_date')}:</span>
            <span class="value">{mapInfo.date}</span>
          </div>
          <div class="info-row">
            <span class="label">{$t('USEFUL.map_size')}:</span>
            <span class="value">
              {#if mapInfo.size !== undefined && mapInfo.size !== null}
                {formatSize(mapInfo.size)}
              {:else}
                {$t('USEFUL.unknown')}
              {/if}
            </span>
          </div>
          <div class="info-row">
            <span class="label">{$t('USEFUL.map_status')}:</span>
            <span class="value status" class:downloaded={mapInfo.is_downloaded}>
              {mapInfo.is_downloaded ? $t('USEFUL.downloaded') : $t('USEFUL.not_downloaded')}
            </span>
          </div>
        </div>

        <div
          class="download-progress-slot"
          class:active={isDownloading && (downloadProgress > 0 || isPaused || isStopping)}
        >
          <div class="progress-bar-container">
            <div
              class="progress-bar"
              style="width: {Math.max(0, Math.min(100, downloadProgress))}%;"
            ></div>
          </div>
          <div class="download-info">
            <span class="download-size">
              {formatSize(downloadedBytes)} / {formatSize(totalBytes)}
            </span>
            <span class="download-speed">
              {#if isStopping}
                {$t('USEFUL.stopping')}...
              {:else if isPaused}
                {$t('USEFUL.paused')}
              {:else}
                {formatSpeed(downloadSpeed)}
              {/if}
            </span>
          </div>
        </div>

        <button on:click={downloadMap} disabled={loading} class="download-btn">
          {#if loading}
            {#if downloadProgress > 0 && downloadProgress < 100}
              {$t('USEFUL.downloading')} ({downloadProgress.toFixed(0)}%)
            {:else}
              {$t('USEFUL.loading')}...
            {/if}
          {:else}
            {$t('USEFUL.download')}
          {/if}
        </button>

        {#if isDownloading}
          <div class="download-controls">
            <button
              class="download-control-btn pause-btn"
              on:click={togglePauseDownload}
              disabled={isStopping}
            >
              {isPaused ? $t('USEFUL.resume') : $t('USEFUL.pause')}
            </button>
            <button
              class="download-control-btn stop-btn"
              on:click={stopDownload}
              disabled={isStopping}
            >
              {#if isStopping}
                {$t('USEFUL.stopping')}...
              {:else}
                {$t('USEFUL.stop')}
              {/if}
            </button>
          </div>
        {/if}
      {:else if !loading}
        <div class="no-map">
          {#if !$appSettings.game_path}
            <p>{$t('SETTINGS.PATHS.paths_not_found')}</p>
          {:else}
            <p>{$t('USEFUL.map_not_found')}</p>
          {/if}
        </div>
      {/if}

      {#if loading && !mapInfo && !error}
        <div class="loading">
          {$t('USEFUL.loading_map_info')}...
        </div>
      {/if}
    </div>

    <!-- Правая колонка - чейнджлог -->
    <div class="changelog-panel">
      <div class="changelog-header">
        <h3>{$t('USEFUL.changelog')}:</h3>
        {#if mapInfo}
          <span class="version-badge">{formatVersion(mapInfo.version)}</span>
        {/if}
      </div>
      {#if changelogData.length > 0}
        <div class="changelog-content">
          {#each changelogData as item}
            <div class="version-section">
              <button class="version-header" on:click={() => handleVersionClick(item.version)}>
                {item.version}
                <span class="toggle-icon">{item.isExpanded ? '▼' : '▶'}</span>
              </button>
              {#if item.isExpanded}
                <div class="version-content">
                  {#if item.sections && item.sections.length > 0}
                    {#each item.sections as section}
                      <section class="cl-section">
                        {#if section.title}
                          <h4 class="cl-section-title">{section.title}</h4>
                        {/if}

                        {#each section.groups || [] as group}
                          <div class="cl-group">
                            {#if group.title}
                              <h5 class="cl-group-title">{group.title}</h5>
                            {/if}

                            {#each group.blocks || [] as block}
                              {#if block.type === 'paragraph'}
                                <p class="cl-paragraph">{block.text}</p>
                              {:else if block.type === 'list'}
                                <ul class="cl-list">
                                  {#each block.items || [] as listItem}
                                    <li>{listItem}</li>
                                  {/each}
                                </ul>
                              {:else if block.type === 'code'}
                                <pre class="cl-code"><code>{block.text}</code></pre>
                              {:else if block.type === 'subtitle'}
                                <h6 class="cl-subtitle">{block.text}</h6>
                              {:else if block.type === 'spacer'}
                                <div class="cl-spacer" aria-hidden="true"></div>
                              {/if}
                            {/each}
                          </div>
                        {/each}
                      </section>
                    {/each}
                  {:else if item.rawHTML}
                    <div class="legacy-content">
                      {@html item.rawHTML}
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {:else}
        <div class="no-changelog">
          <p>{$t('USEFUL.loading_changelog')}...</p>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .tab-page {
    height: calc(100% - 15px);
    padding: 2rem;
    box-sizing: border-box;
    display: flex;
  }

  .content-wrapper {
    display: flex;
    gap: 2rem;
    flex: 1;
    max-width: 1200px;
    margin: 0 auto;
    min-height: 0;
  }

  .map-downloader {
    background: var(--useful-panel-bg, var(--card-bg));
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 1.5rem;
    box-shadow: var(--useful-panel-shadow, 0 2px 4px rgba(0, 0, 0, 0.1));
    min-height: 0;
    flex: 1;
    max-width: 400px;
  }

  .map-downloader > h3 {
    text-align: center;
    margin: 0;
  }

  .changelog-panel {
    flex: 1;
    background: var(--useful-panel-bg, var(--card-bg));
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    box-shadow: var(--useful-panel-shadow, 0 2px 4px rgba(0, 0, 0, 0.1));
    min-height: 0;
  }

  .changelog-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
  }

  .version-badge {
    background: var(--useful-badge-bg, var(--primary-color));
    color: var(--useful-badge-text, #fff);
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 1.1em;
    font-weight: 500;
  }

  .map-info {
    background: var(--bg-color);
    padding: 1rem;
    border-radius: 6px;
    border: 1px solid var(--border-color);
    width: 100%;
    margin: 0;
    align-self: stretch;
    overflow-y: auto;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border-color);
    word-break: break-word;
    min-width: 0;
    gap: 0.5rem;
  }

  .info-row:last-child {
    border-bottom: none;
  }

  .label {
    color: var(--text-secondary);
    flex-shrink: 0;
    min-width: 100px;
  }

  .value {
    font-weight: 500;
    text-align: right;
    flex: 1;
    min-width: 0;
  }

  .save-path {
    font-size: 0.9em;
    width: 100%;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: help;
  }

  .version {
    color: var(--primary-color);
  }

  .download-btn {
    width: 100%;
    height: 40px;
    background-color: var(--useful-download-bg, #3ba475);
    color: var(--useful-download-text, #fff);
    border: none;
    border-radius: 6px;
    font-weight: 500;
    font-size: 1rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin: 0;
  }

  .download-controls {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .download-control-btn {
    width: 100%;
    height: 36px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    color: var(--text-color);
    cursor: pointer;
    background: var(--bg-color);
  }

  .download-control-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .pause-btn:hover:not(:disabled) {
    background: var(--useful-control-pause-bg-hover, rgba(255, 255, 255, 0.06));
  }

  .stop-btn {
    border-color: var(--error-color);
    color: var(--error-color);
  }

  .stop-btn:hover:not(:disabled) {
    background: var(--useful-control-stop-bg-hover, rgba(255, 82, 82, 0.12));
  }

  .download-progress-slot {
    min-height: 44px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    opacity: 0;
    visibility: hidden;
    transition: opacity 0.15s ease;
  }

  .download-progress-slot.active {
    opacity: 1;
    visibility: visible;
  }

  .download-btn:hover:not(:disabled) {
    background-color: var(--useful-download-bg-hover, #2d8c5f);
  }

  .download-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .download-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.9em;
    color: var(--text-secondary);
    padding: 0 0.5rem;
  }

  .download-size,
  .download-speed {
    font-weight: 500;
    color: var(--text-color);
  }

  .status.downloaded {
    color: var(--primary-color);
  }

  .status:not(.downloaded) {
    color: var(--text-secondary);
  }

  .progress-bar-container {
    width: 100%;
    height: 10px;
    background-color: var(--bg-color);
    border-radius: 5px;
    overflow: hidden;
    margin-top: 0.25rem;
  }

  .progress-bar {
    height: 100%;
    background-color: var(--primary-color);
    width: 0%;
    transition: width 0.1s ease-in-out;
  }

  .error {
    color: var(--error-color);
    padding: 1rem;
    background: var(--error-bg);
    border-radius: 6px;
    text-align: center;
    user-select: text;
  }

  .no-map,
  .loading {
    text-align: center;
    padding: 2rem;
    background: var(--bg-color);
    border-radius: 6px;
    border: 1px dashed var(--border-color);
    color: var(--text-secondary);
  }

  .changelog-content {
    padding: 1.5rem;
    background: var(--bg-color);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    min-height: 0;
    font-family:
      'Segoe UI',
      system-ui,
      -apple-system,
      sans-serif;
    font-size: 0.95em;
    line-height: 1.7;
    color: var(--text-color);
    user-select: text;
    letter-spacing: 0.01em;
  }

  .no-changelog {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-color);
    border-radius: 6px;
    border: 1px dashed var(--border-color);
    color: var(--text-secondary);
  }

  .version-section {
    margin-bottom: 1.5em;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    overflow: hidden;
  }

  .version-header {
    text-align: center;
    color: var(--primary-color);
    padding: 0.8em;
    margin: 0;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 3.5rem;
    gap: 0.5em;
    font-size: large;
    background-color: var(--useful-changelog-header-bg, var(--card-bg));
    transition: background-color 0.2s;
  }

  .version-header:hover {
    background-color: var(--useful-changelog-header-bg-hover, var(--bg-color));
  }

  .toggle-icon {
    font-size: 0.8em;
    margin-left: 0.5em;
  }

  .version-content {
    padding: 0.75rem 1.1rem 1.1rem;
    background: var(--bg-color);
    font-size: 0.95rem;
    line-height: 1.55;
    color: var(--text-color);
  }

  .cl-section + .cl-section {
    margin-top: 1.2rem;
  }

  .cl-section-title {
    margin: 0 0 0.65rem;
    font-size: 1.05rem;
    color: var(--primary-color);
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .cl-group + .cl-group {
    margin-top: 0.8rem;
  }

  .cl-group-title {
    margin: 0 0 0.4rem;
    font-size: 0.96rem;
    color: var(--text-color-primary, var(--text-color));
  }

  .cl-paragraph {
    margin: 0.25rem 0;
    color: var(--text-color);
  }

  .cl-list {
    margin: 0.35rem 0 0.45rem 1.1rem;
    padding: 0;
    list-style: disc;
  }

  .cl-list li {
    margin: 0.2rem 0;
    color: var(--text-color-secondary, var(--text-color));
  }

  .cl-code {
    margin: 0.45rem 0;
    padding: 0.55rem 0.65rem;
    border-radius: 6px;
    border: 1px solid var(--border-color);
    background: var(--bg-color-dark, rgba(0, 0, 0, 0.2));
    overflow-x: auto;
  }

  .cl-code code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
    font-size: 0.88rem;
  }

  .cl-subtitle {
    margin: 0.45rem 0 0.25rem;
    font-size: 0.92rem;
    color: var(--accent-color, #ffd700);
  }

  .cl-spacer {
    height: 1rem;
  }

  .legacy-content :global(p) {
    margin: 0.3rem 0;
  }
</style>
