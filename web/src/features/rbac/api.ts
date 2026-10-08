import { queryOptions } from "@tanstack/react-query";
import { api } from "@/api/client";
import type { components } from "@/api/schema";
import { ApiError } from "@/lib/session";

export type Role = components["schemas"]["Role"];
export type RoleChanges = components["schemas"]["RoleChanges"];
export type NewRole = components["schemas"]["NewRole"];

export const adminRole = "admin";

export const rolesQuery = queryOptions({
  queryKey: ["roles"],
  queryFn: async () => {
    const { data, error, response } = await api.GET("/api/v1/rbac/roles");
    if (!data) throw new ApiError(response.status, error);
    return data.items ?? [];
  },
});

export const permissionsQuery = queryOptions({
  queryKey: ["permissions"],
  queryFn: async () => {
    const { data, error, response } = await api.GET("/api/v1/rbac/permissions");
    if (!data) throw new ApiError(response.status, error);
    return data.items ?? [];
  },
});

export const roleQuery = (id: string) =>
  queryOptions({
    queryKey: ["role", id],
    queryFn: async () => {
      const { data, error, response } = await api.GET("/api/v1/rbac/roles/{id}", {
        params: { path: { id } },
      });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
  });

export const userRolesQuery = (id: string) =>
  queryOptions({
    queryKey: ["user-roles", id],
    queryFn: async () => {
      const { data, error, response } = await api.GET("/api/v1/rbac/users/{id}/roles", {
        params: { path: { id } },
      });
      if (!data) throw new ApiError(response.status, error);
      return data.items ?? [];
    },
  });
