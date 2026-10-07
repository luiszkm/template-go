import { QueryClient } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, type RouterHistory } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";

export function makeRouter(history?: RouterHistory) {
  return createRouter({ routeTree, history, defaultPreload: "intent" });
}

/** Router on an in-memory history, for tests. */
export function makeTestRouter(path: string) {
  return makeRouter(createMemoryHistory({ initialEntries: [path] }));
}

export function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof makeRouter>;
  }
}
