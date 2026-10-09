import type { AuditEvent } from "./api";

export function formatWhen(iso: string) {
  const d = new Date(iso);
  const two = (n: number) => String(n).padStart(2, "0");
  return `${two(d.getDate())}/${two(d.getMonth() + 1)}/${d.getFullYear()} ${two(d.getHours())}:${two(d.getMinutes())}:${two(d.getSeconds())}`;
}

export function actorLabel(event: AuditEvent) {
  return event.actor?.email ?? "Sistema";
}
