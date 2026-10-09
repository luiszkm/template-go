import type { QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: () => (
    <main className="mx-auto max-w-5xl p-6">
      <Outlet />
    </main>
  ),
  notFoundComponent: NotFound,
});

function NotFound() {
  return (
    <div className="space-y-2">
      <h1 className="text-xl font-semibold">Página não encontrada</h1>
      <Link to="/" className="underline">
        Voltar para o início
      </Link>
    </div>
  );
}
