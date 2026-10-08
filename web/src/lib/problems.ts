import type { components } from "@/api/schema";

type Problem = components["schemas"]["Problem"];

export type FieldErrors = Partial<Record<string, string>>;

export function fieldErrors(problem: Problem | undefined): FieldErrors {
  const errors: FieldErrors = {};
  for (const detail of problem?.errors ?? []) {
    const field = detail.location?.replace(/^body\./, "");
    if (field && !errors[field]) errors[field] = detail.message;
  }
  return errors;
}

export const forbidden = "Você não tem permissão para acessar esta página.";
