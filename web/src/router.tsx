import { QueryClient } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, type RouterHistory } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";

export function makeRouter(queryClient: QueryClient, history?: RouterHistory) {
  return createRouter({ routeTree, history, context: { queryClient }, defaultPreload: "intent" });
}

export function makeTestRouter(path: string, queryClient: QueryClient) {
  return makeRouter(queryClient, createMemoryHistory({ initialEntries: [path] }));
}

export function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof makeRouter>;
  }
}
