import { createFileRoute } from "@tanstack/react-router";
import { RolesList } from "@/features/rbac/RolesList";

export const Route = createFileRoute("/_authed/roles/")({ component: RolesList });
