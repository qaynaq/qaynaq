import type { ComponentCategory, FlowComponent } from "./types";

const inputs = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./inputs/*/index.ts",
  { eager: true },
);
const processors = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./processors/*/index.ts",
  { eager: true },
);
const outputs = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./outputs/*/index.ts",
  { eager: true },
);
const caches = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./caches/*/index.ts",
  { eager: true },
);
const buffers = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./buffers/*/index.ts",
  { eager: true },
);
const rateLimits = import.meta.glob<{ default: FlowComponent<unknown> }>(
  "./rate-limits/*/index.ts",
  { eager: true },
);

const all: FlowComponent<unknown>[] = [
  ...Object.values(inputs),
  ...Object.values(processors),
  ...Object.values(outputs),
  ...Object.values(caches),
  ...Object.values(buffers),
  ...Object.values(rateLimits),
].map((m) => m.default);

const byCategoryAndId = new Map<string, FlowComponent<unknown>>();
for (const c of all) {
  byCategoryAndId.set(`${c.category}:${c.id}`, c);
}

export function getComponent(
  category: ComponentCategory,
  id: string,
): FlowComponent<unknown> | undefined {
  return byCategoryAndId.get(`${category}:${id}`);
}

export function listComponents(
  category: ComponentCategory,
): FlowComponent<unknown>[] {
  return all.filter((c) => c.category === category);
}

export function listAll(): FlowComponent<unknown>[] {
  return all;
}
