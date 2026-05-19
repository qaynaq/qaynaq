import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({
  into: z.record(z.unknown()),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  into: { to_the_end: {} },
};

const scanner: FlowScanner<Config> = {
  id: "skip_bom",
  name: "Skip BOM",
  description:
    "Strip a leading byte order mark before delegating to a child scanner.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
