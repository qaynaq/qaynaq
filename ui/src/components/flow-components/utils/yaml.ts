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
      ? (deepMerge(defaults, raw) as T)
      : (raw as T);
  const result = schema.safeParse(merged);
  // Validation failure (e.g. unfilled required field) must not wipe the user's
  // input. Return the merged shape; errors surface via the editor's errors prop.
  return result.success ? result.data : (merged as T);
}

export function serializeYaml<T>(config: T): string {
  return yaml.dump(stripEmpty(config) as object, { lineWidth: -1, noRefs: true });
}

// Recursively merges `override` onto `base`. Plain objects merge key-by-key;
// arrays and primitives from `override` replace `base` wholesale. Needed so
// that nested defaults (e.g. batch_policy.period) reappear after stripEmpty
// has dropped them during serialize.
function deepMerge(base: unknown, override: unknown): unknown {
  if (!isPlainObject(base) || !isPlainObject(override)) {
    return override === undefined ? base : override;
  }
  const out: Record<string, unknown> = { ...base };
  for (const [k, v] of Object.entries(override)) {
    out[k] = k in out ? deepMerge(out[k], v) : v;
  }
  return out;
}

function isPlainObject(v: unknown): v is Record<string, unknown> {
  return v !== null && typeof v === "object" && !Array.isArray(v);
}

// Recursively drops keys whose value is "" (and any object/array that becomes
// empty as a result). Bento parses optional fields like `conn_max_life_time`
// as durations and rejects "" - omitting the key lets Bento apply its default.
function stripEmpty(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(stripEmpty);
  }
  if (value !== null && typeof value === "object") {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      const cleaned = stripEmpty(v);
      if (cleaned === "") continue;
      if (
        cleaned !== null &&
        typeof cleaned === "object" &&
        !Array.isArray(cleaned) &&
        Object.keys(cleaned).length === 0
      ) {
        continue;
      }
      out[k] = cleaned;
    }
    return out;
  }
  return value;
}
