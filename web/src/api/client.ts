// The only module allowed to talk to the backend. Types come from src/api/schema.d.ts,
// generated from app/openapi.json by `npm run gen` - never edit the schema by hand.
import createClient from "openapi-fetch";
import type { paths } from "./schema";

export const api = createClient<paths>({
  // Same origin as the page (AD-005); absolute so it also works under jsdom.
  baseUrl: globalThis.location?.origin ?? "",
  // Resolve fetch at call time so tests can stub it.
  fetch: (request) => globalThis.fetch(request),
});

type UnauthorizedListener = () => void;

let unauthorizedListener: UnauthorizedListener | undefined;

export function onUnauthorized(listener: UnauthorizedListener) {
  unauthorizedListener = listener;
  return () => {
    if (unauthorizedListener === listener) unauthorizedListener = undefined;
  };
}

const isSignIn = (request: Request) =>
  request.method === "POST" && new URL(request.url).pathname === "/api/v1/users/session";

api.use({
  onResponse({ request, response }) {
    if (response.status === 401 && !isSignIn(request)) unauthorizedListener?.();
    return response;
  },
});
