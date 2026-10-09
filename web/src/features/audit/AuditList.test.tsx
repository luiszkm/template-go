import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const actorId = "00000000-0000-4000-8000-0000000000a1";
const event = (id: number, overrides: Record<string, unknown> = {}) => ({
  id,
  occurred_at: "2026-10-08T17:30:05Z",
  action: "user.created",
  actor: { id: actorId, email: "a@x.com" },
  resource_type: "user",
  resource_id: "42",
  ip: "10.0.0.7",
  request_id: "req-1",
  ...overrides,
});
const page = (items: unknown[], next: number | null = null) => json(200, { items, next });
const catalogue = json(200, { items: ["role.created", "user.created"] });
const reader = meWith("audit:read");
const eventsPath = "/api/v1/audit/events";
const forbidden = "Você não tem permissão para acessar esta página.";

function eventRequests(requests: Array<{ path: string; search: string }>) {
  return requests.filter((r) => r.path === eventsPath).map((r) => new URLSearchParams(r.search));
}

describe("AuditList on /audit", () => {
  it("shows the events table", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(2), event(1, { actor: null })]),
    });
    renderAt("/audit");
    const table = await screen.findByRole("table");
    expect(
      within(table)
        .getAllByRole("columnheader")
        .map((h) => h.textContent),
    ).toEqual(["Quando", "Ação", "Quem", "Recurso"]);
    const rows = within(table)
      .getAllByRole("row")
      .slice(1)
      .map((r) =>
        within(r)
          .getAllByRole("cell")
          .map((c) => c.textContent),
      );
    expect(rows).toEqual([
      ["08/10/2026 14:30:05", "user.created", "a@x.com", "user 42"],
      ["08/10/2026 14:30:05", "user.created", "Sistema", "user 42"],
    ]);
  });

  it("shows empty state", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([]),
    });
    renderAt("/audit");
    expect(await screen.findByText("Nenhum evento encontrado.")).toBeInTheDocument();
  });

  it("shows loading", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: pending,
    });
    renderAt("/audit");
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows error with retry", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: [json(500, { status: 500 }), page([event(1)])],
    });
    renderAt("/audit");
    expect(await screen.findByText("Não foi possível carregar a auditoria.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByRole("table")).toBeInTheDocument();
    expect(eventRequests(requests)).toHaveLength(2);
  });

  it("forbidden without audit:read", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith("users:read") });
    renderAt("/audit");
    expect(await screen.findByText(forbidden)).toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/audit"))).toHaveLength(0);
  });

  it("forbidden on API 403", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: json(403, { status: 403 }),
    });
    renderAt("/audit");
    expect(await screen.findByText(forbidden)).toBeInTheDocument();
  });

  it("loads older events", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: [page([event(9), event(8)], 7), page([event(7, { action: "role.created" })])],
    });
    renderAt("/audit");
    await userEvent.click(await screen.findByRole("button", { name: "Mais antigos" }));
    await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(4));
    const actions = screen
      .getAllByRole("row")
      .slice(1)
      .map((r) => within(r).getAllByRole("cell")[1]?.textContent);
    expect(actions).toEqual(["user.created", "user.created", "role.created"]);
    expect(eventRequests(requests)[1]?.get("before")).toBe("7");
    expect(screen.queryByRole("button", { name: "Mais antigos" })).not.toBeInTheDocument();
  });

  it("filters by action", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(1)]),
    });
    const { router } = renderAt("/audit");
    const select = await screen.findByLabelText("Ação");
    await waitFor(() =>
      expect(Array.from((select as HTMLSelectElement).options).map((o) => o.textContent)).toEqual([
        "Todas",
        "role.created",
        "user.created",
      ]),
    );
    await userEvent.selectOptions(select, "user.created");
    await waitFor(() => expect(router.state.location.search).toEqual({ action: "user.created" }));
    await waitFor(() => expect(eventRequests(requests).at(-1)?.get("action")).toBe("user.created"));

    await userEvent.selectOptions(select, "Todas");
    await waitFor(() => expect(router.state.location.search).toEqual({}));
    await waitFor(() => expect(eventRequests(requests).at(-1)?.has("action")).toBe(false));
  });

  it("filters by period", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(1)]),
    });
    renderAt("/audit");
    await userEvent.type(await screen.findByLabelText("De"), "2026-10-01");
    await userEvent.type(screen.getByLabelText("Até"), "2026-10-02");
    await waitFor(() => {
      const last = eventRequests(requests).at(-1);
      expect(last?.get("from")).toBe("2026-10-01T03:00:00.000Z");
      expect(last?.get("to")).toBe("2026-10-03T03:00:00.000Z");
    });
  });

  it("filters by actor and resource", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(1)]),
    });
    const { router } = renderAt("/audit");
    await userEvent.click(await screen.findByRole("button", { name: "a@x.com" }));
    await waitFor(() => expect(router.state.location.search).toEqual({ actor_id: actorId }));
    await waitFor(() => expect(eventRequests(requests).at(-1)?.get("actor_id")).toBe(actorId));

    await userEvent.click(await screen.findByRole("button", { name: "Limpar filtros" }));
    await waitFor(() => expect(router.state.location.search).toEqual({}));
    await waitFor(() => expect(eventRequests(requests).at(-1)?.toString()).toBe(""));

    await userEvent.click(await screen.findByRole("button", { name: "user 42" }));
    await waitFor(() =>
      expect(router.state.location.search).toEqual({ resource_type: "user", resource_id: "42" }),
    );
    await waitFor(() => {
      const last = eventRequests(requests).at(-1);
      expect(last?.get("resource_type")).toBe("user");
      expect(last?.get("resource_id")).toBe("42");
    });
    expect(screen.getByRole("button", { name: "Limpar filtros" })).toBeInTheDocument();
  });

  it("reads filters from the URL", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(1)]),
    });
    renderAt(`/audit?action=user.created&actor_id=${actorId}&resource_type=user&resource_id=42`);
    await screen.findByRole("table");
    const first = eventRequests(requests)[0];
    expect(first?.get("action")).toBe("user.created");
    expect(first?.get("actor_id")).toBe(actorId);
    expect(first?.get("resource_type")).toBe("user");
    expect(first?.get("resource_id")).toBe("42");
  });

  it("reads the period from the URL", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(1)]),
    });
    renderAt("/audit?de=2026-10-01&ate=2026-10-02");
    await screen.findByRole("table");
    const first = eventRequests(requests)[0];
    expect(first?.get("from")).toBe("2026-10-01T03:00:00.000Z");
    expect(first?.get("to")).toBe("2026-10-03T03:00:00.000Z");
    expect(screen.getByLabelText("De")).toHaveValue("2026-10-01");
    expect(screen.getByLabelText("Até")).toHaveValue("2026-10-02");
  });

  it("opens an event", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      "GET /api/v1/audit/actions": catalogue,
      [`GET ${eventsPath}`]: page([event(5)]),
      "GET /api/v1/audit/events/5": json(200, { ...event(5), before: null, after: null }),
    });
    const { router } = renderAt("/audit");
    await userEvent.click(await screen.findByRole("link", { name: "08/10/2026 14:30:05" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/audit/5"));
  });
});
