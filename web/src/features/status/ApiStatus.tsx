import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useApiStatus } from "./api";

export function ApiStatus() {
  const status = useApiStatus();

  if (status.isPending) {
    return <Skeleton role="status" aria-label="Verificando a API" className="h-6 w-40" />;
  }
  if (status.isError) {
    return (
      <div className="flex items-center gap-3">
        <span className="text-destructive">API: offline</span>
        <Button variant="outline" size="sm" onClick={() => status.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  }
  return <span className="text-success">API: online</span>;
}
