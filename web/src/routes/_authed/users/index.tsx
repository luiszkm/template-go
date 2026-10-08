import { createFileRoute } from "@tanstack/react-router";
import { UsersList } from "@/features/users/UsersList";

export const Route = createFileRoute("/_authed/users/")({ component: UsersList });
