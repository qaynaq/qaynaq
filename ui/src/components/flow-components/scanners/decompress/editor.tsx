import { SelectField } from "@/components/form-primitives";
import { ScannerField } from "@/components/form-primitives/scanner-field";
import type { ScannerEditorProps } from "../types";
import {
  DECOMPRESS_ALGORITHMS,
  type Config,
  type DecompressAlgorithm,
} from ".";

export default function DecompressScannerEditor({
  value,
  onChange,
  errors,
}: ScannerEditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-3">
      <SelectField
        label="Algorithm"
        description="Compression format of the source stream."
        size="sm"
        value={value.algorithm}
        onChange={(v) => set("algorithm", v as DecompressAlgorithm)}
        options={DECOMPRESS_ALGORITHMS as unknown as string[]}
        error={errors?.algorithm}
      />
      <ScannerField
        label="Into"
        description="The scanner that receives the decompressed bytes."
        size="sm"
        value={value.into}
        onChange={(v) => set("into", v)}
      />
    </div>
  );
}
