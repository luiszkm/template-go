import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: () => (
    <main className="mx-auto max-w-3xl p-6">
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
