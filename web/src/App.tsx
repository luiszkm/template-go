import { type QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { useEffect } from "react";
import { meQuery } from "@/lib/session";
import { onUnauthorized } from "./api/client";
import type { makeRouter } from "./router";

export function App({
  router,
  queryClient,
}: {
  router: ReturnType<typeof makeRouter>;
  queryClient: QueryClient;
}) {
  useEffect(
    () =>
      onUnauthorized(() => {
        const { pathname, href } = router.state.location;
        queryClient.removeQueries({ queryKey: meQuery.queryKey });
        if (pathname !== "/login") void router.navigate({ to: "/login", search: { redirect: href } });
      }),
    [router, queryClient],
  );
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
