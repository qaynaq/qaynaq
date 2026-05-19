import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z
  .object({
    custom_delimiter: z.string(),
    parse_header_row: z.boolean(),
    lazy_quotes: z.boolean(),
    continue_on_error: z.boolean(),
    expected_headers: z.array(z.string()),
    expected_number_of_fields: z.number().int().min(0),
  })
  .refine(
    (c) => c.expected_headers.length === 0 || c.parse_header_row,
    {
      message: "Requires Parse Header Row to be enabled.",
      path: ["expected_headers"],
    },
  );
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  custom_delimiter: "",
  parse_header_row: true,
  lazy_quotes: false,
  continue_on_error: false,
  expected_headers: [],
  expected_number_of_fields: 0,
};

const scanner: FlowScanner<Config> = {
  id: "csv",
  name: "CSV",
  description: "Parse comma-separated values row by row.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
