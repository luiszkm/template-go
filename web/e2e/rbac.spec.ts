import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { createAdmin, signIn } from "./session";

test("admin grants a role and the user gets exactly its access", async ({ page, browser }) => {
  const admin = createAdmin();
  const suffix = randomUUID().slice(0, 8);
  const reader = { email: `leitor-${suffix}@x.com`, password: "senha-leitor-123" };
  const roleName = `Leitor ${suffix}`;

  await page.goto("/login");
  await signIn(page, admin);
  await expect(page.getByRole("button", { name: "Sair" })).toBeVisible();

  await page.goto("/users/new");
  await page.getByLabel("E-mail").fill(reader.email);
  await page.getByLabel("Nome").fill("Leitor");
  await page.getByLabel("Senha").fill(reader.password);
  await page.getByRole("button", { name: "Criar" }).click();
  await expect(page).toHaveURL(/\/users\/[0-9a-f-]{36}$/);
  const readerPage = page.url();

  await page.goto("/roles/new");
  await page.getByLabel("Nome").fill(roleName);
  await page.getByRole("checkbox", { name: "users:read" }).check();
  await page.getByRole("button", { name: "Criar" }).click();
  await expect(page).toHaveURL(/\/roles\/[0-9a-f-]{36}$/);
  await expect(page.getByRole("heading", { name: roleName })).toBeVisible();

  await page.goto(readerPage);
  await page.getByRole("checkbox", { name: roleName }).check();
  await page.getByRole("button", { name: "Salvar papéis" }).click();
  await page.getByRole("button", { name: "Confirmar" }).click();
  await expect(page.getByText("Papéis atualizados.")).toBeVisible();

  const context = await browser.newContext();
  const readerTab = await context.newPage();
  await readerTab.goto("/login?redirect=/users");
  await signIn(readerTab, reader);
  await expect(readerTab.getByRole("table")).toContainText(reader.email);
  await readerTab.goto("/roles");
  await expect(readerTab.getByText("Você não tem permissão para acessar esta página.")).toBeVisible();
  await context.close();
});
