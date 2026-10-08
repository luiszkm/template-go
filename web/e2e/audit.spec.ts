import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { createAdmin, signIn } from "./session";

test("admin finds the creation of a user in the audit log", async ({ page }) => {
  const admin = createAdmin();
  const email = `audit-${randomUUID().slice(0, 8)}@x.com`;

  await page.goto("/login");
  await signIn(page, admin);
  await expect(page.getByRole("button", { name: "Sair" })).toBeVisible();

  await page.goto("/users/new");
  await page.getByLabel("E-mail").fill(email);
  await page.getByLabel("Nome").fill("Auditada");
  await page.getByLabel("Senha").fill("senha-auditada-1");
  await page.getByRole("button", { name: "Criar" }).click();
  await expect(page).toHaveURL(/\/users\/[0-9a-f-]{36}$/);
  const userId = page.url().split("/").at(-1) ?? "";

  await page.goto("/audit");
  await page.getByLabel("Ação").selectOption("user.created");
  const row = page.getByRole("row").filter({ hasText: `user ${userId}` });
  await expect(row).toHaveCount(1);
  await row.getByRole("link").click();

  await expect(page).toHaveURL(/\/audit\/\d+$/);
  await expect(
    page.getByRole("heading", { name: "Depois" }).locator("xpath=following-sibling::pre"),
  ).toContainText(email);
});
