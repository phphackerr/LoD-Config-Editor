<script module>
  export const tabMetadata = {
    order: 6
  };
</script>

<script>
  //@ts-nocheck
  import {
    GetMapInfo,
    GetChangelog,
    DownloadMap
  } from '/bindings/lce/backend/map_downloader/mapdownloader';
  import {
    SetTaskbarProgress,
    SetTaskbarError,
    SetTaskbarCompleteAndFlash
  } from '/bindings/lce/backend/taskbar/taskbarutils';
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { t } from 'svelte-i18n';
  import { tt } from '../lib/tooltip';
  import { appSettings } from '../lib/store/appSettings';

  let mapInfo = null;
  let loading = false;
  let error = null;
  let downloadProgress = 0;
  let downloadSpeed = 0;
  let downloadedBytes = 0;
  let totalBytes = 0;
  let changelog = '';
  let changelogData = [];
  let expandedVersions = new Set();

  // Реактивно загружаем информацию о карте при изменении пути к игре
  $: if ($appSettings.game_path) {
    loadMapInfo();
  } else {
    mapInfo = null;
  }

  async function loadMapInfo() {
    try {
      loading = true;
      error = null;
      // Используем локальную переменную для проверки актуальности
      const currentPath = $appSettings.game_path;

      const result = await GetMapInfo();

      // Если путь изменился или стал пустым пока мы грузили - игнорируем результат
      if ($appSettings.game_path !== currentPath || !$appSettings.game_path) {
        return;
      }

      mapInfo = result;
      if (mapInfo) {
        const rawChangelogHtml = await GetChangelog(mapInfo.version);
        changelogData = parseHTMLToData(rawChangelogHtml);
      }
    } catch (e) {
      // Тоже проверяем актуальность
      if (!$appSettings.game_path) return;
      error = e;
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

  function parseHTMLToData(html) {
    if (!html) return [];

    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const sections = Array.from(doc.body.children);

    const versions = [];
    let currentVersion = null;
    let currentContent = [];
    let isFirstVersion = true; // Флаг для первой версии

    sections.forEach((section) => {
      if (
        section.tagName === 'P' &&
        section.getAttribute('style')?.includes('text-align: center')
      ) {
        const versionText = section.textContent.trim();
        if (versionText) {
          if (currentVersion) {
            versions.push({
              ...currentVersion,
              content: currentContent.join('').trim()
            });
          }
          currentVersion = {
            version: versionText,
            isExpanded: isFirstVersion // Первая версия будет открыта
          };
          if (isFirstVersion) {
            expandedVersions.add(versionText); // Добавляем в Set для управления состоянием
            isFirstVersion = false;
          }
          currentContent = [];
        }
      } else if (currentVersion) {
        currentContent.push(section.outerHTML);
      }
    });

    if (currentVersion) {
      versions.push({
        ...currentVersion,
        content: currentContent.join('').trim()
      });
    }

    return versions;
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
    if (!mapInfo) return;
    try {
      loading = true;
      error = null;
      downloadProgress = 0;
      downloadSpeed = 0;
      downloadedBytes = 0;
      totalBytes = mapInfo.size || 0;

      const unlisten = await Events.On('download-progress', (event) => {
        downloadProgress = event.data.progress;
        downloadedBytes = event.data.downloaded;
        totalBytes = event.data.total;
        downloadSpeed = event.data.speed;
        SetTaskbarProgress(downloadedBytes, totalBytes);
      });

      const result = await DownloadMap(mapInfo);
      mapInfo = { ...mapInfo, ...result };

      SetTaskbarCompleteAndFlash();

      unlisten();
    } catch (e) {
      error = e;
      downloadProgress = 0;
      downloadSpeed = 0;
      SetTaskbarError();
    } finally {
      loading = false;
      SetTaskbarProgress(0, 0);
    }
  }

  function formatSize(bytes) {
    if (bytes === 0 || bytes === undefined) return $t('unknown');
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

  // onMount(loadMapInfo); // Убрали, так как теперь работает реактивность
</script>

<div class="tab-page">
  <div class="content-wrapper">
    <!-- Левая колонка - загрузка карты -->
    <div class="map-downloader">
      <h3>{$t('download_latest')}</h3>

      {#if error}
        <div class="error">{$t('error')}: {error}</div>
      {/if}

      {#if mapInfo}
        <div class="map-info">
          <div class="info-row">
            <span class="label">{$t('save_path')}:</span>
            <span class="value save-path" use:initTooltip>
              {mapInfo.save_path || $t('unknown')}
            </span>
          </div>
          <div class="info-row">
            <span class="label">{$t('map_version')}:</span>
            <span class="value version">{mapInfo.version}</span>
          </div>
          <div class="info-row">
            <span class="label">{$t('map_date')}:</span>
            <span class="value">{mapInfo.date}</span>
          </div>
          <div class="info-row">
            <span class="label">{$t('map_size')}:</span>
            <span class="value">
              {#if mapInfo.size !== undefined && mapInfo.size !== null}
                {formatSize(mapInfo.size)}
              {:else}
                {$t('unknown')}
              {/if}
            </span>
          </div>
          <div class="info-row">
            <span class="label">{$t('map_status')}:</span>
            <span class="value status" class:downloaded={mapInfo.is_downloaded}>
              {mapInfo.is_downloaded ? $t('downloaded') : $t('not_downloaded')}
            </span>
          </div>
        </div>

        {#if loading && downloadProgress > 0 && downloadProgress < 100}
          <div class="progress-bar-container">
            <div class="progress-bar" style="width: {downloadProgress}%;"></div>
          </div>
          <div class="download-info">
            <span class="download-size">
              {formatSize(downloadedBytes)} / {formatSize(totalBytes)}
            </span>
            <span class="download-speed">
              {formatSpeed(downloadSpeed)}
            </span>
          </div>
        {/if}

        <button on:click={downloadMap} disabled={loading} class="download-btn">
          {#if loading}
            {#if downloadProgress > 0 && downloadProgress < 100}
              {$t('downloading')} ({downloadProgress.toFixed(0)}%)
            {:else}
              {$t('loading')}...
            {/if}
          {:else}
            {$t('download')}
          {/if}
        </button>
      {:else if !loading}
        <div class="no-map">
          {#if !$appSettings.game_path}
            <p>{$t('config_not_found')}</p>
          {:else}
            <p>{$t('map_not_found')}</p>
          {/if}
        </div>
      {/if}

      {#if loading && !mapInfo && !error}
        <div class="loading">
          {$t('loading_map_info')}...
        </div>
      {/if}
    </div>

    <!-- Правая колонка - чейнджлог -->
    <div class="changelog-panel">
      <div class="changelog-header">
        <h3>{$t('changelog')}:</h3>
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
                  {@html item.content}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {:else}
        <div class="no-changelog">
          <p>{$t('loading_changelog')}...</p>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .tab-page {
    height: calc(100% - 15px); /* Учитываем margin-top от .tabcontent */
    padding: 2rem;
    box-sizing: border-box;
    display: flex; /* Добавляем flex для правильного расчета высоты */
  }

  .content-wrapper {
    display: flex;
    gap: 2rem;
    flex: 1; /* Занимаем все доступное пространство */
    max-width: 1200px;
    margin: 0 auto;
    min-height: 0; /* Важно для flex-контейнеров */
  }

  .map-downloader {
    background: var(--card-bg);
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    min-height: 0;
    flex: 1; /* Растягиваемся по высоте */
    max-width: 400px; /* Ограничиваем максимальную ширину */
  }

  .changelog-panel {
    flex: 1;
    background: var(--card-bg);
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    min-height: 0; /* Важно для flex-контейнеров */
  }

  .changelog-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
  }

  .version-badge {
    background: var(--primary-color);
    color: white;
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
    width: calc(100% - 2rem);
    margin: auto 0; /* Центрируем по вертикали */
    overflow-y: auto; /* Добавляем прокрутку для содержимого map-info */
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border-color);
    word-break: break-word;
    min-width: 0;
    gap: 0.5rem; /* Добавляем отступ между label и value */
  }

  .info-row:last-child {
    border-bottom: none;
  }

  .label {
    color: var(--text-secondary);
    flex-shrink: 0;
    min-width: 100px; /* Фиксированная минимальная ширина для меток */
  }

  .value {
    font-weight: 500;
    text-align: right; /* Выравниваем значения по правому краю */
    flex: 1; /* Занимаем оставшееся пространство */
    min-width: 0; /* Позволяем сжиматься */
  }

  .save-path {
    font-size: 0.9em;
    width: 100%; /* Занимаем всю доступную ширину */
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
    background-color: #3ba475;
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: 500;
    font-size: 1rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin: auto 0 0 0; /* Прижимаем к низу */
  }

  .download-btn:hover:not(:disabled) {
    background-color: #2d8c5f;
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
    background-color: var(--card-bg);
    transition: background-color 0.2s;
  }

  .version-header:hover {
    background-color: var(--bg-color);
  }

  .toggle-icon {
    font-size: 0.8em;
    margin-left: 0.5em;
  }

  .version-content {
    padding: 0 1.5em 1em;
    background: var(--bg-color);
    font-family: 'Gill Sans', sans-serif; /* Пример шрифта */
    font-size: 1em; /* Размер шрифта */
    line-height: 1.6; /* Межстрочный интервал */
    color: var(--text-color); /* Цвет текста */
    word-spacing: 2px;
    letter-spacing: 1px;
  }
</style>
