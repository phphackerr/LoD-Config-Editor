// @ts-nocheck
import { writable, get } from 'svelte/store';
import {
  LoadConfig,
  IsConfigAvailable,
  GetConfigValue,
  SetConfigValue
} from '/bindings/lce/backend/config_editor/configeditor';
import { Events } from '@wailsio/runtime';
import { appSettings } from './appSettings';

export const configStore = writable({
  loading: false,
  error: null,
  data: null,
  path: null
});

// Helper to normalize paths for comparison (handles mixed slashes)
function normalizePath(p) {
  return p ? p.replace(/\\/g, '/').toLowerCase() : null;
}

// === helpers ===
export async function loadConfig(path) {
  if (!path) {
    resetConfig();
    return null;
  }

  // Запоминаем путь, который начали грузить
  const currentPath = path;

  configStore.set({
    loading: true,
    error: null,
    data: null,
    path
  });

  try {
    const config = await LoadConfig(path);

    // Проверка на Race Condition:
    // Если пока мы грузили, путь в сторе изменился (кто-то вызвал loadConfig с другим путем),
    // то игнорируем этот результат.
    const storePath = get(configStore).path;
    if (normalizePath(storePath) !== normalizePath(currentPath)) {
      console.warn(
        `[configStore] Загрузка для ${currentPath} отменена, так как путь изменился на ${storePath}`
      );
      return null;
    }

    configStore.set({
      loading: false,
      error: null,
      data: config,
      path
    });
    console.log('✅ Конфиг загружен:', get(configStore));
    return config;
  } catch (error) {
    // Тоже проверяем актуальность перед записью ошибки
    const storePath = get(configStore).path;
    if (normalizePath(storePath) !== normalizePath(currentPath)) {
      return null;
    }

    configStore.set({
      loading: false,
      error: error?.message ?? String(error),
      data: null,
      path
    });
    console.error('❌ Ошибка загрузки конфига:', error);
    return null;
  }
}

export function resetConfig() {
  configStore.set({
    loading: false,
    error: null,
    data: null,
    path: null
  });
  console.log('🔄 Стор конфига сброшен');
}

// === backend wrappers ===
export async function isConfigAvailable() {
  try {
    return await IsConfigAvailable();
  } catch (err) {
    console.error('Ошибка при проверке наличия конфига:', err);
    return false;
  }
}

export async function getConfigValue(section, option) {
  try {
    return await GetConfigValue(section, option);
  } catch (err) {
    console.error(`Ошибка при получении значения [${section}] ${option}:`, err);
    return null;
  }
}

export async function setConfigValue(section, option, value) {
  try {
    await SetConfigValue(section, option, value);
    console.log(`Значение [${section}] ${option} = ${value} сохранено`);

    configStore.update((s) => {
      // Сбрасываем ошибку при успешном сохранении
      s.error = null;
      if (s.data) {
        if (!s.data[section]) s.data[section] = {};
        s.data[section][option] = value;
      }
      return s;
    });
  } catch (err) {
    console.error(`Ошибка при установке значения [${section}] ${option}:`, err);
    // Обновляем стор, чтобы показать ошибку UI
    configStore.update((s) => {
      s.error = `Не удалось сохранить [${section}] ${option}: ${err}`;
      return s;
    });
    // Можно добавить alert, если нет toast-системы
    // alert(`Ошибка сохранения: ${err}`);
  }
}

// === Автосинхронизация с appSettings ===
let lastLoadedPath = null;

appSettings.subscribe((settings) => {
  const newPath = settings.game_path?.trim() ? `${settings.game_path}/config.lod.ini` : null;

  if (normalizePath(newPath) === normalizePath(lastLoadedPath)) {
    // путь не изменился → не перезагружаем
    return;
  }

  if (newPath) {
    console.log('[configStore] Загружаем конфиг из appSettings:', newPath);
    loadConfig(newPath);
    lastLoadedPath = newPath;
  } else {
    resetConfig();
    lastLoadedPath = null;
  }
});
