import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const signedIn = json(204, null);
const page = json(200, { items: [], total: 0 });

async function signIn(email = "ana@x.com", password = "senha-longa-123") {
  await userEvent.type(await screen.findByLabelText("E-mail"), email);
  await userEvent.type(screen.getByLabelText("Senha"), password);
  await userEvent.click(screen.getByRole("button", { name: "Entrar" }));
}

describe("LoginPage", () => {
  it("navigates after login", async () => {
    const requests = stubApi({
      "POST /api/v1/users/session": signedIn,
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users": page,
    });
    const withRedirect = renderAt("/login?redirect=%2Fusers");
    await signIn();
    await waitFor(() => expect(withRedirect.router.state.location.pathname).toBe("/users"));
    expect(requests.find((r) => r.method === "POST")?.body).toEqual({
      email: "ana@x.com",
      password: "senha-longa-123",
    });
    withRedirect.unmount();

    const withoutRedirect = renderAt("/login");
    await signIn();
    await waitFor(() => expect(withoutRedirect.router.state.location.pathname).toBe("/"));
  });

  it("shows invalid credentials", async () => {
    stubApi({
      "POST /api/v1/users/session": json(401, { status: 401, detail: "invalid email or password" }),
    });
    renderAt("/login");
    await signIn();
    expect(await screen.findByText("E-mail ou senha inválidos.")).toBeInTheDocument();
    expect(screen.getByLabelText("E-mail")).toHaveValue("ana@x.com");
  });

  it("shows rate limit minutes", async () => {
    stubApi({
      "POST /api/v1/users/session": [
        json(429, { status: 429 }, { "Retry-After": "61" }),
        json(429, { status: 429 }, { "Retry-After": "900" }),
      ],
    });
    renderAt("/login");
    await signIn();
    expect(await screen.findByText("Muitas tentativas. Tente novamente em 2 minutos.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Entrar" }));
    expect(await screen.findByText("Muitas tentativas. Tente novamente em 15 minutos.")).toBeInTheDocument();
  });

  it("disables submit while pending", async () => {
    stubApi({ "POST /api/v1/users/session": pending });
    renderAt("/login");
    await signIn();
    const button = await screen.findByRole("button", { name: "Entrando…" });
    expect(button).toBeDisabled();
  });

  it("shows a generic message for other failures", async () => {
    stubApi({ "POST /api/v1/users/session": json(500, { status: 500 }) });
    renderAt("/login");
    await signIn();
    expect(await screen.findByText("Não foi possível entrar. Tente novamente.")).toBeInTheDocument();
  });

  it("ignores redirects that leave the site", async () => {
    stubApi({
      "POST /api/v1/users/session": signedIn,
      "GET /api/v1/users/me": meWith(),
      "GET /readyz": json(200, { status: "ready" }),
    });
    for (const redirect of ["%2F%2Fevil.example", "https%3A%2F%2Fevil.example", "users"]) {
      const rendered = renderAt(`/login?redirect=${redirect}`);
      await signIn();
      await waitFor(() => expect(rendered.router.state.location.href).toBe("/"));
      rendered.unmount();
    }
  });
});
