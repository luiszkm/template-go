import { expect, test } from "@playwright/test";
import { createAdmin, signIn } from "./session";

test("login and logout", async ({ page }) => {
  const admin = createAdmin();

  await page.goto("/users");
  await expect(page).toHaveURL(/\/login\?redirect=%2Fusers$/);
  await signIn(page, admin);

  await expect(page).toHaveURL(/\/users$/);
  await expect(page.getByRole("table")).toContainText(admin.email);
  expect(await page.evaluate(() => document.cookie)).not.toContain("session=");

  await page.getByRole("button", { name: "Sair" }).click();
  await expect(page).toHaveURL(/\/login$/);
});
