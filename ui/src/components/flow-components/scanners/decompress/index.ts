import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

export const DECOMPRESS_ALGORITHMS = [
  "gzip",
  "pgzip",
  "zlib",
  "bzip2",
  "flate",
  "snappy",
  "lz4",
  "zstd",
] as const;
export type DecompressAlgorithm = (typeof DECOMPRESS_ALGORITHMS)[number];

const configSchema = z.object({
  algorithm: z.enum(DECOMPRESS_ALGORITHMS),
  into: z.record(z.unknown()),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  algorithm: "gzip",
  into: { to_the_end: {} },
};

const scanner: FlowScanner<Config> = {
  id: "decompress",
  name: "Decompress",
  description:
    "Decompress the byte stream before delegating to a child scanner.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
