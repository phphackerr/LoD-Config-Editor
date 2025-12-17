// @ts-nocheck
import './App/lib/icons';
import './App/lib/tooltip';
import { mount } from 'svelte';
import App from './App.svelte';
import { initGoI18n } from './App/lib/store/i18n';
import { loadSettings, appSettings } from './App/lib/store/appSettings';
import { applyTheme } from './App/lib/theming';
import { get } from 'svelte/store';
import { loadConfig } from './App/lib/store/config';

async function initialiseApp() {
  await initGoI18n(); // Ждем инициализации i18n

  await loadSettings().then(() => {
    const theme = get(appSettings).theme;
    // const config = get(appSettings).game_path + "\\config.lod.ini"; // Config is now loaded automatically by config.js subscription

    applyTheme(theme);
    // if (get(appSettings).game_path) {
    //   loadConfig(config);
    // }
  });

  const app = mount(App, { target: document.body });

  return app;
}

const app = initialiseApp();
export default app;
