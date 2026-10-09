import { useQuery } from "@tanstack/react-query";
import { ApiError } from "@/api/result";
import { LoadError, Loading } from "@/components/States";
import { forbidden } from "@/lib/problems";
import { can, useMe } from "@/lib/session";
import { eventQuery } from "./api";
import { loadFailed } from "./copy";
import { actorLabel, formatWhen } from "./format";

function Json({ title, value }: { title: string; value: unknown }) {
  return (
    <section className="min-w-0 flex-1 space-y-1">
      <h2 className="text-sm font-semibold">{title}</h2>
      <pre className="overflow-auto rounded-md bg-muted p-3 text-xs">
        {value === null || value === undefined ? "—" : JSON.stringify(value, null, 2)}
      </pre>
    </section>
  );
}

export function AuditDetail({ id }: { id: number }) {
  const { data: me } = useMe();
  const allowed = can(me, "audit:read");
  const event = useQuery({ ...eventQuery(id), enabled: allowed });

  if (!allowed || (event.error instanceof ApiError && event.error.status === 403)) return <p>{forbidden}</p>;
  if (event.error instanceof ApiError && event.error.status === 404) return <p>Evento não encontrado.</p>;
  if (event.isPending) {
    return <Loading />;
  }
  if (event.isError) {
    return <LoadError message={loadFailed} retry={() => event.refetch()} />;
  }

  const e = event.data;
  const fields: Array<[string, string]> = [
    ["Ação", e.action],
    ["Quando", formatWhen(e.occurred_at)],
    ["Quem", actorLabel(e)],
    ["Recurso", `${e.resource_type} ${e.resource_id}`],
    ["IP", e.ip ?? "—"],
    ["Request ID", e.request_id],
  ];
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Evento {e.id}</h1>
      <dl className="grid max-w-xl grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        {fields.map(([label, value]) => (
          <div key={label} className="contents">
            <dt className="font-medium">{label}</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
      <div className="flex flex-col gap-4 md:flex-row">
        <Json title="Antes" value={e.before} />
        <Json title="Depois" value={e.after} />
      </div>
    </div>
  );
}
