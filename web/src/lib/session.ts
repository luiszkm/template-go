import { queryOptions, useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import { dataOf } from "@/api/result";
import type { components } from "@/api/schema";

export type Me = components["schemas"]["Me"];

export const meQuery = queryOptions({
  queryKey: ["me"],
  queryFn: async (): Promise<Me> => dataOf(await api.GET("/api/v1/users/me")),
  retry: false,
  staleTime: 60_000,
});

export function useMe() {
  return useQuery(meQuery);
}

export function can(me: Me | undefined, permission: string) {
  return me?.permissions?.includes(permission) ?? false;
}
