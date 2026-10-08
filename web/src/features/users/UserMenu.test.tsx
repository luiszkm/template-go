import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, renderAt, stubApi, stubFetch } from "@/test/render";

describe("UserMenu", () => {
  it("signs out", async () => {
    stubFetch(json(200, { status: "ready" }));
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(),
      "GET /readyz": json(200, { status: "ready" }),
      "DELETE /api/v1/users/session": json(204, null),
    });
    const { router } = renderAt("/");
    await userEvent.click(await screen.findByRole("button", { name: "Sair" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/login"));
    expect(requests.some((r) => r.method === "DELETE" && r.path === "/api/v1/users/session")).toBe(true);
  });

  it("hides links without permission", async () => {
    stubApi({ "GET /api/v1/users/me": meWith(), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("button", { name: "Sair" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Usuários" })).not.toBeInTheDocument();
  });

  it("shows the users link with users:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("users:read"), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("link", { name: "Usuários" })).toBeInTheDocument();
  });

  it("roles link only with rbac:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("rbac:read"), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("link", { name: "Papéis" })).toBeInTheDocument();
  });

  it("hides the roles link without rbac:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("users:read"), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("link", { name: "Usuários" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Papéis" })).not.toBeInTheDocument();
  });

  it("audit link with audit:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("audit:read"), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("link", { name: "Auditoria" })).toBeInTheDocument();
  });

  it("hides the audit link without audit:read", async () => {
    stubApi({ "GET /api/v1/users/me": meWith("users:read"), "GET /readyz": json(200, { status: "ready" }) });
    renderAt("/");
    expect(await screen.findByRole("link", { name: "Usuários" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Auditoria" })).not.toBeInTheDocument();
  });
});
