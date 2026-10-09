import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useState } from "react";
import { ApiError } from "@/api/result";
import { LoadError, Loading } from "@/components/States";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { forbidden } from "@/lib/problems";
import { can, useMe } from "@/lib/session";
import { pageSize, usersQuery } from "./api";

export function UsersList() {
  const { data: me } = useMe();
  const [offset, setOffset] = useState(0);
  const allowed = can(me, "users:read");
  const users = useQuery({ ...usersQuery(offset), enabled: allowed });

  if (!allowed || (users.error instanceof ApiError && users.error.status === 403)) {
    return <p>{forbidden}</p>;
  }
  if (users.isPending) {
    return <Loading />;
  }
  if (users.isError) {
    return <LoadError message="Não foi possível carregar os usuários." retry={() => users.refetch()} />;
  }

  const items = users.data.items ?? [];
  const { total } = users.data;
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Usuários</h1>
        {can(me, "users:create") && (
          <Button asChild>
            <Link to="/users/new">Novo usuário</Link>
          </Button>
        )}
      </div>
      {items.length === 0 ? (
        <p>Nenhum usuário encontrado.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>E-mail</TableHead>
              <TableHead>Nome</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((user) => (
              <TableRow key={user.id}>
                <TableCell>
                  <Link to="/users/$id" params={{ id: user.id }} className="underline">
                    {user.email}
                  </Link>
                </TableCell>
                <TableCell>{user.name}</TableCell>
                <TableCell>
                  <Badge variant={user.active ? "secondary" : "outline"}>
                    {user.active ? "Ativo" : "Desativado"}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {total > pageSize && (
        <div className="flex gap-2">
          <Button variant="outline" disabled={offset === 0} onClick={() => setOffset(offset - pageSize)}>
            Anterior
          </Button>
          <Button
            variant="outline"
            disabled={offset + pageSize >= total}
            onClick={() => setOffset(offset + pageSize)}
          >
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}
