import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, signedInAdmin, stubApi } from "@/test/render";

const id = "00000000-0000-4000-8000-000000000042";
const path = `/api/v1/users/${id}`;
const bia = {
  id,
  email: "b@x.com",
  name: "Bia",
  active: true,
  created_at: "2026-10-08T00:00:00Z",
  deactivated_at: null,
};
const all = ["users:read", "users:update", "users:deactivate", "users:activate"];

describe("UserDetail on /users/$id", () => {
  it("shows loading", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all), [`GET ${path}`]: pending });
    renderAt(`/users/${id}`);
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows forbidden on 403", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all), [`GET ${path}`]: json(403, { status: 403 }) });
    renderAt(`/users/${id}`);
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
  });

  it("shows not found", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all), [`GET ${path}`]: json(404, { status: 404 }) });
    renderAt(`/users/${id}`);
    expect(await screen.findByText("Usuário não encontrado.")).toBeInTheDocument();
  });

  it("patches only changed fields", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET ${path}`]: json(200, bia),
      [`PATCH ${path}`]: json(200, { ...bia, name: "Bia Souza" }),
    });
    renderAt(`/users/${id}`);
    const name = await screen.findByLabelText("Nome");
    await waitFor(() => expect(name).toHaveValue("Bia"));
    await userEvent.clear(name);
    await userEvent.type(name, "Bia Souza");
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByText("Alterações salvas.")).toBeInTheDocument();
    expect(requests.find((r) => r.method === "PATCH")?.body).toEqual({ name: "Bia Souza" });
  });

  it("confirms before deactivating", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET ${path}`]: [json(200, bia), json(200, { ...bia, active: false })],
      [`POST ${path}/deactivate`]: json(204, null),
    });
    renderAt(`/users/${id}`);
    await userEvent.click(await screen.findByRole("button", { name: "Desativar" }));
    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("Desativar b@x.com? As sessões deste usuário serão encerradas.");

    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(requests.filter((r) => r.method === "POST")).toHaveLength(0);

    await userEvent.click(screen.getByRole("button", { name: "Desativar" }));
    const confirm = await screen.findByRole("dialog");
    await userEvent.click(
      Array.from(confirm.querySelectorAll("button")).find(
        (b) => b.textContent === "Desativar",
      ) as HTMLElement,
    );
    await waitFor(() => expect(requests.filter((r) => r.path === `${path}/deactivate`)).toHaveLength(1));
  });

  it("hides deactivate for self", async () => {
    const self = { ...bia, id: signedInAdmin.id, email: signedInAdmin.email };
    stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET /api/v1/users/${signedInAdmin.id}`]: json(200, self),
    });
    renderAt(`/users/${signedInAdmin.id}`);
    expect(await screen.findByRole("button", { name: "Salvar" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Desativar" })).not.toBeInTheDocument();
  });

  it("activates", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET ${path}`]: [json(200, { ...bia, active: false }), json(200, bia)],
      [`POST ${path}/activate`]: json(204, null),
    });
    renderAt(`/users/${id}`);
    expect(await screen.findByText("Desativado")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Ativar" }));
    expect(await screen.findByText("Ativo")).toBeInTheDocument();
    expect(requests.filter((r) => r.path === `${path}/activate`)).toHaveLength(1);
  });

  it("shows error and retries", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET ${path}`]: [json(500, { status: 500 }), json(200, bia)],
    });
    renderAt(`/users/${id}`);
    expect(await screen.findByText("Não foi possível carregar os usuários.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByRole("heading", { name: "b@x.com" })).toBeInTheDocument();
    expect(requests.filter((r) => r.method === "GET" && r.path === path)).toHaveLength(2);
  });

  it("maps 422 errors to fields on edit", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith(...all),
      [`GET ${path}`]: json(200, bia),
      [`PATCH ${path}`]: json(422, {
        status: 422,
        errors: [{ location: "body.name", message: "expected length <= 100" }],
      }),
    });
    renderAt(`/users/${id}`);
    const name = await screen.findByLabelText("Nome");
    await waitFor(() => expect(name).toHaveValue("Bia"));
    await userEvent.type(name, "x");
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByText("expected length <= 100")).toHaveAttribute("id", "name-error");
  });

  it("links to the audit trail of the user", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all, "audit:read"), [`GET ${path}`]: json(200, bia) });
    renderAt(`/users/${id}`);
    const link = await screen.findByRole("link", { name: "Ver auditoria" });
    expect(link).toHaveAttribute("href", `/audit?resource_type=user&resource_id=${id}`);
  });

  it("links to the actions of the user", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all, "audit:read"), [`GET ${path}`]: json(200, bia) });
    renderAt(`/users/${id}`);
    const link = await screen.findByRole("link", { name: "Ver ações" });
    expect(link).toHaveAttribute("href", `/audit?actor_id=${id}`);
  });

  it("hides audit links without audit:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(...all), [`GET ${path}`]: json(200, bia) });
    renderAt(`/users/${id}`);
    expect(await screen.findByRole("heading", { name: "b@x.com" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Salvar" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Ver ações" })).not.toBeInTheDocument();
  });

  it.each([
    ["pending", pending, "Carregando…"],
    ["not found", json(404, { status: 404 }), "Usuário não encontrado."],
    ["forbidden", json(403, { status: 403 }), "Você não tem permissão para acessar esta página."],
    ["failed", json(500, { status: 500 }), "Não foi possível carregar os usuários."],
  ])("hides audit links while %s", async (_, response, text) => {
    stubApi({ "GET /api/v1/users/me": meWith(...all, "audit:read"), [`GET ${path}`]: response });
    renderAt(`/users/${id}`);
    expect(await screen.findByText(text)).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Ver ações" })).not.toBeInTheDocument();
  });
});
