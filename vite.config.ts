import { defineConfig } from "vite";
import { resolve } from "path";
import { fileURLToPath } from "url";
import tailwindcss from "@tailwindcss/vite";
import inject from "@rollup/plugin-inject";

const __dirname = fileURLToPath(new URL(".", import.meta.url));

export default defineConfig({
  root: resolve(__dirname, "frontend"),
  base: "/",
  build: {
    outDir: resolve(__dirname, "frontend/dist"),
    emptyOutDir: true,
    manifest: true,
    minify: false,
    sourcemap: true,
    rollupOptions: {
      input: resolve(__dirname, "frontend/main.ts"),
    },
  },
  plugins: [
    tailwindcss(),
    inject({ htmx: 'htmx.org' }),
  ],
});

