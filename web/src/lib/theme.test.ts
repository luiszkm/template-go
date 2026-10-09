import { applyTheme, readTheme, resolveTheme } from "./theme";

describe("resolveTheme", () => {
  it.each([
    ["light", true, "light"],
    ["dark", false, "dark"],
    [null, true, "dark"],
    [null, false, "light"],
    ["sepia", true, "dark"],
  ] as const)("stored %s with prefersDark %s resolves to %s", (stored, prefersDark, expected) => {
    expect(resolveTheme(stored, prefersDark)).toBe(expected);
  });
});

describe("applyTheme", () => {
  afterEach(() => localStorage.clear());

  it("marks the document dark and remembers the choice", () => {
    applyTheme("dark");
    expect(document.documentElement).toHaveClass("dark");
    expect(readTheme()).toBe("dark");
  });

  it("clears the dark mark for the light theme", () => {
    applyTheme("dark");
    applyTheme("light");
    expect(document.documentElement).not.toHaveClass("dark");
    expect(readTheme()).toBe("light");
  });
});
