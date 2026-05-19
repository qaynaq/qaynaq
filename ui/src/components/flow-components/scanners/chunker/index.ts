import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({
  size: z.number().int().min(1, "Required"),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  size: 1024,
};

const scanner: FlowScanner<Config> = {
  id: "chunker",
  name: "Chunker",
  description: "Split the stream into fixed-size byte chunks.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
