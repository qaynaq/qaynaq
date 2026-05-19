import { EmptyEditor } from "@/components/form-primitives";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({}).strict();
export type Config = z.infer<typeof configSchema>;

const scanner: FlowScanner<Config> = {
  id: "json_documents",
  name: "JSON Documents",
  description: "Consume a stream of one or more JSON documents.",
  configSchema,
  defaultConfig: {},
  Editor: EmptyEditor,
};

export default scanner;
