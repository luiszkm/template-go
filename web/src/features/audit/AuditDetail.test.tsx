import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, meWith, pending, renderAt, stubApi } from "@/test/render";

const path = "/api/v1/audit/events/5";
const detail = {
  id: 5,
  occurred_at: "2026-10-08T17:30:05Z",
  action: "user.updated",
  actor: { id: "00000000-0000-4000-8000-0000000000a1", email: "a@x.com" },
  resource_type: "user",
  resource_id: "42",
  ip: "10.0.0.7",
  request_id: "req-1",
  before: { name: "Ana" },
  after: { name: "Bia" },
};
const reader = meWith("audit:read");

function fieldValue(label: string) {
  const term = screen.getByText(label, { selector: "dt" });
  return term.nextElementSibling?.textContent;
}

function block(title: string) {
  return screen.getByRole("heading", { name: title }).nextElementSibling?.textContent;
}

describe("AuditDetail on /audit/$id", () => {
  it("shows the event", async () => {
    stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: json(200, detail) });
    renderAt("/audit/5");
    await screen.findByRole("heading", { name: "Antes" });
    expect(fieldValue("Ação")).toBe("user.updated");
    expect(fieldValue("Quando")).toBe("08/10/2026 14:30:05");
    expect(fieldValue("Quem")).toBe("a@x.com");
    expect(fieldValue("Recurso")).toBe("user 42");
    expect(fieldValue("IP")).toBe("10.0.0.7");
    expect(fieldValue("Request ID")).toBe("req-1");
    expect(block("Antes")).toBe('{\n  "name": "Ana"\n}');
    expect(block("Depois")).toBe('{\n  "name": "Bia"\n}');
  });

  it("shows a dash for a null block", async () => {
    stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: json(200, { ...detail, before: null }) });
    renderAt("/audit/5");
    await screen.findByRole("heading", { name: "Antes" });
    expect(block("Antes")).toBe("—");
  });

  it("shows fallbacks for a missing ip and actor", async () => {
    stubApi({
      "GET /api/v1/users/me": reader,
      [`GET ${path}`]: json(200, { ...detail, ip: null, actor: null }),
    });
    renderAt("/audit/5");
    await screen.findByRole("heading", { name: "Antes" });
    expect(fieldValue("IP")).toBe("—");
    expect(fieldValue("Quem")).toBe("Sistema");
  });

  it("requests the path id", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: json(200, detail) });
    renderAt("/audit/5");
    await screen.findByRole("heading", { name: "Antes" });
    expect(requests.some((r) => r.path === path)).toBe(true);
  });

  it("shows not found", async () => {
    stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: json(404, { status: 404 }) });
    renderAt("/audit/5");
    expect(await screen.findByText("Evento não encontrado.")).toBeInTheDocument();
  });

  it("shows loading", async () => {
    stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: pending });
    renderAt("/audit/5");
    expect(await screen.findByRole("status")).toBeInTheDocument();
  });

  it("shows error with retry", async () => {
    const requests = stubApi({
      "GET /api/v1/users/me": reader,
      [`GET ${path}`]: [json(500, { status: 500 }), json(200, detail)],
    });
    renderAt("/audit/5");
    expect(await screen.findByText("Não foi possível carregar a auditoria.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByRole("heading", { name: "Antes" })).toBeInTheDocument();
    expect(requests.filter((r) => r.path === path)).toHaveLength(2);
  });

  it("forbidden without audit:read", async () => {
    const requests = stubApi({ "GET /api/v1/users/me": meWith("users:read") });
    renderAt("/audit/5");
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
    expect(requests.filter((r) => r.path.startsWith("/api/v1/audit"))).toHaveLength(0);
  });

  it("forbidden on API 403", async () => {
    stubApi({ "GET /api/v1/users/me": reader, [`GET ${path}`]: json(403, { status: 403 }) });
    renderAt("/audit/5");
    expect(await screen.findByText("Você não tem permissão para acessar esta página.")).toBeInTheDocument();
  });
});
