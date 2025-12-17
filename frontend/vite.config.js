import { defineConfig } from "vite";
import wails from "@wailsio/runtime/plugins/vite";
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [wails("./bindings"), svelte()],
});
