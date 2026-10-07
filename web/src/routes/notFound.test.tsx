import { screen } from "@testing-library/react";
import { renderAt } from "@/test/render";

describe("unknown route", () => {
  // C43
  it("not found page links back to /", async () => {
    renderAt("/nao-existe");
    expect(await screen.findByText("Página não encontrada")).toBeInTheDocument();
    expect(screen.getByRole("link")).toHaveAttribute("href", "/");
  });
});
