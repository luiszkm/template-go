import { Button } from "@/components/ui/button";
import { loadFailed } from "./copy";

export function Loading() {
  return (
    <p role="status" className="text-neutral-600">
      Carregando…
    </p>
  );
}

export function LoadError({ retry }: { retry: () => void }) {
  return (
    <div className="space-y-2">
      <p>{loadFailed}</p>
      <Button variant="outline" onClick={retry}>
        Tentar novamente
      </Button>
    </div>
  );
}
