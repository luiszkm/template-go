import { queryOptions } from "@tanstack/react-query";
import { api } from "@/api/client";
import type { components } from "@/api/schema";
import { ApiError } from "@/lib/session";

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

export const loadFailed = "Não foi possível carregar a auditoria.";

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

export function formatWhen(iso: string) {
  const d = new Date(iso);
  const two = (n: number) => String(n).padStart(2, "0");
  return `${two(d.getDate())}/${two(d.getMonth() + 1)}/${d.getFullYear()} ${two(d.getHours())}:${two(d.getMinutes())}:${two(d.getSeconds())}`;
}

export function actorLabel(event: AuditEvent) {
  return event.actor?.email ?? "Sistema";
}

export const actionsQuery = queryOptions({
  queryKey: ["audit-actions"],
  queryFn: async () => {
    const { data, error, response } = await api.GET("/api/v1/audit/actions");
    if (!data) throw new ApiError(response.status, error);
    return data.items ?? [];
  },
});

export async function fetchEvents(filters: AuditFilters, before?: number) {
  const { de: _de, ate: _ate, ...rest } = filters;
  const { data, error, response } = await api.GET("/api/v1/audit/events", {
    params: { query: { ...rest, ...periodOf(filters), before } },
  });
  if (!data) throw new ApiError(response.status, error);
  return data;
}

export const eventQuery = (id: number) =>
  queryOptions({
    queryKey: ["audit-event", id],
    queryFn: async () => {
      const { data, error, response } = await api.GET("/api/v1/audit/events/{id}", {
        params: { path: { id } },
      });
      if (!data) throw new ApiError(response.status, error);
      return data;
    },
  });
