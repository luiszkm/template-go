import { type FormEvent, useState } from "react";
import { ApiError } from "@/api/result";
import { Field } from "@/components/Field";
import { Button } from "@/components/ui/button";
import { type FieldErrors, fieldErrors } from "@/lib/problems";
import { useChangePassword } from "./api";

export function ChangePassword() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});

  const change = useChangePassword();

  function submit(event: FormEvent) {
    event.preventDefault();
    if (next !== confirmation) {
      setErrors({ confirmation: "As senhas não conferem." });
      return;
    }
    setErrors({});
    change.mutate(
      { current_password: current, new_password: next },
      {
        onError: (error) => {
          if (error instanceof ApiError && error.status === 422) setErrors(fieldErrors(error.problem));
        },
      },
    );
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
