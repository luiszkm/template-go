import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { api } from "@/api/client";
import { Button } from "@/components/ui/button";
import { can, meQuery, useMe } from "@/lib/session";

export function UserMenu() {
  const { data: me } = useMe();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  async function signOut() {
    await api.DELETE("/api/v1/users/session");
    queryClient.removeQueries({ queryKey: meQuery.queryKey });
    await navigate({ to: "/login" });
  }

  return (
    <nav className="flex items-center gap-4 border-b border-neutral-200 pb-3 text-sm">
      <Link to="/">Início</Link>
      {can(me, "users:read") && <Link to="/users">Usuários</Link>}
      {can(me, "rbac:read") && <Link to="/roles">Papéis</Link>}
      <Link to="/account/password">Minha senha</Link>
      <span className="ml-auto text-neutral-600">{me?.email}</span>
      <Button variant="outline" size="sm" onClick={signOut}>
        Sair
      </Button>
    </nav>
  );
}
