import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

export const XML_OPERATORS = ["to_json"] as const;
export type XmlOperator = (typeof XML_OPERATORS)[number];

const configSchema = z.object({
  operator: z.enum(XML_OPERATORS),
  cast: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  operator: "to_json",
  cast: false,
};

const scanner: FlowScanner<Config> = {
  id: "xml_documents",
  name: "XML Documents",
  description:
    "Consume a stream of XML documents, optionally converting each one to JSON.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
