import { Button } from "@/components/ui/button";

export function Loading() {
  return (
    <p role="status" className="text-muted-foreground">
      Carregando…
    </p>
  );
}

export function LoadError({ message, retry }: { message: string; retry: () => void }) {
  return (
    <div className="space-y-2">
      <p>{message}</p>
      <Button variant="outline" onClick={retry}>
        Tentar novamente
      </Button>
    </div>
  );
}
