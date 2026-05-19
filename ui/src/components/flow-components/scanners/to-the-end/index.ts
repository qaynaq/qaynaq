import { EmptyEditor } from "@/components/form-primitives";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({}).strict();
export type Config = z.infer<typeof configSchema>;

const scanner: FlowScanner<Config> = {
  id: "to_the_end",
  name: "To The End",
  description:
    "Read the entire stream as a single message. Only safe when the stream has a clear end and fits in memory.",
  configSchema,
  defaultConfig: {},
  Editor: EmptyEditor,
};

export default scanner;
