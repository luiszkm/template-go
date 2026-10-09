import { useRouter } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SignInRefused, useSignIn } from "./api";

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
  const signIn = useSignIn();

  function submit(event: FormEvent) {
    event.preventDefault();
    signIn.mutate({ email, password }, { onSuccess: () => router.history.push(safeRedirect(redirect)) });
  }

  return (
    <Card className="mx-auto mt-16 max-w-sm">
      <CardHeader>
        <CardTitle>
          <h1 className="text-2xl font-semibold">Entrar</h1>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={submit} className="space-y-4">
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
            <Alert variant="destructive">
              <AlertDescription>{refusalMessage(signIn.error)}</AlertDescription>
            </Alert>
          )}
          <Button type="submit" disabled={signIn.isPending} className="w-full">
            {signIn.isPending ? "Entrando…" : "Entrar"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
