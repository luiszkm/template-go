import { execFileSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import path from "node:path";
import type { Page } from "@playwright/test";

const binary = path.resolve(
  import.meta.dirname,
  "../../app/bin",
  process.platform === "win32" ? "api.exe" : "api",
);
const databaseURL = process.env.DATABASE_URL ?? "postgres://app:app@localhost:5432/app?sslmode=disable";

export function createAdmin() {
  const email = `e2e-${randomUUID()}@x.com`;
  const password = "senha-e2e-12345";
  execFileSync(binary, ["users", "create-admin", "--email", email, "--name", "E2E Admin"], {
    input: `${password}\n`,
    env: { ...process.env, DATABASE_URL: databaseURL },
  });
  return { email, password };
}

export async function signIn(page: Page, admin: { email: string; password: string }) {
  await page.getByLabel("E-mail").fill(admin.email);
  await page.getByLabel("Senha").fill(admin.password);
  await page.getByRole("button", { name: "Entrar" }).click();
}
