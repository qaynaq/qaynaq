import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";

const configSchema = z.object({
  mapping: z.string().min(1, "Required"),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { mapping: "" };

const component: FlowComponent<Config> = {
  id: "mapping",
  name: "Mapping",
  category: "processor",
  configSchema,
  defaultConfig,
  parse: (s) => ({ mapping: s ?? "" }),
  serialize: (c) => c.mapping ?? "",
  toListItem: (c) => c.mapping ?? "",
  fromListItem: (raw) => ({ mapping: typeof raw === "string" ? raw : "" }),
  Editor: lazy(() => import("./editor")),
};

export default component;
