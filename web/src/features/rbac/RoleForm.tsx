import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { api } from "@/api/client";
import { Field } from "@/components/Field";
import { Button } from "@/components/ui/button";
import { type FieldErrors, forbidden } from "@/lib/problems";
import { ApiError, can, useMe } from "@/lib/session";
import { type NewRole, permissionsQuery } from "./api";
import { PermissionPicker } from "./PermissionPicker";
import { roleErrors } from "./roleErrors";
import { LoadError, Loading } from "./States";

export function RoleForm() {
  const { data: me } = useMe();
  const allowed = can(me, "rbac:read") && can(me, "rbac:create");
  const catalogue = useQuery({ ...permissionsQuery, enabled: allowed });
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [permissions, setPermissions] = useState<string[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});

  const create = useMutation({
    mutationFn: async (role: NewRole) => {
      const { data, error, response } = await api.POST("/api/v1/rbac/roles", { body: role });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
    onSuccess: async (created) => {
      await queryClient.invalidateQueries({ queryKey: ["roles"] });
      await navigate({ to: "/roles/$id", params: { id: created.id } });
    },
    onError: (error) => setErrors(roleErrors(error)),
  });

  if (!allowed || (catalogue.error instanceof ApiError && catalogue.error.status === 403)) {
    return <p>{forbidden}</p>;
  }
  if (catalogue.isPending) return <Loading />;
  if (catalogue.isError) return <LoadError retry={() => catalogue.refetch()} />;

  function submit(event: FormEvent) {
    event.preventDefault();
    setErrors({});
    create.mutate({ name, permissions });
  }

  return (
    <form onSubmit={submit} className="max-w-md space-y-4">
      <h1 className="text-2xl font-semibold">Novo papel</h1>
      <Field
        id="name"
        label="Nome"
        value={name}
        onChange={(e) => setName(e.target.value)}
        error={errors.name}
      />
      <PermissionPicker
        catalogue={catalogue.data}
        checked={permissions}
        onChange={setPermissions}
        error={errors.permissions}
      />
      <Button type="submit" disabled={create.isPending}>
        Criar
      </Button>
    </form>
  );
}
