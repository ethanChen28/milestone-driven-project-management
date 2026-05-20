import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  test: {
    include: ["src/**/*.test.ts"],
    exclude: ["e2e/**", "node_modules/**"],
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.API_PROXY_TARGET || "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
});
