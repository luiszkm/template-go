import { useQuery } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { type FormEvent, useEffect, useState } from "react";
import { ApiError } from "@/api/result";
import { Field } from "@/components/Field";
import { LoadError, Loading } from "@/components/States";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { type FieldErrors, forbidden } from "@/lib/problems";
import { can, useMe } from "@/lib/session";
import {
  adminRole,
  permissionsQuery,
  type Role,
  type RoleChanges,
  roleQuery,
  useDeleteRole,
  useUpdateRole,
} from "./api";
import { adminLocked, loadFailed, notFound } from "./copy";
import { PermissionPicker } from "./PermissionPicker";
import { roleErrors } from "./roleErrors";

function sameSet(a: string[], b: string[]) {
  return [...a].sort().join("\n") === [...b].sort().join("\n");
}

function changedFields(role: Role, name: string, permissions: string[]): RoleChanges {
  const changes: RoleChanges = {};
  if (name !== role.name) changes.name = name;
  if (!sameSet(role.permissions ?? [], permissions)) changes.permissions = permissions;
  return changes;
}

export function RoleDetail({ id }: { id: string }) {
  const { data: me } = useMe();
  const allowed = can(me, "rbac:read");
  const role = useQuery({ ...roleQuery(id), enabled: allowed });
  const catalogue = useQuery({ ...permissionsQuery, enabled: allowed });
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [permissions, setPermissions] = useState<string[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (role.data) {
      setName(role.data.name);
      setPermissions(role.data.permissions ?? []);
    }
  }, [role.data]);

  const save = useUpdateRole(id);
  const remove = useDeleteRole(id);

  const failure = role.error ?? catalogue.error;
  if (!allowed || (failure instanceof ApiError && failure.status === 403)) return <p>{forbidden}</p>;
  if (role.error instanceof ApiError && role.error.status === 404) return <p>{notFound}</p>;
  if (role.isPending || catalogue.isPending) return <Loading />;
  if (role.isError || catalogue.isError) {
    return (
      <LoadError
        message={loadFailed}
        retry={() => {
          role.refetch();
          catalogue.refetch();
        }}
      />
    );
  }

  const current = role.data;
  const locked = current.name === adminRole;
  const editable = !locked && can(me, "rbac:update");

  function submit(event: FormEvent) {
    event.preventDefault();
    setErrors({});
    setSaved(false);
    const changes = changedFields(current, name, permissions);
    if (Object.keys(changes).length === 0) return;
    save.mutate(changes, {
      onSuccess: () => setSaved(true),
      onError: (error) => setErrors(roleErrors(error)),
    });
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">{current.name}</h1>
      {can(me, "audit:read") && (
        <Link
          to="/audit"
          search={{ resource_type: "role", resource_id: current.id }}
          className="block text-sm underline"
        >
          Ver auditoria
        </Link>
      )}
      {locked && <p>{adminLocked}</p>}
      <form onSubmit={submit} className="max-w-md space-y-4">
        <Field
          id="name"
          label="Nome"
          value={name}
          disabled={!editable}
          onChange={(e) => setName(e.target.value)}
          error={errors.name}
        />
        <PermissionPicker
          catalogue={catalogue.data}
          checked={permissions}
          disabled={!editable}
          onChange={setPermissions}
          error={errors.permissions}
        />
        {editable && (
          <Button type="submit" disabled={save.isPending}>
            Salvar
          </Button>
        )}
        {saved && <p role="status">Alterações salvas.</p>}
      </form>
      {!locked && can(me, "rbac:delete") && (
        <DeleteRole
          role={current}
          onConfirm={() => remove.mutate(undefined, { onSuccess: () => navigate({ to: "/roles" }) })}
        />
      )}
    </div>
  );
}

function DeleteRole({ role, onConfirm }: { role: Role; onConfirm: () => void }) {
  if (role.user_count > 0) {
    return (
      <div className="space-y-2">
        <p>Remova este papel dos {role.user_count} usuários antes de excluí-lo.</p>
        <Button variant="outline" disabled>
          Excluir
        </Button>
      </div>
    );
  }
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">Excluir</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>Excluir papel</DialogTitle>
        <DialogDescription>Excluir o papel {role.name}?</DialogDescription>
        <div className="flex justify-end gap-2">
          <DialogClose asChild>
            <Button variant="outline">Cancelar</Button>
          </DialogClose>
          <DialogClose asChild>
            <Button onClick={onConfirm}>Excluir</Button>
          </DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  );
}
