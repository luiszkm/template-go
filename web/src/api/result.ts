import type { components } from "./schema";

type Problem = components["schemas"]["Problem"];

type Answer<T> = { data?: T; error?: Problem; response: Response };

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly problem?: Problem,
  ) {
    super(problem?.detail ?? `API answered ${status}`);
  }
}

export function dataOf<T>({ data, error, response }: Answer<T>): T {
  if (data === undefined) throw new ApiError(response.status, error);
  return data;
}

export function noContent({ error, response }: Answer<unknown>) {
  if (response.status !== 204) throw new ApiError(response.status, error);
}
