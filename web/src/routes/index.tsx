import { createFileRoute } from "@tanstack/react-router";
import { ApiStatus } from "@/features/status/ApiStatus";

export const Route = createFileRoute("/")({
  component: Home,
});

function Home() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Template</h1>
      <ApiStatus />
    </div>
  );
}
