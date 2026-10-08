import { screen } from "@testing-library/react";
import { json, meWith, renderAt, stubApi } from "@/test/render";

const id = "00000000-0000-4000-8000-000000000042";

describe("/users/$id route", () => {
  it("composes user and roles", async () => {
    stubApi({
      "GET /api/v1/users/me": meWith("users:read", "rbac:read"),
      [`GET /api/v1/users/${id}`]: json(200, {
        id,
        email: "b@x.com",
        name: "Bia",
        active: true,
        created_at: "2026-10-08T00:00:00Z",
        deactivated_at: null,
      }),
      "GET /api/v1/rbac/roles": json(200, { items: [] }),
      [`GET /api/v1/rbac/users/${id}/roles`]: json(200, { items: [] }),
    });
    renderAt(`/users/${id}`);
    expect(await screen.findByRole("heading", { name: "b@x.com" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Papéis" })).toBeInTheDocument();
  });
});
