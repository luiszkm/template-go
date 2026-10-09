import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { dataOf, noContent } from "@/api/result";
import type { components } from "@/api/schema";
import { meQuery } from "@/lib/session";
import { userQuery } from "@/lib/user";

export type UserChanges = components["schemas"]["UserChanges"];
export type NewUser = components["schemas"]["NewUser"];
type Credentials = { email: string; password: string };
type PasswordChange = { current_password: string; new_password: string };

export const pageSize = 50;

export const usersQuery = (offset: number) =>
  queryOptions({
    queryKey: ["users", offset],
    queryFn: async () =>
      dataOf(await api.GET("/api/v1/users", { params: { query: { limit: pageSize, offset } } })),
  });

export class SignInRefused extends Error {
  constructor(
    readonly status: number,
    readonly retryAfterSeconds: number,
  ) {
    super(`sign in answered ${status}`);
  }
}

export function useSignIn() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (credentials: Credentials) => {
      const { response } = await api.POST("/api/v1/users/session", { body: credentials });
      if (response.status !== 204) {
        throw new SignInRefused(response.status, Number(response.headers.get("Retry-After") ?? 0));
      }
    },
    onSuccess: () => queryClient.removeQueries({ queryKey: meQuery.queryKey }),
  });
}

export function useSignOut() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await api.DELETE("/api/v1/users/session");
    },
    onSettled: () => queryClient.removeQueries({ queryKey: meQuery.queryKey }),
  });
}

export function useCreateUser() {
  return useMutation({
    mutationFn: async (user: NewUser) => dataOf(await api.POST("/api/v1/users", { body: user })),
  });
}

export function useUpdateUser(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (changes: UserChanges) =>
      dataOf(await api.PATCH("/api/v1/users/{id}", { params: { path: { id } }, body: changes })),
    onSuccess: (updated) => queryClient.setQueryData(userQuery(id).queryKey, updated),
  });
}

export function useSetUserActive(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (active: boolean) => {
      const path = active ? "/api/v1/users/{id}/activate" : "/api/v1/users/{id}/deactivate";
      noContent(await api.POST(path, { params: { path: { id } } }));
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: userQuery(id).queryKey }),
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (change: PasswordChange) =>
      noContent(await api.PUT("/api/v1/users/me/password", { body: change })),
  });
}
