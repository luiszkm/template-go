import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { AuditList } from "@/features/audit/AuditList";

export const Route = createFileRoute("/_authed/audit/")({
  validateSearch: z.object({
    action: z.coerce.string().optional(),
    actor_id: z.coerce.string().optional(),
    resource_type: z.coerce.string().optional(),
    resource_id: z.coerce.string().optional(),
    de: z.coerce.string().optional(),
    ate: z.coerce.string().optional(),
  }),
  component: AuditRoute,
});

function AuditRoute() {
  const filters = Route.useSearch();
  const navigate = Route.useNavigate();
  return <AuditList filters={filters} onFiltersChange={(next) => navigate({ search: next })} />;
}
