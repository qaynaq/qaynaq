import type { FlowScanner } from "./types";

const modules = import.meta.glob<{ default: FlowScanner<unknown> }>(
  "./*/index.ts",
  { eager: true },
);

const all: FlowScanner<unknown>[] = Object.values(modules).map((m) => m.default);
const byId = new Map<string, FlowScanner<unknown>>();
for (const s of all) byId.set(s.id, s);

export function getScanner(id: string): FlowScanner<unknown> | undefined {
  return byId.get(id);
}

export function listScanners(): FlowScanner<unknown>[] {
  return all;
}
