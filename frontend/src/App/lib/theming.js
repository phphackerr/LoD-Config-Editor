import { LoadTheme } from "/bindings/lce/backend/theming/themeservice";

export async function applyTheme(name) {
  try {
    const theme = await LoadTheme(name);
    console.log(theme);

    // применяем CSS-переменные
    for (const key in theme) {
      document.documentElement.style.setProperty(`--${key}`, theme[key]);
    }
  } catch (e) {
    console.error("Ошибка загрузки темы:", e);
  }
}
