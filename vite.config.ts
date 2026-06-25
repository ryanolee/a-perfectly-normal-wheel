import { defineConfig } from "vite";
import { resolve } from "path";
import { fileURLToPath } from "url";
import tailwindcss from "@tailwindcss/vite";

const __dirname = fileURLToPath(new URL(".", import.meta.url));

export default defineConfig({
  root: resolve(__dirname, "frontend"),
  base: "/",
  build: {
    outDir: resolve(__dirname, "frontend/dist"),
    emptyOutDir: true,
    manifest: true,
    minify: 'oxc',
    sourcemap: true,
    rollupOptions: {
      input: resolve(__dirname, "frontend/main.ts"),
      // htmx ext files (sse.js, remove-me.js) reference a bare global `htmx`.
      // Rolldown's native inject replaces it with the default import.
      transform: { inject: { htmx: 'htmx.org' } },
    },
  },
  plugins: [
    tailwindcss(),
  ],
});

