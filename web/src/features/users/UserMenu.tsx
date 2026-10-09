import { Link, type LinkProps, useNavigate } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Button } from "@/components/ui/button";
import { can, useMe } from "@/lib/session";
import { useSignOut } from "./api";

function NavLink({ to, children }: { to: LinkProps["to"]; children: ReactNode }) {
  return (
    <Button variant="ghost" size="sm" asChild>
      <Link to={to} activeOptions={{ exact: to === "/" }} activeProps={{ className: "bg-accent" }}>
        {children}
      </Link>
    </Button>
  );
}

export function UserMenu() {
  const { data: me } = useMe();
  const navigate = useNavigate();
  const signOut = useSignOut();

  return (
    <header className="flex flex-wrap items-center gap-1 border-b pb-3 text-sm">
      <nav className="flex flex-wrap items-center gap-1">
        <NavLink to="/">Início</NavLink>
        {can(me, "users:read") && <NavLink to="/users">Usuários</NavLink>}
        {can(me, "rbac:read") && <NavLink to="/roles">Papéis</NavLink>}
        {can(me, "audit:read") && <NavLink to="/audit">Auditoria</NavLink>}
        <NavLink to="/account/password">Minha senha</NavLink>
      </nav>
      <span className="ml-auto px-2 text-muted-foreground">{me?.email}</span>
      <ThemeToggle />
      <Button
        variant="outline"
        size="sm"
        onClick={() => signOut.mutate(undefined, { onSettled: () => navigate({ to: "/login" }) })}
      >
        Sair
      </Button>
    </header>
  );
}
