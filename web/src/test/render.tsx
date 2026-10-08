import { render } from "@testing-library/react";
import { App } from "@/App";
import { makeQueryClient, makeTestRouter } from "@/router";

export const signedInAdmin = {
  id: "00000000-0000-4000-8000-000000000001",
  email: "admin@x.com",
  name: "Admin",
  permissions: ["users:activate", "users:create", "users:deactivate", "users:read", "users:update"],
};

export function renderAt(path: string) {
  const queryClient = makeQueryClient();
  const router = makeTestRouter(path, queryClient);
  return { router, ...render(<App router={router} queryClient={queryClient} />) };
}

export const meWith = (...permissions: string[]) => json(200, { ...signedInAdmin, permissions });

const isMe = (url: string) => new URL(url).pathname === "/api/v1/users/me";

export function stubFetch(...responses: Array<() => Promise<Response>>) {
  const calls: string[] = [];
  let served = 0;
  vi.stubGlobal("fetch", (input: Request | string) => {
    const url = typeof input === "string" ? input : input.url;
    calls.push(url);
    if (isMe(url)) return json(200, signedInAdmin)();
    const next = responses[Math.min(served, responses.length - 1)];
    served++;
    if (!next) throw new Error("no stubbed response");
    return next();
  });
  return calls;
}

export const json =
  (status: number, body: unknown, headers: Record<string, string> = {}) =>
  () =>
    Promise.resolve(
      new Response(status === 204 ? null : JSON.stringify(body), {
        status,
        headers: {
          "Content-Type": status >= 400 ? "application/problem+json" : "application/json",
          ...headers,
        },
      }),
    );

export const pending = () => new Promise<Response>(() => {});

export type Recorded = { method: string; path: string; search: string; body: unknown };
type Responder = () => Promise<Response>;

export function stubApi(routes: Record<string, Responder | Responder[]>) {
  const requests: Recorded[] = [];
  const served: Record<string, number> = {};
  vi.stubGlobal("fetch", async (input: Request) => {
    const url = new URL(input.url);
    const key = `${input.method} ${url.pathname}`;
    const text = await input.clone().text();
    requests.push({
      method: input.method,
      path: url.pathname,
      search: url.search,
      body: text ? JSON.parse(text) : undefined,
    });
    const route = routes[key];
    if (!route) throw new Error(`unstubbed request ${key}`);
    const list = Array.isArray(route) ? route : [route];
    const index = Math.min(served[key] ?? 0, list.length - 1);
    served[key] = (served[key] ?? 0) + 1;
    const respond = list[index];
    if (!respond) throw new Error(`no response for ${key}`);
    return respond();
  });
  return requests;
}
