import { createFileRoute } from "@tanstack/react-router";
import { UserDetail } from "@/features/users/UserDetail";

export const Route = createFileRoute("/_authed/users/$id")({ component: UserDetailRoute });

function UserDetailRoute() {
  const { id } = Route.useParams();
  return <UserDetail id={id} />;
}
