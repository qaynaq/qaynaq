import { EmptyEditor } from "@/components/form-primitives";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({}).strict();
export type Config = z.infer<typeof configSchema>;

const scanner: FlowScanner<Config> = {
  id: "tar",
  name: "Tar",
  description: "Read each file in a tar archive as a separate message.",
  configSchema,
  defaultConfig: {},
  Editor: EmptyEditor,
};

export default scanner;
