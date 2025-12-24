import { writable } from 'svelte/store';

export const isSettingsOpen = writable(false);
export const activeSettingsTab = writable('');

export const openSettings = (tabName = '') => {
  if (tabName) activeSettingsTab.set(tabName);
  isSettingsOpen.set(true);
};
export const closeSettings = () => isSettingsOpen.set(false);
export const toggleSettings = () => isSettingsOpen.update((n) => !n);
