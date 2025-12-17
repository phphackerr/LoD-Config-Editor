//@ts-nocheck
import { Events } from '@wailsio/runtime';
import { StartWatching, StopWatching } from '/bindings/lce/backend/config_watcher/configwatcher';
import { get } from 'svelte/store';
import { isInternalChange } from './store/internalChange';

let currentPath = null;

// --- API для UI --- //
export function onConfigChanged(callback) {
  const listener = (event) => {
    const filePath = Array.isArray(event.data) ? event.data[0] : event.data;

    // Игнорируем изменения, сделанные программой
    if (get(isInternalChange)) {
      console.log('⚡ Internal config changed:', filePath);
      return;
    }

    // Внешние изменения
    console.log('⚡ External config changed:', filePath);
    // Events.Emit("external-config-changed", filePath); // для бэка
    callback(filePath);
  };

  Events.On('config-changed', listener);

  return {
    off: () => Events.Off('config-changed', listener)
  };
}

export async function startWatcher(path) {
  if (!path) return;

  if (currentPath) {
    try {
      await StopWatching();
      console.log('🛑 Previous watcher stopped');
    } catch (err) {
      console.warn('⚠ Ошибка при остановке предыдущего вотчера:', err);
    } finally {
      currentPath = null;
    }
  }

  try {
    await StartWatching(path, 200);
    currentPath = path;
    console.log('👀 Watcher started for:', path);
  } catch (err) {
    console.error('Ошибка запуска вотчера:', err);
    throw err; // Это очень важно! Перебрасываем ошибку, чтобы Svelte мог ее поймать.
  }
}

export async function stopWatcher() {
  try {
    await StopWatching();
    console.log('🛑 Watcher stopped');
  } catch (err) {
    console.error('Ошибка остановки вотчера:', err);
    throw err; // Это очень важно! Перебрасываем ошибку.
  } finally {
    currentPath = null;
  }
}
