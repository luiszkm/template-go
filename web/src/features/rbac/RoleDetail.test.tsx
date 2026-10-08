import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const id = "00000000-0000-4000-8000-0000000000c1";
const path = `/api/v1/rbac/roles/${id}`;
const leitor = { id, name: "Leitor", permissions: ["users:read"], user_count: 0 };
const catalogue = json(200, { items: ["rbac:read", "users:create", "users:read"] });
const all = meWith("rbac:read", "rbac:update", "rbac:delete");

function routes(
  role: unknown,
  extra: Record<string, ReturnType<typeof json> | ReturnType<typeof json>[]> = {},
) {
  return {
    "GET /api/v1/users/me": all,
    "GET /api/v1/rbac/permissions": catalogue,
    [`GET ${path}`]: json(200, role),
    ...extra,
  };
}

describe("RoleDetail on /roles/$id", () => {
  it("shows loading", async () => {
    stubApi({
      "GET /api/v1/users/me": all,
      "GET /api/v1/rbac/permissions": catalogue,
      [`GET ${path}`]: pending,
    });
    renderAt(`/roles/${id}`);
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows error with retry", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": all,
      "GET /api/v1/rbac/permissions": catalogue,
      [`GET ${path}`]: [json(500, { status: 500 }), json(200, leitor)],
    });
    renderAt(`/roles/${id}`);
    expect(await screen.findByText("Não foi possível carregar os papéis.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByRole("heading", { name: "Leitor" })).toBeInTheDocument();
    expect(requests.filter((r) => r.path === path)).toHaveLength(2);
  });

  it("forbidden without rbac:read", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith("users:read") });
    renderAt(`/roles/${id}`);
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/rbac"))).toHaveLength(0);
  });

  it("shows not found", async () => {
    stubApi({
      "GET /api/v1/users/me": all,
      "GET /api/v1/rbac/permissions": catalogue,
      [`GET ${path}`]: json(404, { status: 404 }),
    });
    renderAt(`/roles/${id}`);
    expect(await screen.findByText("Papel não encontrado.")).toBeInTheDocument();
  });

  it("groups permissions", async () => {
    stubApi(routes(leitor));
    renderAt(`/roles/${id}`);
    await screen.findByRole("heading", { name: "rbac" });
    const sequence = Array.from(document.querySelectorAll("h3, input[type=checkbox]")).map((el) =>
      el.tagName === "H3" ? `# ${el.textContent}` : (el.parentElement?.textContent ?? ""),
    );
    expect(sequence).toEqual(["# rbac", "rbac:read", "# users", "users:create", "users:read"]);
  });

  it("patches only changed fields", async () => {
    const requests = stubApi(
      routes(leitor, {
        [`PATCH ${path}`]: [
          json(200, { ...leitor, name: "Consulta" }),
          json(200, { ...leitor, name: "Consulta", permissions: ["users:create", "users:read"] }),
        ],
      }),
    );
    renderAt(`/roles/${id}`);
    const name = await screen.findByLabelText("Nome");
    await waitFor(() => expect(name).toHaveValue("Leitor"));
    await userEvent.clear(name);
    await userEvent.type(name, "Consulta");
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByText("Alterações salvas.")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("checkbox", { name: "users:create" }));
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    await waitFor(() => expect(requests.filter((r) => r.method === "PATCH")).toHaveLength(2));
    const patches = requests.filter((r) => r.method === "PATCH").map((r) => r.body);
    expect(patches).toEqual([{ name: "Consulta" }, { permissions: ["users:create", "users:read"] }]);
    expect(await screen.findByText("Alterações salvas.")).toBeInTheDocument();
  });

  it("shows name conflict", async () => {
    stubApi(routes(leitor, { [`PATCH ${path}`]: json(409, { status: 409 }) }));
    renderAt(`/roles/${id}`);
    const name = await screen.findByLabelText("Nome");
    await waitFor(() => expect(name).toHaveValue("Leitor"));
    await userEvent.type(name, "2");
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    const conflict = await screen.findByText("Já existe um papel com este nome.");
    expect(conflict.id).toBe("name-error");
    expect(screen.getByLabelText("Nome")).toHaveAttribute("aria-describedby", "name-error");
  });

  it("shows field errors", async () => {
    stubApi(
      routes(leitor, {
        [`PATCH ${path}`]: json(422, {
          status: 422,
          errors: [
            { location: "body.name", message: "nome inválido" },
            { location: "body.permissions", message: "permissão inválida" },
          ],
        }),
      }),
    );
    renderAt(`/roles/${id}`);
    const name = await screen.findByLabelText("Nome");
    await waitFor(() => expect(name).toHaveValue("Leitor"));
    await userEvent.type(name, "2");
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByText("nome inválido")).toBeInTheDocument();
    expect(name).toHaveAttribute("aria-describedby", "name-error");
    expect(screen.getByText("permissão inválida").id).toBe("permissions-error");
  });

  it("locks the admin role", async () => {
    stubApi(routes({ id, name: "admin", permissions: ["*"], user_count: 1 }));
    renderAt(`/roles/${id}`);
    expect(
      await screen.findByText("O papel admin tem todas as permissões e não pode ser alterado."),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Nome")).toBeDisabled();
    const boxes = screen.getAllByRole("checkbox");
    expect(boxes.length).toBe(3);
    for (const box of boxes) expect(box).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Salvar" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Excluir" })).not.toBeInTheDocument();
  });

  it("confirms deletion", async () => {
    const requests = stubApi(
      routes(leitor, {
        [`DELETE ${path}`]: json(204, null),
        "GET /api/v1/rbac/roles": json(200, { items: [] }),
      }),
    );
    const { router } = renderAt(`/roles/${id}`);
    await userEvent.click(await screen.findByRole("button", { name: "Excluir" }));
    expect(await screen.findByRole("dialog")).toHaveTextContent("Excluir o papel Leitor?");
    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(requests.filter((r) => r.method === "DELETE")).toHaveLength(0);

    await userEvent.click(screen.getByRole("button", { name: "Excluir" }));
    const dialog = await screen.findByRole("dialog");
    await userEvent.click(
      Array.from(dialog.querySelectorAll("button")).find((b) => b.textContent === "Excluir") as HTMLElement,
    );
    await waitFor(() => expect(router.state.location.pathname).toBe("/roles"));
    expect(requests.filter((r) => r.method === "DELETE" && r.path === path)).toHaveLength(1);
  });

  it("blocks deletion in use", async () => {
    stubApi(routes({ ...leitor, user_count: 2 }));
    renderAt(`/roles/${id}`);
    expect(
      await screen.findByText("Remova este papel dos 2 usuários antes de excluí-lo."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Excluir" })).toBeDisabled();
  });

  it("forbidden on API 403", async () => {
    stubApi({
      "GET /api/v1/users/me": all,
      "GET /api/v1/rbac/permissions": catalogue,
      [`GET ${path}`]: json(403, { status: 403 }),
    });
    renderAt(`/roles/${id}`);
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
  });
});
