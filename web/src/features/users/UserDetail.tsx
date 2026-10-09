import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { type FormEvent, useEffect, useState } from "react";
import { ApiError } from "@/api/result";
import { Field } from "@/components/Field";
import { LoadError, Loading } from "@/components/States";
import { Badge } from "@/components/ui/badge";
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
import { can, useMe } from "@/lib/session";
import { type User, userQuery } from "@/lib/user";
import { type UserChanges, useSetUserActive, useUpdateUser } from "./api";
import { emailInUse } from "./copy";

function changedFields(user: User, email: string, name: string): UserChanges {
  const changes: UserChanges = {};
  if (email !== user.email) changes.email = email;
  if (name !== user.name) changes.name = name;
  return changes;
}

export function UserDetail({ id }: { id: string }) {
  const { data: me } = useMe();
  const user = useQuery({ ...userQuery(id), enabled: can(me, "users:read") });

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

  const save = useUpdateUser(id);
  const setActive = useSetUserActive(id);

  if (!can(me, "users:read") || (user.error instanceof ApiError && user.error.status === 403)) {
    return <p>{forbidden}</p>;
  }
  if (user.error instanceof ApiError && user.error.status === 404) return <p>Usuário não encontrado.</p>;
  if (user.isPending) {
    return <Loading />;
  }
  if (user.isError) {
    return <LoadError message="Não foi possível carregar os usuários." retry={() => user.refetch()} />;
  }

  const current = user.data;
  const isSelf = current.id === me?.id;

  function submit(event: FormEvent) {
    event.preventDefault();
    setErrors({});
    setSaved(false);
    const changes = changedFields(current, email, name);
    if (Object.keys(changes).length === 0) return;
    save.mutate(changes, {
      onSuccess: () => setSaved(true),
      onError: (error) => {
        if (!(error instanceof ApiError)) return;
        if (error.status === 409) setErrors({ email: emailInUse });
        if (error.status === 422) setErrors(fieldErrors(error.problem));
      },
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-semibold">{current.email}</h1>
        <Badge variant={current.active ? "secondary" : "outline"}>
          {current.active ? "Ativo" : "Desativado"}
        </Badge>
      </div>
      {can(me, "audit:read") && (
        <div className="flex gap-4 text-sm">
          <Link to="/audit" search={{ resource_type: "user", resource_id: current.id }} className="underline">
            Ver auditoria
          </Link>
          <Link to="/audit" search={{ actor_id: current.id }} className="underline">
            Ver ações
          </Link>
        </div>
      )}
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
