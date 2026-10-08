import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useState } from "react";
import { api } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { forbidden } from "./problems";
import { ApiError, can, useMe } from "./session";

const pageSize = 50;

export function UsersList() {
  const { data: me } = useMe();
  const [offset, setOffset] = useState(0);
  const allowed = can(me, "users:read");
  const users = useQuery({
    queryKey: ["users", offset],
    enabled: allowed,
    queryFn: async () => {
      const { data, error, response } = await api.GET("/api/v1/users", {
        params: { query: { limit: pageSize, offset } },
      });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
  });

  if (!allowed || (users.error instanceof ApiError && users.error.status === 403)) {
    return <p>{forbidden}</p>;
  }
  if (users.isPending) {
    return (
      <p role="status" className="text-neutral-600">
        Carregando…
      </p>
    );
  }
  if (users.isError) {
    return (
      <div className="space-y-2">
        <p>Não foi possível carregar os usuários.</p>
        <Button variant="outline" onClick={() => users.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  }

  const items = users.data.items ?? [];
  const { total } = users.data;
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Usuários</h1>
        {can(me, "users:create") && (
          <Link to="/users/new" className="underline">
            Novo usuário
          </Link>
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
                <TableCell>{user.active ? "Ativo" : "Desativado"}</TableCell>
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
