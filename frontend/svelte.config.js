import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    // ВАЖНО: включаем новый API (Svelte 5)
    compatibility: {
      componentApi: 5,
    },
  },
};
