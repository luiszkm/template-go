import { Button } from "@/components/ui/button";
import { useApiStatus } from "./useApiStatus";

export function ApiStatus() {
  const status = useApiStatus();

  if (status.isPending) {
    return (
      <div
        role="status"
        aria-label="Verificando a API"
        className="h-6 w-40 animate-pulse rounded bg-neutral-200"
      />
    );
  }
  if (status.isError) {
    return (
      <div className="flex items-center gap-3">
        <span className="text-red-700">API: offline</span>
        <Button variant="outline" size="sm" onClick={() => status.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  }
  return <span className="text-green-700">API: online</span>;
}
