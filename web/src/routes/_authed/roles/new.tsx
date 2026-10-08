import { createFileRoute } from "@tanstack/react-router";
import { RoleForm } from "@/features/rbac/RoleForm";

export const Route = createFileRoute("/_authed/roles/new")({ component: RoleForm });
