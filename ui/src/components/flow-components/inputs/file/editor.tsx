import {
  ArrayField,
  CheckboxField,
  ScannerField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";
import type { Config } from ".";

export default function FileInputEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ArrayField
        label="Paths"
        description="Files to consume sequentially. Glob patterns are supported, including super globs (double star)."
        value={value.paths}
        onChange={(v) => set("paths", v)}
        error={errors?.paths}
      />
      <ScannerField
        label="Scanner"
        description="How the byte stream from each file is broken into messages."
        value={value.scanner}
        onChange={(v) => set("scanner", v)}
      />
      <CheckboxField
        label="Delete On Finish"
        description="Delete each file from disk once it has been fully consumed."
        checked={value.delete_on_finish}
        onChange={(c) => set("delete_on_finish", c)}
      />
      <CheckboxField
        label="Auto Replay Nacks"
        description="Whether messages rejected at the output level should be automatically replayed."
        checked={value.auto_replay_nacks}
        onChange={(c) => set("auto_replay_nacks", c)}
      />
    </div>
  );
}
