import { randomUUID } from "node:crypto";
import { expect, type Page, test } from "@playwright/test";
import { createAdmin, signIn } from "./session";

async function signInAsNewAdmin(page: Page) {
  const admin = createAdmin();
  await page.goto("/login");
  await signIn(page, admin);
  await expect(page.getByRole("button", { name: "Sair" })).toBeVisible();
  return admin;
}

const eventRows = (page: Page) => page.locator("tbody tr");

test("admin follows Ver auditoria from a user", async ({ page }) => {
  await signInAsNewAdmin(page);
  await page.goto("/users/new");
  await page.getByLabel("E-mail").fill(`links-${randomUUID().slice(0, 8)}@x.com`);
  await page.getByLabel("Nome").fill("Com Links");
  await page.getByLabel("Senha").fill("senha-links-1");
  await page.getByRole("button", { name: "Criar" }).click();
  await expect(page).toHaveURL(/\/users\/[0-9a-f-]{36}$/);
  const userId = page.url().split("/").at(-1) ?? "";

  await page.getByRole("link", { name: "Ver auditoria" }).click();

  await expect(page).toHaveURL(/\/audit\?/);
  const search = new URL(page.url()).searchParams;
  expect(search.get("resource_type")).toBe("user");
  expect(search.get("resource_id")).toBe(userId);
  await expect(page.getByRole("button", { name: "Limpar filtros" })).toBeVisible();
  await expect(eventRows(page).filter({ hasText: "user.created" })).toHaveCount(1);
  const rows = await eventRows(page).all();
  for (const row of rows) await expect(row.locator("td").nth(3)).toHaveText(`user ${userId}`);
});

test("admin follows Ver ações on self", async ({ page }) => {
  const admin = await signInAsNewAdmin(page);
  const me = await (await page.request.get("/api/v1/users/me")).json();

  await page.goto(`/users/${me.id}`);
  await page.getByRole("link", { name: "Ver ações" }).click();

  await expect(page).toHaveURL(/\/audit\?/);
  expect(new URL(page.url()).searchParams.get("actor_id")).toBe(me.id);
  await expect(eventRows(page).first()).toBeVisible();
  const rows = await eventRows(page).all();
  for (const row of rows) await expect(row.locator("td").nth(2)).toHaveText(admin.email);
});
