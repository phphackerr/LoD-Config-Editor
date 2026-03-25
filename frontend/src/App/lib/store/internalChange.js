import { writable } from 'svelte/store';

function createInternalChangeStore() {
  const { subscribe, set } = writable(false);
  let timeoutId;

  return {
    subscribe,
    // Метод для отметки внутреннего изменения
    mark() {
      set(true);
      // Если был предыдущий таймаут, сбрасываем его
      if (timeoutId) clearTimeout(timeoutId);

      // Держим флаг дольше дебаунса watcher, чтобы не ловить свои же fs-события.
      timeoutId = setTimeout(() => set(false), 500);
    }
  };
}

export const isInternalChange = createInternalChangeStore();
