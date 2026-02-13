import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";
import generateHTML, { getPageEntries } from "./plugins/generate-html";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [wails("./bindings"), react(), tailwindcss(), generateHTML()],
  resolve: {
    alias: {
      "@": "/src",
      "@wails": "/bindings/github.com/chenyang-zz/boxify/internal",
    },
  },

  build: {
    rollupOptions: {
      input: getPageEntries(), // 从配置文件读取入口点
      output: {},
    },
  },
});
