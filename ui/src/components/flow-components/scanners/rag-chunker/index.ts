import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

export const STRATEGIES = ["recursive", "token", "markdown"] as const;
export type Strategy = (typeof STRATEGIES)[number];

const configSchema = z
  .object({
    strategy: z.enum(STRATEGIES),
    chunk_size: z.number().int().min(1, "Required"),
    overlap: z.number().int().min(0),
  })
  .refine((c) => c.overlap < c.chunk_size, {
    message: "Overlap must be smaller than chunk size.",
    path: ["overlap"],
  });
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  strategy: "recursive",
  chunk_size: 1000,
  overlap: 200,
};

const scanner: FlowScanner<Config> = {
  id: "rag_chunker",
  name: "RAG Chunker",
  description:
    "Split text into overlapping chunks for RAG indexing, with recursive, token, or markdown strategies.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
