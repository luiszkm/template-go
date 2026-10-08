import { createFileRoute } from "@tanstack/react-router";
import { AuditDetail } from "@/features/audit/AuditDetail";

export const Route = createFileRoute("/_authed/audit/$id")({ component: AuditDetailRoute });

function AuditDetailRoute() {
  const { id } = Route.useParams();
  return <AuditDetail id={Number(id)} />;
}
