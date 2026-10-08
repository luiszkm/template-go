import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, renderAt, signedInAdmin, stubApi } from "@/test/render";

const created = {
  id: "00000000-0000-4000-8000-000000000042",
  email: "b@x.com",
  name: "Bia",
  active: true,
  created_at: "2026-10-08T00:00:00Z",
};

async function fillNewUser() {
  await userEvent.type(await screen.findByLabelText("E-mail"), "b@x.com");
  await userEvent.type(screen.getByLabelText("Nome"), "Bia");
  await userEvent.type(screen.getByLabelText("Senha"), "senha-longa-123");
  await userEvent.click(screen.getByRole("button", { name: "Criar" }));
}

describe("UserForm", () => {
  it("creates and navigates", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith("users:create", "users:read"),
      "POST /api/v1/users": json(201, created),
      [`GET /api/v1/users/${created.id}`]: json(200, { ...created, deactivated_at: null }),
    });
    const { router } = renderAt("/users/new");
    await fillNewUser();
    await waitFor(() => expect(router.state.location.pathname).toBe(`/users/${created.id}`));
    expect(requests.find((r) => r.method === "POST")?.body).toEqual({
      email: "b@x.com",
      name: "Bia",
      password: "senha-longa-123",
    });
  });

  it("shows email conflict", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:create", "users:read", "users:update"),
      "POST /api/v1/users": json(409, { status: 409, detail: "email already in use" }),
      [`GET /api/v1/users/${created.id}`]: json(200, { ...created, deactivated_at: null }),
      [`PATCH /api/v1/users/${created.id}`]: json(409, { status: 409, detail: "email already in use" }),
    });
    const createPage = renderAt("/users/new");
    await fillNewUser();
    expect(await screen.findByText("Este e-mail já está em uso.")).toBeInTheDocument();
    expect(screen.getByLabelText("E-mail")).toHaveAttribute("aria-describedby", "email-error");
    createPage.unmount();

    renderAt(`/users/${created.id}`);
    const email = await screen.findByLabelText("E-mail");
    await waitFor(() => expect(email).toHaveValue("b@x.com"));
    await userEvent.clear(email);
    await userEvent.type(email, signedInAdmin.email);
    await userEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByText("Este e-mail já está em uso.")).toBeInTheDocument();
  });

  it("maps 422 errors to fields", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:create"),
      "POST /api/v1/users": json(422, {
        status: 422,
        detail: "validation failed",
        errors: [
          { location: "body.name", message: "expected length >= 1" },
          { location: "body.password", message: "expected length >= 12" },
        ],
      }),
    });
    renderAt("/users/new");
    await fillNewUser();
    expect(await screen.findByText("expected length >= 1")).toHaveAttribute("id", "name-error");
    expect(screen.getByText("expected length >= 12")).toHaveAttribute("id", "password-error");
  });
});
