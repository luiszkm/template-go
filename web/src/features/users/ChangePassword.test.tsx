import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, renderAt, stubApi } from "@/test/render";

async function fill(current: string, next: string, confirmation: string) {
  await userEvent.type(await screen.findByLabelText("Senha atual"), current);
  await userEvent.type(screen.getByLabelText("Nova senha"), next);
  await userEvent.type(screen.getByLabelText("Confirme a nova senha"), confirmation);
  await userEvent.click(screen.getByRole("button", { name: "Alterar senha" }));
}

describe("ChangePassword on /account/password", () => {
  it("changes password", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": meWith(),
      "PUT /api/v1/users/me/password": [
        json(204, null),
        json(422, {
          status: 422,
          errors: [{ location: "body.current_password", message: "incorrect password" }],
        }),
      ],
    });
    renderAt("/account/password");
    await fill("senha-antiga-123", "senha-nova-1234", "senha-nova-1234");
    expect(await screen.findByText("Senha alterada.")).toBeInTheDocument();
    expect(requests.find((r) => r.method === "PUT")?.body).toEqual({
      current_password: "senha-antiga-123",
      new_password: "senha-nova-1234",
    });

    await userEvent.click(screen.getByRole("button", { name: "Alterar senha" }));
    expect(await screen.findByText("incorrect password")).toHaveAttribute("id", "current_password-error");
  });

  it("rejects mismatch", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith() });
    renderAt("/account/password");
    await fill("senha-antiga-123", "senha-nova-1234", "senha-outra-1234");
    expect(await screen.findByText("As senhas não conferem.")).toBeInTheDocument();
    expect(requests.filter((r) => r.method === "PUT")).toHaveLength(0);
  });
});
