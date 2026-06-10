// @vyzorix/config/presets/lib.ts - Library Preset (no SSR)
import { defineViteConfig } from "../vite";

export const libPreset = defineViteConfig({
  proxy: {}, // No proxy needed for library mode
  vite: {
    build: {
      target: "node",
      lib: {
        entry: "src/index.ts",
        formats: ["es", "cjs"],
        fileName: (format) => `index.${format}.js`,
      },
      rollupOptions: {
        external: ["react", "react-dom"],
        output: {
          globals: {
            react: "React",
            "react-dom": "ReactDOM",
          },
        },
      },
    },
  },
});

export default libPreset;