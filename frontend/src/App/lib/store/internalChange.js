import { writable } from "svelte/store";

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

      // Через 50 мс сбрасываем автоматически
      timeoutId = setTimeout(() => set(false), 50);
    },
  };
}

export const isInternalChange = createInternalChangeStore();
