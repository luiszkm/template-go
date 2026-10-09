import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { dataOf, noContent } from "@/api/result";
import type { components } from "@/api/schema";

export type Role = components["schemas"]["Role"];
export type RoleChanges = components["schemas"]["RoleChanges"];
export type NewRole = components["schemas"]["NewRole"];

export const adminRole = "admin";

export const rolesQuery = queryOptions({
  queryKey: ["roles"],
  queryFn: async () => dataOf(await api.GET("/api/v1/rbac/roles")).items ?? [],
});

export const permissionsQuery = queryOptions({
  queryKey: ["permissions"],
  queryFn: async () => dataOf(await api.GET("/api/v1/rbac/permissions")).items ?? [],
});

export const roleQuery = (id: string) =>
  queryOptions({
    queryKey: ["role", id],
    queryFn: async () => dataOf(await api.GET("/api/v1/rbac/roles/{id}", { params: { path: { id } } })),
  });

export const userRolesQuery = (id: string) =>
  queryOptions({
    queryKey: ["user-roles", id],
    queryFn: async () =>
      dataOf(await api.GET("/api/v1/rbac/users/{id}/roles", { params: { path: { id } } })).items ?? [],
  });

export function useCreateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (role: NewRole) => dataOf(await api.POST("/api/v1/rbac/roles", { body: role })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: rolesQuery.queryKey }),
  });
}

export function useUpdateRole(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (changes: RoleChanges) =>
      dataOf(await api.PATCH("/api/v1/rbac/roles/{id}", { params: { path: { id } }, body: changes })),
    onSuccess: (updated) => queryClient.setQueryData(roleQuery(id).queryKey, updated),
  });
}

export function useDeleteRole(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () =>
      noContent(await api.DELETE("/api/v1/rbac/roles/{id}", { params: { path: { id } } })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: rolesQuery.queryKey }),
  });
}

export function useAssignRoles(userId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (roleIds: string[]) =>
      noContent(
        await api.PUT("/api/v1/rbac/users/{id}/roles", {
          params: { path: { id: userId } },
          body: { role_ids: roleIds },
        }),
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: userRolesQuery(userId).queryKey });
      await queryClient.invalidateQueries({ queryKey: rolesQuery.queryKey });
    },
  });
}
