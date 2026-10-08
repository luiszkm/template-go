import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const admin = {
  id: "00000000-0000-4000-8000-0000000000a1",
  name: "admin",
  permissions: ["*"],
  user_count: 1,
};
const leitor = {
  id: "00000000-0000-4000-8000-0000000000a2",
  name: "Leitor",
  permissions: ["users:create", "users:read"],
  user_count: 3,
};
const forbidden = "Você não tem permissão para acessar esta página.";

describe("RolesList on /roles", () => {
  it("shows the roles table", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("rbac:read"),
      "GET /api/v1/rbac/roles": json(200, { items: [admin, leitor] }),
    });
    renderAt("/roles");
    const table = await screen.findByRole("table");
    const headers = within(table)
      .getAllByRole("columnheader")
      .map((h) => h.textContent);
    expect(headers).toEqual(["Nome", "Permissões", "Usuários"]);
    const rows = within(table).getAllByRole("row").slice(1);
    expect(
      rows.map((r) =>
        within(r)
          .getAllByRole("cell")
          .map((c) => c.textContent),
      ),
    ).toEqual([
      ["admin", "Todas", "1"],
      ["Leitor", "2", "3"],
    ]);
  });

  it("shows loading", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("rbac:read"), "GET /api/v1/rbac/roles": pending });
    renderAt("/roles");
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows error with retry", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith("rbac:read"),
      "GET /api/v1/rbac/roles": [json(500, { status: 500 }), json(200, { items: [leitor] })],
    });
    renderAt("/roles");
    expect(await screen.findByText("Não foi possível carregar os papéis.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByText("Leitor")).toBeInTheDocument();
    expect(requests.filter((r) => r.path === "/api/v1/rbac/roles")).toHaveLength(2);
  });

  it("forbidden without rbac:read", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith("users:read") });
    renderAt("/roles");
    expect(await screen.findByText(forbidden)).toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/rbac"))).toHaveLength(0);
  });

  it("forbidden on API 403", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("rbac:read"),
      "GET /api/v1/rbac/roles": json(403, { status: 403 }),
    });
    renderAt("/roles");
    expect(await screen.findByText(forbidden)).toBeInTheDocument();
  });
});
