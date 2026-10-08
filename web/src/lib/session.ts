import { queryOptions, useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import type { components } from "@/api/schema";

export type Me = components["schemas"]["Me"];

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly problem?: components["schemas"]["Problem"],
  ) {
    super(problem?.detail ?? `API answered ${status}`);
  }
}

export const meQuery = queryOptions({
  queryKey: ["me"],
  queryFn: async (): Promise<Me> => {
    const { data, error, response } = await api.GET("/api/v1/users/me");
    if (!data) throw new ApiError(response.status, error);
    return data;
  },
  retry: false,
  staleTime: 60_000,
});

export function useMe() {
  return useQuery(meQuery);
}

export function can(me: Me | undefined, permission: string) {
  return me?.permissions?.includes(permission) ?? false;
}
