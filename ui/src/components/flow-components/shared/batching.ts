import { z } from "zod";

export const batchingSchema = z.object({
  count: z.number().int().min(0),
  byte_size: z.number().int().min(0),
  period: z.string(),
  jitter: z.number().min(0),
  check: z.string(),
});

export type Batching = z.infer<typeof batchingSchema>;

export const defaultBatching: Batching = {
  count: 0,
  byte_size: 0,
  period: "",
  jitter: 0,
  check: "",
};
