import { render } from "@testing-library/react";
import { App } from "@/App";
import { makeQueryClient, makeTestRouter } from "@/router";

/** Renders the whole app at path, with a fresh router and query cache. */
export function renderAt(path: string) {
  return render(<App router={makeTestRouter(path)} queryClient={makeQueryClient()} />);
}

/** Stubs fetch; each call returns the next response (the last one repeats). Returns the requested URLs. */
export function stubFetch(...responses: Array<() => Promise<Response>>) {
  const calls: string[] = [];
  vi.stubGlobal("fetch", (input: Request | string) => {
    calls.push(typeof input === "string" ? input : input.url);
    const next = responses[Math.min(calls.length - 1, responses.length - 1)];
    if (!next) throw new Error("no stubbed response");
    return next();
  });
  return calls;
}

export const json = (status: number, body: unknown) => () =>
  Promise.resolve(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": status >= 400 ? "application/problem+json" : "application/json" },
    }),
  );
