import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, renderAt, stubApi } from "@/test/render";

const id = "00000000-0000-4000-8000-000000000042";
const bia = {
  id,
  email: "b@x.com",
  name: "Bia",
  active: true,
  created_at: "2026-10-08T00:00:00Z",
  deactivated_at: null,
};
const admin = {
  id: "00000000-0000-4000-8000-0000000000a1",
  name: "admin",
  permissions: ["*"],
  user_count: 1,
};
const leitor = { id: "00000000-0000-4000-8000-0000000000a2", name: "Leitor", permissions: [], user_count: 0 };
const rolesPath = `/api/v1/rbac/users/${id}/roles`;

function routes(me: ReturnType<typeof json>, extra: Record<string, ReturnType<typeof json>> = {}) {
  return {
    "GET /api/v1/users/me": me,
    [`GET /api/v1/users/${id}`]: json(200, bia),
    "GET /api/v1/rbac/roles": json(200, { items: [admin, leitor] }),
    [`GET ${rolesPath}`]: json(200, { items: [{ id: admin.id, name: "admin" }] }),
    ...extra,
  };
}

async function section() {
  const heading = await screen.findByRole("heading", { name: "Papéis" });
  const element = heading.closest("section") as HTMLElement;
  await within(element).findByRole("checkbox", { name: "Leitor" });
  return within(element);
}

describe("UserRoles on /users/$id", () => {
  it("checks held roles", async () => {
    stubApi(routes(meWith("users:read", "rbac:read")));
    renderAt(`/users/${id}`);
    const roles = await section();
    await waitFor(() => expect(roles.getByRole("checkbox", { name: "admin" })).toBeChecked());
    expect(roles.getByRole("checkbox", { name: "Leitor" })).not.toBeChecked();
  });

  it("read only without assign", async () => {
    stubApi(routes(meWith("users:read", "rbac:read")));
    renderAt(`/users/${id}`);
    const roles = await section();
    for (const box of roles.getAllByRole("checkbox")) expect(box).toBeDisabled();
    expect(roles.queryByRole("button", { name: "Salvar papéis" })).not.toBeInTheDocument();
  });

  it("confirms and saves", async () => {
    const requests = stubApi(
      routes(meWith("users:read", "rbac:read", "rbac:assign"), { [`PUT ${rolesPath}`]: json(204, null) }),
    );
    renderAt(`/users/${id}`);
    const roles = await section();
    await waitFor(() => expect(roles.getByRole("checkbox", { name: "admin" })).toBeChecked());
    await userEvent.click(roles.getByRole("checkbox", { name: "Leitor" }));

    await userEvent.click(roles.getByRole("button", { name: "Salvar papéis" }));
    expect(await screen.findByRole("dialog")).toHaveTextContent(
      "Alterar os papéis de b@x.com? As sessões deste usuário serão encerradas.",
    );
    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(requests.filter((r) => r.method === "PUT")).toHaveLength(0);

    await userEvent.click(roles.getByRole("button", { name: "Salvar papéis" }));
    await userEvent.click(await screen.findByRole("button", { name: "Confirmar" }));
    expect(await screen.findByText("Papéis atualizados.")).toBeInTheDocument();
    const put = requests.find((r) => r.method === "PUT");
    expect(put?.path).toBe(rolesPath);
    expect(put?.body).toEqual({ role_ids: [admin.id, leitor.id] });
  });

  it("shows last admin conflict", async () => {
    stubApi(
      routes(meWith("users:read", "rbac:read", "rbac:assign"), {
        [`PUT ${rolesPath}`]: json(409, { status: 409 }),
      }),
    );
    renderAt(`/users/${id}`);
    const roles = await section();
    await waitFor(() => expect(roles.getByRole("checkbox", { name: "admin" })).toBeChecked());
    await userEvent.click(roles.getByRole("checkbox", { name: "admin" }));
    await userEvent.click(roles.getByRole("button", { name: "Salvar papéis" }));
    await userEvent.click(await screen.findByRole("button", { name: "Confirmar" }));
    expect(await screen.findByText("Não é possível remover o último administrador.")).toBeInTheDocument();
  });

  it("hidden without read", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      [`GET /api/v1/users/${id}`]: json(200, bia),
    });
    renderAt(`/users/${id}`);
    expect(await screen.findByRole("heading", { name: "b@x.com" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Papéis" })).not.toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/rbac"))).toHaveLength(0);
  });

  it("shows load error", async () => {
    stubApi(
      routes(meWith("users:read", "rbac:read"), { "GET /api/v1/rbac/roles": json(500, { status: 500 }) }),
    );
    renderAt(`/users/${id}`);
    const heading = await screen.findByRole("heading", { name: "Papéis" });
    const element = heading.closest("section") as HTMLElement;
    expect(await within(element).findByText("Não foi possível carregar os papéis.")).toBeInTheDocument();
  });

  it("shows load error for held roles", async () => {
    stubApi(routes(meWith("users:read", "rbac:read"), { [`GET ${rolesPath}`]: json(500, { status: 500 }) }));
    renderAt(`/users/${id}`);
    const heading = await screen.findByRole("heading", { name: "Papéis" });
    const element = heading.closest("section") as HTMLElement;
    expect(await within(element).findByText("Não foi possível carregar os papéis.")).toBeInTheDocument();
  });
});
