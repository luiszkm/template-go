import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ThemeToggle } from "./ThemeToggle";

describe("ThemeToggle", () => {
  afterEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark");
  });

  it("switches to dark and back to light", async () => {
    localStorage.setItem("theme", "light");
    render(<ThemeToggle />);

    await userEvent.click(screen.getByRole("button", { name: "Usar tema escuro" }));
    expect(document.documentElement).toHaveClass("dark");

    await userEvent.click(screen.getByRole("button", { name: "Usar tema claro" }));
    expect(document.documentElement).not.toHaveClass("dark");
  });

  it("starts from the remembered theme", () => {
    localStorage.setItem("theme", "dark");
    render(<ThemeToggle />);
    expect(screen.getByRole("button", { name: "Usar tema claro" })).toBeInTheDocument();
  });
});
