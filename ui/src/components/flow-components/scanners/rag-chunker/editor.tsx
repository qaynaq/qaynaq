import { NumberField, SelectField } from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import { STRATEGIES, type Config, type Strategy } from ".";

const STRATEGY_OPTIONS = [
  { value: "recursive", label: "Recursive (paragraph → line → sentence)" },
  { value: "token", label: "Token (4-character estimate)" },
  { value: "markdown", label: "Markdown (split on headings)" },
];

export default function RagChunkerScannerEditor({
  value,
  onChange,
  errors,
}: ScannerEditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  const unit = value.strategy === "token" ? "tokens" : "characters";
  return (
    <div className="space-y-3">
      <SelectField
        label="Strategy"
        description="How the text is broken into chunks."
        size="sm"
        value={value.strategy}
        onChange={(v) =>
          set(
            "strategy",
            (STRATEGIES.includes(v as Strategy) ? v : "recursive") as Strategy,
          )
        }
        options={STRATEGY_OPTIONS}
      />
      <NumberField
        label="Chunk Size"
        description={`Maximum ${unit} per chunk. The chunker prefers coarser boundaries (paragraphs, then lines, ...) and only falls back to character-level splitting when nothing else fits.`}
        required
        size="sm"
        min={1}
        value={value.chunk_size}
        onChange={(v) => set("chunk_size", v)}
        error={errors?.chunk_size}
      />
      <NumberField
        label="Overlap"
        description={`How many ${unit} of the previous chunk are carried into the start of the next. Aligned to the nearest boundary. A typical value is around 20% of the chunk size.`}
        size="sm"
        min={0}
        value={value.overlap}
        onChange={(v) => set("overlap", v)}
        error={errors?.overlap}
      />
    </div>
  );
}
