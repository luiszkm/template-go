import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ApiError } from "@/api/result";
import { LoadError, Loading } from "@/components/States";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { forbidden } from "@/lib/problems";
import { can, useMe } from "@/lib/session";
import { type Role, rolesQuery } from "./api";
import { loadFailed } from "./copy";

function permissionCount(role: Role) {
  const permissions = role.permissions ?? [];
  return permissions.includes("*") ? "Todas" : String(permissions.length);
}

export function RolesList() {
  const { data: me } = useMe();
  const allowed = can(me, "rbac:read");
  const roles = useQuery({ ...rolesQuery, enabled: allowed });

  if (!allowed || (roles.error instanceof ApiError && roles.error.status === 403)) return <p>{forbidden}</p>;
  if (roles.isPending) return <Loading />;
  if (roles.isError) return <LoadError message={loadFailed} retry={() => roles.refetch()} />;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Papéis</h1>
        {can(me, "rbac:create") && (
          <Button asChild>
            <Link to="/roles/new">Novo papel</Link>
          </Button>
        )}
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Permissões</TableHead>
            <TableHead>Usuários</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {roles.data.map((role) => (
            <TableRow key={role.id}>
              <TableCell>
                <Link to="/roles/$id" params={{ id: role.id }} className="underline">
                  {role.name}
                </Link>
              </TableCell>
              <TableCell>{permissionCount(role)}</TableCell>
              <TableCell>{role.user_count}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
