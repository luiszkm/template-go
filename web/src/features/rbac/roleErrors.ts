import { ApiError } from "@/api/result";
import { type FieldErrors, fieldErrors } from "@/lib/problems";
import { nameInUse } from "./copy";

export function roleErrors(error: Error): FieldErrors {
  if (!(error instanceof ApiError)) return {};
  if (error.status === 409) return { name: nameInUse };
  if (error.status === 422) return fieldErrors(error.problem);
  return {};
}
