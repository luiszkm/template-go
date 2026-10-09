import { useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { ApiError } from "@/api/result";
import { Field } from "@/components/Field";
import { Button } from "@/components/ui/button";
import { type FieldErrors, fieldErrors, forbidden } from "@/lib/problems";
import { can, useMe } from "@/lib/session";
import { useCreateUser } from "./api";
import { emailInUse } from "./copy";

export function UserForm() {
  const { data: me } = useMe();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});

  const create = useCreateUser();

  if (!can(me, "users:create")) return <p>{forbidden}</p>;

  function submit(event: FormEvent) {
    event.preventDefault();
    setErrors({});
    create.mutate(
      { email, name, password },
      {
        onSuccess: (user) => navigate({ to: "/users/$id", params: { id: user.id } }),
        onError: (error) => {
          if (!(error instanceof ApiError)) return;
          if (error.status === 409) setErrors({ email: emailInUse });
          if (error.status === 422) setErrors(fieldErrors(error.problem));
        },
      },
    );
  }

  return (
    <form onSubmit={submit} className="max-w-sm space-y-4">
      <h1 className="text-2xl font-semibold">Novo usuário</h1>
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
      <Field
        id="password"
        label="Senha"
        type="password"
        autoComplete="new-password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        error={errors.password}
      />
      <Button type="submit" disabled={create.isPending}>
        Criar
      </Button>
    </form>
  );
}
