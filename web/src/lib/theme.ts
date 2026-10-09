export type Theme = "light" | "dark";

const storageKey = "theme";

export function resolveTheme(stored: string | null, prefersDark: boolean): Theme {
  if (stored === "light" || stored === "dark") return stored;
  return prefersDark ? "dark" : "light";
}

export function readTheme(): Theme {
  const prefersDark = globalThis.matchMedia?.("(prefers-color-scheme: dark)").matches ?? false;
  return resolveTheme(localStorage.getItem(storageKey), prefersDark);
}

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", theme === "dark");
  localStorage.setItem(storageKey, theme);
}
