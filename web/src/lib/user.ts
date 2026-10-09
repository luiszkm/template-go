import { queryOptions } from "@tanstack/react-query";
import { api } from "@/api/client";
import { dataOf } from "@/api/result";
import type { components } from "@/api/schema";

export type User = components["schemas"]["UserDetail"];

export const userQuery = (id: string) =>
  queryOptions({
    queryKey: ["user", id],
    queryFn: async () => dataOf(await api.GET("/api/v1/users/{id}", { params: { path: { id } } })),
  });
