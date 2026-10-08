import { createFileRoute } from "@tanstack/react-router";
import { UserRoles } from "@/features/rbac/UserRoles";
import { UserDetail } from "@/features/users/UserDetail";

export const Route = createFileRoute("/_authed/users/$id")({ component: UserDetailRoute });

function UserDetailRoute() {
  const { id } = Route.useParams();
  return (
    <div className="space-y-8">
      <UserDetail id={id} />
      <UserRoles id={id} />
    </div>
  );
}
