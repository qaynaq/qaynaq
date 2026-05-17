import {
  TextField,
  NumberField,
  SelectField,
  ArrayField,
  CodeField,
} from "@/components/form-primitives";
import { BatchingEditor } from "../../shared/batching-editor";
import { SQL_DRIVERS } from "../sql-raw";
import type { Batching } from "../../shared/batching";
import type { EditorProps } from "../../types";

interface Config {
  driver: (typeof SQL_DRIVERS)[number];
  dsn: string;
  table: string;
  columns: string[];
  args_mapping: string;
  suffix: string;
  max_in_flight: number;
  batching: Batching;
}

export default function SqlInsertEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <SelectField label="Driver" required value={value.driver} onChange={(v) => set("driver", v as Config["driver"])} options={SQL_DRIVERS as unknown as string[]} />
      <TextField label="DSN" required value={value.dsn} onChange={(v) => set("dsn", v)} error={errors?.dsn} />
      <TextField label="Table" required value={value.table} onChange={(v) => set("table", v)} error={errors?.table} />
      <ArrayField label="Columns" required value={value.columns} onChange={(v) => set("columns", v)} />
      <CodeField label="Args Mapping" description="A Bloblang mapping evaluating to an array of values matching the columns." required value={value.args_mapping} onChange={(v) => set("args_mapping", v)} error={errors?.args_mapping} />
      <CodeField label="Suffix" description="Optional suffix (e.g. ON CONFLICT (name) DO NOTHING)." language="sql" value={value.suffix} onChange={(v) => set("suffix", v)} />
      <NumberField label="Max In Flight" min={1} value={value.max_in_flight} onChange={(v) => set("max_in_flight", v)} />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Batching</h4>
        <BatchingEditor value={value.batching} onChange={(v) => set("batching", v)} errors={errors} errorPathPrefix="batching" />
      </section>
    </div>
  );
}
