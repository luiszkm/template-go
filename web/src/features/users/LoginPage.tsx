import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { api } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { meQuery } from "@/lib/session";

type Credentials = { email: string; password: string };

class SignInRefused extends Error {
  constructor(
    readonly status: number,
    readonly retryAfterSeconds: number,
  ) {
    super(`sign in answered ${status}`);
  }
}

function safeRedirect(redirect: string | undefined) {
  return redirect?.startsWith("/") && !redirect.startsWith("//") ? redirect : "/";
}

function refusalMessage(error: Error) {
  if (!(error instanceof SignInRefused)) return "Não foi possível entrar. Tente novamente.";
  if (error.status === 401) return "E-mail ou senha inválidos.";
  if (error.status === 429) {
    return `Muitas tentativas. Tente novamente em ${Math.ceil(error.retryAfterSeconds / 60)} minutos.`;
  }
  return "Não foi possível entrar. Tente novamente.";
}

export function LoginPage({ redirect }: { redirect?: string }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const router = useRouter();
  const queryClient = useQueryClient();

  const signIn = useMutation({
    mutationFn: async (credentials: Credentials) => {
      const { response } = await api.POST("/api/v1/users/session", { body: credentials });
      if (response.status !== 204) {
        throw new SignInRefused(response.status, Number(response.headers.get("Retry-After") ?? 0));
      }
    },
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: meQuery.queryKey });
      router.history.push(safeRedirect(redirect));
    },
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    signIn.mutate({ email, password });
  }

  return (
    <form onSubmit={submit} className="mx-auto max-w-sm space-y-4">
      <h1 className="text-2xl font-semibold">Entrar</h1>
      <div className="space-y-1">
        <Label htmlFor="email">E-mail</Label>
        <Input
          id="email"
          type="email"
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
      </div>
      <div className="space-y-1">
        <Label htmlFor="password">Senha</Label>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
      </div>
      {signIn.error && (
        <p role="alert" className="text-sm text-red-700">
          {refusalMessage(signIn.error)}
        </p>
      )}
      <Button type="submit" disabled={signIn.isPending} className="w-full">
        {signIn.isPending ? "Entrando…" : "Entrar"}
      </Button>
    </form>
  );
}
