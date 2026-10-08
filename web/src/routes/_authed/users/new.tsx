import { createFileRoute } from "@tanstack/react-router";
import { UserForm } from "@/features/users/UserForm";

export const Route = createFileRoute("/_authed/users/new")({ component: UserForm });
