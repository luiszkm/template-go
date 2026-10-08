import { useMutation } from "@tanstack/react-query";
import { type FormEvent, useState } from "react";
import { api } from "@/api/client";
import { Field } from "@/components/Field";
import { Button } from "@/components/ui/button";
import { type FieldErrors, fieldErrors } from "@/lib/problems";
import { ApiError } from "@/lib/session";

export function ChangePassword() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});

  const change = useMutation({
    mutationFn: async () => {
      const { error, response } = await api.PUT("/api/v1/users/me/password", {
        body: { current_password: current, new_password: next },
      });
      if (response.status !== 204) throw new ApiError(response.status, error);
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 422) setErrors(fieldErrors(error.problem));
    },
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (next !== confirmation) {
      setErrors({ confirmation: "As senhas não conferem." });
      return;
    }
    setErrors({});
    change.mutate();
  }

  return (
    <form onSubmit={submit} className="max-w-sm space-y-4">
      <h1 className="text-2xl font-semibold">Minha senha</h1>
      <Field
        id="current_password"
        label="Senha atual"
        type="password"
        autoComplete="current-password"
        value={current}
        onChange={(e) => setCurrent(e.target.value)}
        error={errors.current_password}
      />
      <Field
        id="new_password"
        label="Nova senha"
        type="password"
        autoComplete="new-password"
        value={next}
        onChange={(e) => setNext(e.target.value)}
        error={errors.new_password}
      />
      <Field
        id="confirmation"
        label="Confirme a nova senha"
        type="password"
        autoComplete="new-password"
        value={confirmation}
        onChange={(e) => setConfirmation(e.target.value)}
        error={errors.confirmation}
      />
      <Button type="submit" disabled={change.isPending}>
        Alterar senha
      </Button>
      {change.isSuccess && <p role="status">Senha alterada.</p>}
    </form>
  );
}
