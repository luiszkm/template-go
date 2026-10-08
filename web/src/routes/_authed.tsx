import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { ApiError, meQuery } from "@/features/users/session";
import { UserMenu } from "@/features/users/UserMenu";

export const Route = createFileRoute("/_authed")({
  beforeLoad: async ({ context, location }) => {
    try {
      await context.queryClient.ensureQueryData(meQuery);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        throw redirect({ to: "/login", search: { redirect: location.href } });
      }
      throw error;
    }
  },
  component: AuthedLayout,
});

function AuthedLayout() {
  return (
    <div className="space-y-6">
      <UserMenu />
      <Outlet />
    </div>
  );
}
