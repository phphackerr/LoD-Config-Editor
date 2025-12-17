//@ts-nocheck

import { writable, get } from 'svelte/store';
import {
  GetSettings,
  UpdateSettings,
  GetOption,
  GetDefaultSettings
} from '/bindings/lce/backend/app_settings/appsettings';

import { CheckAndFindPaths } from '/bindings/lce/backend/paths_scanner/scanner';

// --- Store и базовые методы --- //

export const appSettings = writable({
  width: 1600,
  height: 900,
  language: 'en',
  game_path: '',
  first_run: true,
  all_paths: [],
  theme: 'default',
  show_extra_hotkeys: false
});

export async function loadSettings() {
  try {
    const settings = await GetSettings();
    appSettings.set(settings);
    console.log('Настройки загружены:', settings);
    return settings;
  } catch (error) {
    console.error('Ошибка загрузки настроек:', error);
    // Пытаемся получить дефолтные настройки с бэкенда
    try {
      const defaultSettings = await GetDefaultSettings();
      appSettings.set(defaultSettings);
      return defaultSettings;
    } catch (e) {
      console.error('Critical: Failed to load even default settings', e);
      return {}; // Fallback to empty object if everything fails
    }
  }
}

export async function updateSettings(newSettings) {
  try {
    // Сначала обновляем стор оптимистично (или берем текущее состояние)
    const current = get(appSettings);
    const fullSettings = { ...current, ...newSettings };

    appSettings.set(fullSettings);

    // Отправляем ПОЛНЫЙ объект настроек на бэкенд
    const updatedSettings = await UpdateSettings(fullSettings);

    // Обновляем стор ответом от бэкенда (на случай если там что-то изменилось/валидировалось)
    appSettings.set(updatedSettings);
    console.log('Настройки обновлены:', updatedSettings);
    return updatedSettings;
  } catch (error) {
    console.error('Ошибка обновления настроек:', error);
    await loadSettings(); // Откат к сохраненным настройкам
    throw error;
  }
}

export async function getSetting(key) {
  try {
    return await GetOption(key);
  } catch (error) {
    console.error(`Ошибка получения настройки ${key}:`, error);
    return null;
  }
}

export async function resetSettings() {
  try {
    const defaultSettings = await GetDefaultSettings();
    const updatedSettings = await UpdateSettings(defaultSettings);
    appSettings.set(updatedSettings);
    console.log('Настройки сброшены к значениям по умолчанию');
    return updatedSettings;
  } catch (error) {
    console.error('Ошибка сброса настроек:', error);
    throw error;
  }
}

// --- Хелперы для настроек --- //

export async function updateWindowSize(width, height) {
  return updateSettings({ width, height });
}

export async function updateLanguage(language) {
  console.log(`Язык изменен на: ${language}`);
  return updateSettings({ language });
}

export async function updateGamePath(game_path) {
  console.log(`Путь к игре изменен на: ${game_path}`);
  return updateSettings({ game_path });
}

export async function updateFirstRun(first_run) {
  console.log(`First run установлен в: ${first_run}`);
  return updateSettings({ first_run });
}

export async function updateAllPaths(all_paths) {
  console.log(`Все пути обновлены на: ${all_paths}`);
  return updateSettings({ all_paths });
}

export async function updateTheme(theme) {
  console.log(`Тема изменена на: ${theme}`);
  return updateSettings({ theme });
}

// --- Управление путями (NEW) --- //

// Сканирование системы на предмет путей
export async function runScanner() {
  try {
    const pathsFound = await CheckAndFindPaths();
    console.log('Найденные пути:', pathsFound);

    await updateAllPaths(pathsFound);

    if (pathsFound.length === 1) {
      await updateGamePath(pathsFound[0]);
    }

    const isFirstRun = get(appSettings).first_run;
    if (isFirstRun) {
      await updateFirstRun(false);
    }

    return pathsFound;
  } catch (error) {
    console.error('Ошибка при сканировании путей:', error);
    return [];
  }
}

// Удаление конкретного пути
export async function deletePath(pathToDelete) {
  const current = get(appSettings);
  const updatedPaths = current.all_paths.filter((p) => p !== pathToDelete);

  await updateAllPaths(updatedPaths);

  if (pathToDelete === current.game_path) {
    await updateGamePath(updatedPaths[0] || '');
  }
}
