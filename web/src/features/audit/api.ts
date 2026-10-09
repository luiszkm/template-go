import { queryOptions } from "@tanstack/react-query";
import { api } from "@/api/client";
import { dataOf } from "@/api/result";
import type { components } from "@/api/schema";

export type AuditEvent = components["schemas"]["AuditEvent"];
export type AuditEventDetail = components["schemas"]["AuditEventDetail"];

export type AuditFilters = {
  action?: string;
  actor_id?: string;
  resource_type?: string;
  resource_id?: string;
  de?: string;
  ate?: string;
};

function localMidnight(date: string, plusDays = 0) {
  const [year, month, day] = date.split("-").map(Number);
  return new Date(year ?? 0, (month ?? 1) - 1, (day ?? 1) + plusDays).toISOString();
}

export function periodOf(filters: AuditFilters) {
  return {
    from: filters.de ? localMidnight(filters.de) : undefined,
    to: filters.ate ? localMidnight(filters.ate, 1) : undefined,
  };
}

export const actionsQuery = queryOptions({
  queryKey: ["audit-actions"],
  queryFn: async () => dataOf(await api.GET("/api/v1/audit/actions")).items ?? [],
});

export async function fetchEvents(filters: AuditFilters, before?: number) {
  const { de: _de, ate: _ate, ...rest } = filters;
  return dataOf(
    await api.GET("/api/v1/audit/events", { params: { query: { ...rest, ...periodOf(filters), before } } }),
  );
}

export const eventQuery = (id: number) =>
  queryOptions({
    queryKey: ["audit-event", id],
    queryFn: async () => dataOf(await api.GET("/api/v1/audit/events/{id}", { params: { path: { id } } })),
  });
