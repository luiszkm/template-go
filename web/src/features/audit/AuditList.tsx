import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { forbidden } from "@/lib/problems";
import { ApiError, can, useMe } from "@/lib/session";
import { type AuditFilters, actionsQuery, actorLabel, fetchEvents, formatWhen, loadFailed } from "./api";

export function AuditList({
  filters,
  onFiltersChange,
}: {
  filters: AuditFilters;
  onFiltersChange: (filters: AuditFilters) => void;
}) {
  const { data: me } = useMe();
  const allowed = can(me, "audit:read");
  const actions = useQuery({ ...actionsQuery, enabled: allowed });
  const events = useInfiniteQuery({
    queryKey: ["audit-events", filters],
    enabled: allowed,
    initialPageParam: undefined as number | undefined,
    queryFn: ({ pageParam }) => fetchEvents(filters, pageParam),
    getNextPageParam: (last) => last.next ?? undefined,
  });

  if (!allowed || (events.error instanceof ApiError && events.error.status === 403))
    return <p>{forbidden}</p>;

  function change(next: AuditFilters) {
    onFiltersChange(Object.fromEntries(Object.entries(next).filter(([, v]) => v)) as AuditFilters);
  }

  const filtered = Object.values(filters).some(Boolean);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Auditoria</h1>
      <div className="flex flex-wrap items-end gap-4">
        <div className="space-y-1">
          <Label htmlFor="action">Ação</Label>
          <select
            id="action"
            className="h-9 rounded-md border border-neutral-300 px-2 text-sm"
            value={filters.action ?? ""}
            onChange={(e) => change({ ...filters, action: e.target.value })}
          >
            <option value="">Todas</option>
            {(actions.data ?? []).map((a) => (
              <option key={a} value={a}>
                {a}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <Label htmlFor="de">De</Label>
          <Input
            id="de"
            type="date"
            value={filters.de ?? ""}
            onChange={(e) => change({ ...filters, de: e.target.value })}
          />
        </div>
        <div className="space-y-1">
          <Label htmlFor="ate">Até</Label>
          <Input
            id="ate"
            type="date"
            value={filters.ate ?? ""}
            onChange={(e) => change({ ...filters, ate: e.target.value })}
          />
        </div>
        {filtered && (
          <Button variant="outline" onClick={() => change({})}>
            Limpar filtros
          </Button>
        )}
      </div>
      <Events events={events} filters={filters} onFilter={change} />
    </div>
  );
}

function Events({
  events,
  filters,
  onFilter,
}: {
  events: ReturnType<typeof useInfiniteQuery<Awaited<ReturnType<typeof fetchEvents>>>>;
  filters: AuditFilters;
  onFilter: (filters: AuditFilters) => void;
}) {
  if (events.isPending) {
    return (
      <p role="status" className="text-neutral-600">
        Carregando…
      </p>
    );
  }
  if (events.isError) {
    return (
      <div className="space-y-2">
        <p>{loadFailed}</p>
        <Button variant="outline" onClick={() => events.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  }
  const items = events.data.pages.flatMap((p) => p.items ?? []);
  if (items.length === 0) return <p>Nenhum evento encontrado.</p>;
  return (
    <div className="space-y-3">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Quando</TableHead>
            <TableHead>Ação</TableHead>
            <TableHead>Quem</TableHead>
            <TableHead>Recurso</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((event) => (
            <TableRow key={event.id}>
              <TableCell>
                <Link to="/audit/$id" params={{ id: String(event.id) }} className="underline">
                  {formatWhen(event.occurred_at)}
                </Link>
              </TableCell>
              <TableCell>{event.action}</TableCell>
              <TableCell>
                {event.actor ? (
                  <button
                    type="button"
                    className="underline"
                    onClick={() => onFilter({ ...filters, actor_id: event.actor?.id })}
                  >
                    {actorLabel(event)}
                  </button>
                ) : (
                  actorLabel(event)
                )}
              </TableCell>
              <TableCell>
                <button
                  type="button"
                  className="underline"
                  onClick={() =>
                    onFilter({
                      ...filters,
                      resource_type: event.resource_type,
                      resource_id: event.resource_id,
                    })
                  }
                >
                  {event.resource_type} {event.resource_id}
                </button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {events.hasNextPage && (
        <Button variant="outline" disabled={events.isFetchingNextPage} onClick={() => events.fetchNextPage()}>
          Mais antigos
        </Button>
      )}
    </div>
  );
}
