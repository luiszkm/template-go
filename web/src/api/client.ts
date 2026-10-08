import createClient from "openapi-fetch";
import type { paths } from "./schema";

export const api = createClient<paths>({
  baseUrl: globalThis.location?.origin ?? "",
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
