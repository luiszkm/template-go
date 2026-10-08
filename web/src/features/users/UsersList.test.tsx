import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const user = (n: number, active = true) => ({
  id: `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`,
  email: `u${n}@x.com`,
  name: `User ${n}`,
  active,
  created_at: "2026-10-08T00:00:00Z",
  deactivated_at: active ? null : "2026-10-08T01:00:00Z",
});

describe("UsersList on /users", () => {
  it("shows the users table", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": json(200, { items: [user(1), user(2, false)], total: 2 }),
    });
    renderAt("/users");
    const table = await screen.findByRole("table");
    const headers = within(table)
      .getAllByRole("columnheader")
      .map((h) => h.textContent);
    expect(headers).toEqual(["E-mail", "Nome", "Status"]);
    expect(within(table).getByText("Ativo")).toBeInTheDocument();
    expect(within(table).getByText("Desativado")).toBeInTheDocument();
  });

  it("shows empty state", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": json(200, { items: [], total: 0 }),
    });
    renderAt("/users");
    expect(await screen.findByText("Nenhum usuário encontrado.")).toBeInTheDocument();
  });

  it("shows loading", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("users:read"), "GET /api/v1/users": pending });
    renderAt("/users");
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows error and retries", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": [json(500, { status: 500 }), json(200, { items: [user(1)], total: 1 })],
    });
    renderAt("/users");
    expect(await screen.findByText("Não foi possível carregar os usuários.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByText("u1@x.com")).toBeInTheDocument();
    expect(requests.filter((r) => r.path === "/api/v1/users")).toHaveLength(2);
  });

  it("shows forbidden", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith() });
    renderAt("/users");
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
    expect(requests.filter((r) => r.path === "/api/v1/users")).toHaveLength(0);
  });

  it("paginates by 50", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": json(200, { items: [user(1)], total: 120 }),
    });
    renderAt("/users");
    const previous = await screen.findByRole("button", { name: "Anterior" });
    const next = screen.getByRole("button", { name: "Próxima" });
    expect(previous).toBeDisabled();
    expect(next).toBeEnabled();

    await userEvent.click(next);
    await waitFor(() => expect(requests.at(-1)?.search).toBe("?limit=50&offset=50"));
    await userEvent.click(screen.getByRole("button", { name: "Próxima" }));
    await waitFor(() => expect(requests.at(-1)?.search).toBe("?limit=50&offset=100"));
    await waitFor(() => expect(screen.getByRole("button", { name: "Próxima" })).toBeDisabled());
    expect(screen.getByRole("button", { name: "Anterior" })).toBeEnabled();
  });

  it("hides pagination for 50 users", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": json(200, { items: [user(1)], total: 50 }),
    });
    renderAt("/users");
    expect(await screen.findByRole("table")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Anterior" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Próxima" })).not.toBeInTheDocument();
  });
});
