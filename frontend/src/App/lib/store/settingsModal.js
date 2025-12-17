import { writable } from 'svelte/store';

export const isSettingsOpen = writable(false);

export const openSettings = () => isSettingsOpen.set(true);
export const closeSettings = () => isSettingsOpen.set(false);
export const toggleSettings = () => isSettingsOpen.update((n) => !n);
