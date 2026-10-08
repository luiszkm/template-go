import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export const backendProxy = {
  "/api": "http://localhost:8080",
  "/healthz": "http://localhost:8080",
  "/readyz": "http://localhost:8080",
};

export default defineConfig({
  plugins: [
    tanstackRouter({
      target: "react",
      autoCodeSplitting: true,
      routeFileIgnorePattern: String.raw`\.test\.`,
    }),
    react(),
    tailwindcss(),
  ],
  resolve: { alias: { "@": path.resolve(import.meta.dirname, "src") } },
  server: { port: 5173, proxy: backendProxy },
  test: {
    environment: "jsdom",
    testTimeout: 30_000,
    globals: true,
    setupFiles: [path.resolve(import.meta.dirname, "src/test/setup.ts")],
    include: ["src/**/*.test.{ts,tsx}", "vite.config.test.ts"],
  },
});
