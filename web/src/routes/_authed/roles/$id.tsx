import { createFileRoute } from "@tanstack/react-router";
import { RoleDetail } from "@/features/rbac/RoleDetail";

export const Route = createFileRoute("/_authed/roles/$id")({ component: RoleDetailRoute });

function RoleDetailRoute() {
  const { id } = Route.useParams();
  return <RoleDetail id={id} />;
}
