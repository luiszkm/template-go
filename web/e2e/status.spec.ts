import { expect, test } from "@playwright/test";
import { createAdmin, signIn } from "./session";

// C49
test("api online", async ({ page }) => {
  await page.goto("/");
  await signIn(page, createAdmin());
  await expect(page.getByText("API: online")).toBeVisible();
});

test("unknown page falls back to the SPA not-found view", async ({ page }) => {
  await page.goto("/nao-existe");
  await expect(page.getByText("Página não encontrada")).toBeVisible();
});
