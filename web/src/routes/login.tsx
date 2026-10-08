import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { LoginPage } from "@/features/users/LoginPage";

export const Route = createFileRoute("/login")({
  validateSearch: z.object({ redirect: z.string().optional() }),
  component: LoginRoute,
});

function LoginRoute() {
  const { redirect } = Route.useSearch();
  return <LoginPage redirect={redirect} />;
}
