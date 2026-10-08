import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { type FormEvent, useEffect, useState } from "react";
import { api } from "@/api/client";
import type { components } from "@/api/schema";
import { Field } from "@/components/Field";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { type FieldErrors, fieldErrors, forbidden } from "@/lib/problems";
import { ApiError, can, useMe } from "@/lib/session";
import { emailInUse } from "./copy";

type User = components["schemas"]["UserDetail"];
type Changes = components["schemas"]["UserChanges"];

function changedFields(user: User, email: string, name: string): Changes {
  const changes: Changes = {};
  if (email !== user.email) changes.email = email;
  if (name !== user.name) changes.name = name;
  return changes;
}

export function UserDetail({ id }: { id: string }) {
  const { data: me } = useMe();
  const queryClient = useQueryClient();
  const queryKey = ["user", id];
  const user = useQuery({
    queryKey,
    enabled: can(me, "users:read"),
    queryFn: async () => {
      const { data, error, response } = await api.GET("/api/v1/users/{id}", { params: { path: { id } } });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
  });

  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (user.data) {
      setEmail(user.data.email);
      setName(user.data.name);
    }
  }, [user.data]);

  const save = useMutation({
    mutationFn: async (changes: Changes) => {
      const { data, error, response } = await api.PATCH("/api/v1/users/{id}", {
        params: { path: { id } },
        body: changes,
      });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKey, updated);
      setSaved(true);
    },
    onError: (error) => {
      if (!(error instanceof ApiError)) return;
      if (error.status === 409) setErrors({ email: emailInUse });
      if (error.status === 422) setErrors(fieldErrors(error.problem));
    },
  });

  const setActive = useMutation({
    mutationFn: async (active: boolean) => {
      const path = active ? "/api/v1/users/{id}/activate" : "/api/v1/users/{id}/deactivate";
      const { error, response } = await api.POST(path, { params: { path: { id } } });
      if (response.status !== 204) throw new ApiError(response.status, error);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  });

  if (!can(me, "users:read") || (user.error instanceof ApiError && user.error.status === 403)) {
    return <p>{forbidden}</p>;
  }
  if (user.error instanceof ApiError && user.error.status === 404) return <p>Usuário não encontrado.</p>;
  if (user.isPending) {
    return (
      <p role="status" className="text-neutral-600">
        Carregando…
      </p>
    );
  }
  if (user.isError) {
    return (
      <div className="space-y-2">
        <p>Não foi possível carregar os usuários.</p>
        <Button variant="outline" onClick={() => user.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  }

  const current = user.data;
  const isSelf = current.id === me?.id;

  function submit(event: FormEvent) {
    event.preventDefault();
    setErrors({});
    setSaved(false);
    const changes = changedFields(current, email, name);
    if (Object.keys(changes).length > 0) save.mutate(changes);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-semibold">{current.email}</h1>
        <span className="text-sm text-neutral-600">{current.active ? "Ativo" : "Desativado"}</span>
      </div>
      <form onSubmit={submit} className="max-w-sm space-y-4">
        <Field
          id="email"
          label="E-mail"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={errors.email}
        />
        <Field
          id="name"
          label="Nome"
          value={name}
          onChange={(e) => setName(e.target.value)}
          error={errors.name}
        />
        <Button type="submit" disabled={save.isPending}>
          Salvar
        </Button>
        {saved && <p role="status">Alterações salvas.</p>}
      </form>
      {current.active && !isSelf && can(me, "users:deactivate") && (
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="outline">Desativar</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogTitle>Desativar usuário</DialogTitle>
            <DialogDescription>
              Desativar {current.email}? As sessões deste usuário serão encerradas.
            </DialogDescription>
            <div className="flex justify-end gap-2">
              <DialogClose asChild>
                <Button variant="outline">Cancelar</Button>
              </DialogClose>
              <DialogClose asChild>
                <Button onClick={() => setActive.mutate(false)}>Desativar</Button>
              </DialogClose>
            </div>
          </DialogContent>
        </Dialog>
      )}
      {!current.active && can(me, "users:activate") && (
        <Button variant="outline" onClick={() => setActive.mutate(true)}>
          Ativar
        </Button>
      )}
    </div>
  );
}
