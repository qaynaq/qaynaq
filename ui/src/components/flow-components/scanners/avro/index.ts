import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({
  avro_raw_json: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  avro_raw_json: false,
};

const scanner: FlowScanner<Config> = {
  id: "avro",
  name: "Avro",
  description: "Consume Avro OCF datum.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
