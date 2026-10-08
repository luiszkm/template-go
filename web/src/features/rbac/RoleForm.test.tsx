import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, renderAt, stubApi } from "@/test/render";

const created = "00000000-0000-4000-8000-0000000000b1";
const catalogue = json(200, { items: ["rbac:read", "users:create", "users:read"] });
const me = meWith("rbac:read", "rbac:create");

async function fill(name: string, ...permissions: string[]) {
  await userEvent.type(await screen.findByLabelText("Nome"), name);
  for (const p of permissions) await userEvent.click(screen.getByRole("checkbox", { name: p }));
  await userEvent.click(screen.getByRole("button", { name: "Criar" }));
}

describe("RoleForm on /roles/new", () => {
  it("creates and navigates", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": me,
      "GET /api/v1/rbac/permissions": catalogue,
      "POST /api/v1/rbac/roles": json(201, {
        id: created,
        name: "Leitor",
        permissions: ["users:read"],
        user_count: 0,
      }),
      [`GET /api/v1/rbac/roles/${created}`]: json(200, {
        id: created,
        name: "Leitor",
        permissions: ["users:read"],
        user_count: 0,
      }),
    });
    const { router } = renderAt("/roles/new");
    await fill("Leitor", "users:read");
    await waitFor(() => expect(router.state.location.pathname).toBe(`/roles/${created}`));
    expect(requests.find((r) => r.method === "POST")?.body).toEqual({
      name: "Leitor",
      permissions: ["users:read"],
    });
  });

  it("groups permissions", async () => {
    stubApi({ "GET /api/v1/users/me": me, "GET /api/v1/rbac/permissions": catalogue });
    renderAt("/roles/new");
    await screen.findByRole("heading", { name: "rbac" });
    const sequence = Array.from(document.querySelectorAll("h3, input[type=checkbox]")).map((el) =>
      el.tagName === "H3" ? `# ${el.textContent}` : (el.parentElement?.textContent ?? ""),
    );
    expect(sequence).toEqual(["# rbac", "rbac:read", "# users", "users:create", "users:read"]);
  });

  it("shows name conflict", async () => {
    stubApi({
      "GET /api/v1/users/me": me,
      "GET /api/v1/rbac/permissions": catalogue,
      "POST /api/v1/rbac/roles": json(409, { status: 409 }),
    });
    renderAt("/roles/new");
    await fill("Leitor");
    expect(await screen.findByText("Já existe um papel com este nome.")).toBeInTheDocument();
  });

  it("shows field errors", async () => {
    stubApi({
      "GET /api/v1/users/me": me,
      "GET /api/v1/rbac/permissions": catalogue,
      "POST /api/v1/rbac/roles": json(422, {
        status: 422,
        errors: [
          { location: "body.name", message: "nome inválido" },
          { location: "body.permissions", message: "permissão inválida" },
        ],
      }),
    });
    renderAt("/roles/new");
    await fill("x");
    const name = await screen.findByLabelText("Nome");
    expect(await screen.findByText("nome inválido")).toBeInTheDocument();
    expect(name).toHaveAttribute("aria-describedby", "name-error");
    expect(within(document.body).getByText("permissão inválida").id).toBe("permissions-error");
  });

  it("forbidden without rbac:read", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith("rbac:create") });
    renderAt("/roles/new");
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/rbac"))).toHaveLength(0);
  });
});
