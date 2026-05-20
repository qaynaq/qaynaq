import { z } from "zod";
import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  CheckboxField,
  FieldWrapper,
  TextField,
} from "@/components/form-primitives";

export const hugotModelSchema = z.object({
  name: z.string().min(1, "Required"),
  path: z.string().min(1, "Required"),
  enable_download: z.boolean(),
  download_options: z.object({
    repository: z.string(),
    onnx_filepath: z.string(),
  }),
});

export type HugotModelConfig = z.infer<typeof hugotModelSchema>;

export const hugotModelDefaults: HugotModelConfig = {
  name: "",
  path: "./models",
  enable_download: false,
  download_options: {
    repository: "",
    onnx_filepath: "model.onnx",
  },
};

interface Props {
  value: HugotModelConfig;
  onChange: (next: HugotModelConfig) => void;
  errors?: Record<string, string>;
}

function shortId(): string {
  return Math.random().toString(36).slice(2, 8);
}

function generateName(repo: string): string {
  const tail = repo.split("/").pop() ?? "";
  const slug = tail
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
  return `${slug || "pipeline"}-${shortId()}`;
}

export function HugotModelFields({ value, onChange, errors }: Props) {
  const set = <K extends keyof HugotModelConfig>(
    k: K,
    v: HugotModelConfig[K],
  ) => onChange({ ...value, [k]: v });
  const setDownloadOption = <K extends keyof HugotModelConfig["download_options"]>(
    k: K,
    v: HugotModelConfig["download_options"][K],
  ) =>
    onChange({
      ...value,
      download_options: { ...value.download_options, [k]: v },
    });

  const pathDescription = value.enable_download
    ? "Parent directory for downloaded models. Qaynaq creates a sub-directory per repository inside it, so flows that use the same repository safely share files. ./models is a fine default."
    : "Path to your existing ONNX model file, or a directory containing model.onnx and tokenizer.json. Qaynaq does not download anything in this mode.";

  return (
    <div className="space-y-3">
      <FieldWrapper
        label="Name"
        description="Identifier for this pipeline instance. Used in logs and metrics, and must be unique within the flow."
        required
        error={errors?.name}
      >
        <div className="flex items-center gap-2">
          <Input
            value={value.name}
            onChange={(e) => set("name", e.target.value)}
            placeholder="my-embedder"
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-10 shrink-0"
            onClick={() => set("name", generateName(value.download_options.repository))}
            title="Generate a name for this pipeline"
          >
            <Sparkles className="h-3.5 w-3.5 mr-1" />
            Generate
          </Button>
        </div>
      </FieldWrapper>
      <TextField
        label="Path"
        description={pathDescription}
        required
        value={value.path}
        onChange={(v) => set("path", v)}
        error={errors?.path}
        placeholder={value.enable_download ? "./models" : "/path/to/model.onnx"}
      />
      <CheckboxField
        label="Enable Download"
        description="Fetch the model from HuggingFace into Path at startup. When off, Path must already contain the model."
        checked={value.enable_download}
        onChange={(c) => set("enable_download", c)}
      />
      {value.enable_download && (
        <div className="space-y-3 rounded-md border border-border bg-muted/30 p-3">
          <TextField
            label="Repository"
            description="HuggingFace model repository to download from. Format: owner/model-name."
            size="sm"
            value={value.download_options.repository}
            onChange={(v) => setDownloadOption("repository", v)}
            placeholder="sentence-transformers/all-MiniLM-L6-v2"
          />
          <TextField
            label="ONNX File Path"
            description="Path of the .onnx file inside the HuggingFace repository (not on your disk). Most repos publish a single model.onnx; only change this when a repo ships multiple variants like onnx/model_quantized.onnx."
            size="sm"
            value={value.download_options.onnx_filepath}
            onChange={(v) => setDownloadOption("onnx_filepath", v)}
            placeholder="model.onnx"
          />
        </div>
      )}
    </div>
  );
}
