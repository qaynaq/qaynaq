import * as yaml from "js-yaml";
import type { z } from "zod";

export function parseYaml<T>(
  schema: z.ZodType<T>,
  yamlStr: string,
  defaults: T,
): T {
  const trimmed = yamlStr?.trim();
  if (!trimmed) return structuredClone(defaults);
  let raw: unknown;
  try {
    raw = yaml.load(trimmed);
  } catch {
    return structuredClone(defaults);
  }
  if (raw === null || raw === undefined) return structuredClone(defaults);
  const merged =
    typeof raw === "object" && !Array.isArray(raw)
      ? { ...defaults, ...(raw as object) }
      : raw;
  const result = schema.safeParse(merged);
  // Validation failure (e.g. unfilled required field) must not wipe the user's
  // input. Return the merged shape; errors surface via the editor's errors prop.
  return result.success ? result.data : (merged as T);
}

export function serializeYaml<T>(config: T): string {
  return yaml.dump(config as object, { lineWidth: -1, noRefs: true });
}
