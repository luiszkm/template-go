import { defineConfig, devices } from "@playwright/test";

// `task e2e` runs this; webServer builds the binary, starts Postgres, migrates and serves on :8080.
export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  use: { baseURL: "http://localhost:8080", trace: "retain-on-failure" },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: {
    command: "task e2e:serve",
    url: "http://localhost:8080/readyz",
    reuseExistingServer: !process.env.CI,
    timeout: 300_000,
    stdout: "pipe",
  },
});
