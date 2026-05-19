import { ScannerField } from "@/components/form-primitives/scanner-field";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function SkipBomScannerEditor({
  value,
  onChange,
}: ScannerEditorProps<Config>) {
  return (
    <div className="space-y-3">
      <ScannerField
        label="Into"
        description="The scanner that receives the stream once the BOM has been stripped."
        size="sm"
        value={value.into}
        onChange={(v) => onChange({ ...value, into: v })}
      />
    </div>
  );
}
