import { mount } from 'svelte';
import { Events } from '@wailsio/runtime';
import App from './App.svelte';
import { initGoI18n } from './App/lib/store/i18n';
import { loadSettings, appSettings } from './App/lib/store/appSettings';
import { applyTheme } from './App/lib/store/theming';
import { notifyError } from './App/lib/store/notifications';
import { tr } from './App/lib/store/storeUtils';
import { get } from 'svelte/store';

async function initialiseApp() {
  const i18nResult = await initGoI18n(); // Ждем инициализации i18n
  if (!i18nResult.ok) {
    notifyError(i18nResult.error || tr('ERRORS.i18n.init'));
  }

  const settingsResult = await loadSettings();
  if (!settingsResult.ok && settingsResult.error) {
    notifyError(settingsResult.error);
  }

  const theme = get(appSettings).theme;
  const applyResult = await applyTheme(theme);
  if (!applyResult.ok) {
    notifyError(applyResult.error || tr('ERRORS.theme.apply_initial', { theme }));
  }

  const app = mount(App, { target: document.body });

  Events.On('theme:updated', async (event) => {
    const payload = Array.isArray(event?.data) ? event.data[0] : event?.data;
    const nextTheme =
      String(payload?.theme || get(appSettings).theme || 'default').trim() || 'default';
    const applyResult = await applyTheme(nextTheme, { rollbackOnError: false });
    if (!applyResult.ok) {
      notifyError(applyResult.error || tr('ERRORS.theme.refresh', { theme: nextTheme }));
    }
  });

  return app;
}

const app = initialiseApp();
export default app;
