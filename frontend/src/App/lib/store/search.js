//@ts-nocheck
import { writable } from "svelte/store";

// Хранилище для текущего поискового запроса
export const searchQuery = writable("");

// Хранилище для всех элементов, которые можно найти на странице.
// У него будут методы для регистрации и удаления элементов.
function createSearchableItemsStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,
    register: (item) => {
      update((items) => [...items, item]);
    },
    unregister: (id) => {
      update((items) => items.filter((item) => item.id !== id));
    },
    update: (id, newData) => {
      update((items) =>
        items.map((item) => (item.id === id ? { ...item, ...newData } : item)),
      );
    },
  };
}

export const searchableItems = createSearchableItemsStore();

export const activeTab = writable("");
