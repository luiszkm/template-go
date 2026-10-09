import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { ApiError } from "@/api/result";
import { Loading } from "@/components/States";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { can, useMe } from "@/lib/session";
import { userQuery } from "@/lib/user";
import { rolesQuery, useAssignRoles, userRolesQuery } from "./api";
import { lastAdmin, loadFailed } from "./copy";

export function UserRoles({ id }: { id: string }) {
  const { data: me } = useMe();
  if (!can(me, "rbac:read")) return null;
  return <UserRolesSection id={id} canAssign={can(me, "rbac:assign")} />;
}

function UserRolesSection({ id, canAssign }: { id: string; canAssign: boolean }) {
  const roles = useQuery(rolesQuery);
  const held = useQuery(userRolesQuery(id));
  const user = useQuery(userQuery(id));
  const email = user.data?.email ?? "";
  const [checked, setChecked] = useState<string[]>([]);
  const [outcome, setOutcome] = useState<string>();

  useEffect(() => {
    if (held.data) setChecked(held.data.map((r) => r.id));
  }, [held.data]);

  const assign = useAssignRoles(id);

  function toggle(roleId: string) {
    setOutcome(undefined);
    setChecked(checked.includes(roleId) ? checked.filter((r) => r !== roleId) : [...checked, roleId]);
  }

  const heldIds = (held.data ?? []).map((r) => r.id);
  const changed = [...heldIds].sort().join() !== [...checked].sort().join();

  return (
    <section className="max-w-md space-y-3">
      <h2 className="text-lg font-semibold">Papéis</h2>
      {roles.isError || held.isError ? (
        <p>{loadFailed}</p>
      ) : roles.isPending || held.isPending ? (
        <Loading />
      ) : (
        <>
          {roles.data.map((role) => (
            <label key={role.id} className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={checked.includes(role.id)}
                disabled={!canAssign}
                onChange={() => toggle(role.id)}
              />
              {role.name}
            </label>
          ))}
          {canAssign && (
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="outline" disabled={!changed || assign.isPending}>
                  Salvar papéis
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogTitle>Alterar papéis</DialogTitle>
                <DialogDescription>
                  Alterar os papéis de {email}? As sessões deste usuário serão encerradas.
                </DialogDescription>
                <div className="flex justify-end gap-2">
                  <DialogClose asChild>
                    <Button variant="outline">Cancelar</Button>
                  </DialogClose>
                  <DialogClose asChild>
                    <Button
                      onClick={() =>
                        assign.mutate(checked, {
                          onSuccess: () => setOutcome("Papéis atualizados."),
                          onError: (error) => {
                            if (error instanceof ApiError && error.status === 409) setOutcome(lastAdmin);
                          },
                        })
                      }
                    >
                      Confirmar
                    </Button>
                  </DialogClose>
                </div>
              </DialogContent>
            </Dialog>
          )}
          {outcome && <p role="status">{outcome}</p>}
        </>
      )}
    </section>
  );
}
