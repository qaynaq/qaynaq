import { z } from "zod";

export const retriesSchema = z.object({
  initial_interval: z.string(),
  max_interval: z.string(),
  max_elapsed_time: z.string(),
});

export type Retries = z.infer<typeof retriesSchema>;

export const defaultRetries: Retries = {
  initial_interval: "500ms",
  max_interval: "1s",
  max_elapsed_time: "5s",
};
