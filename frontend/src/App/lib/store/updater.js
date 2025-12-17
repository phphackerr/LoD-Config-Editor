import { writable } from 'svelte/store';
import {
  CheckForUpdates,
  DoUpdate,
  CheckForComponentUpdates,
  UpdateComponent
} from '/bindings/lce/backend/updater/updater';
import { Events } from '@wailsio/runtime';

export const updaterStore = writable({
  available: false,
  version: '',
  body: '',
  checking: false,
  downloading: false,
  progress: 0,
  error: null,
  componentUpdates: [] // Array of {type, name, version, changelog}
});

export async function checkForUpdates() {
  updaterStore.update((s) => ({ ...s, checking: true, error: null }));

  try {
    // Check app updates
    const result = await CheckForUpdates();

    // Check component updates
    const components = await CheckForComponentUpdates();

    updaterStore.update((s) => ({
      ...s,
      checking: false,
      available: result.available,
      version: result.version,
      body: result.body,
      componentUpdates: components || [],
      error: result.error || null
    }));
  } catch (err) {
    updaterStore.update((s) => ({ ...s, checking: false, error: err.message }));
  }
}

export async function doUpdate(version) {
  updaterStore.update((s) => ({ ...s, downloading: true, error: null, progress: 0 }));

  try {
    await DoUpdate(version);
    // Success - usually the app restarts or closes, but we can show a message
    updaterStore.update((s) => ({ ...s, downloading: false, progress: 100 }));
    alert('Update downloaded! Please restart the application.');
  } catch (err) {
    updaterStore.update((s) => ({ ...s, downloading: false, error: err.message }));
  }
}

export async function updateComponent(component) {
  updaterStore.update((s) => ({ ...s, downloading: true, error: null }));

  try {
    await UpdateComponent(component);

    // Remove from list
    updaterStore.update((s) => ({
      ...s,
      downloading: false,
      componentUpdates: s.componentUpdates.filter(
        (c) => c.name !== component.name || c.type !== component.type
      )
    }));

    // Reload window to apply changes (simplest way for themes/locales)
    window.location.reload();
  } catch (err) {
    updaterStore.update((s) => ({ ...s, downloading: false, error: err.message }));
  }
}

// Listen for progress events
Events.On('update:progress', (data) => {
  if (data.status === 'downloading') {
    updaterStore.update((s) => ({ ...s, progress: data.percent }));
  }
});
