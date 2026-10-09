import { Moon, Sun } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { applyTheme, readTheme, type Theme } from "@/lib/theme";

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>(readTheme);
  const next: Theme = theme === "dark" ? "light" : "dark";

  function toggle() {
    applyTheme(next);
    setTheme(next);
  }

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      onClick={toggle}
      aria-label={next === "dark" ? "Usar tema escuro" : "Usar tema claro"}
    >
      {theme === "dark" ? <Sun /> : <Moon />}
    </Button>
  );
}
