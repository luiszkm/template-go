import { screen, waitFor } from "@testing-library/react";
import { json, meWith, renderAt, stubApi } from "@/test/render";

const unauthorized = json(401, { status: 401, title: "Unauthorized", detail: "authentication required" });

describe("protected routes", () => {
  it("redirects to login without session", async () => {
    stubApi({ "GET /api/v1/users/me": unauthorized });
    const { router } = renderAt("/users");
    await waitFor(() => expect(router.state.location.pathname).toBe("/login"));
    expect(router.state.location.href).toBe("/login?redirect=%2Fusers");
    expect(await screen.findByRole("button", { name: "Entrar" })).toBeInTheDocument();
  });

  it("redirects on 401 while signed in", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:read"),
      "GET /api/v1/users/abc": unauthorized,
    });
    const { router } = renderAt("/users/abc");
    await waitFor(() => expect(router.state.location.pathname).toBe("/login"));
    expect(router.state.location.href).toBe("/login?redirect=%2Fusers%2Fabc");
  });
});
