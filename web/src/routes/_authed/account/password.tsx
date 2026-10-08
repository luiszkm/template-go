import { createFileRoute } from "@tanstack/react-router";
import { ChangePassword } from "@/features/users/ChangePassword";

export const Route = createFileRoute("/_authed/account/password")({ component: ChangePassword });
