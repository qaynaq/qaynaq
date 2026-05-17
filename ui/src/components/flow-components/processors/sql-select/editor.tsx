import {
  TextField,
  SelectField,
  ArrayField,
  CodeField,
} from "@/components/form-primitives";
import { SQL_DRIVERS } from "../sql-raw";
import type { EditorProps } from "../../types";

interface Config {
  driver: (typeof SQL_DRIVERS)[number];
  dsn: string;
  table: string;
  columns: string[];
  where: string;
  args_mapping: string;
  prefix: string;
  suffix: string;
  init_statement: string;
  conn_max_idle_time: string;
  conn_max_life_time: string;
}

export default function SqlSelectEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <SelectField
        label="Driver"
        required
        value={value.driver}
        onChange={(v) =>
          onChange({ ...value, driver: v as Config["driver"] })
        }
        options={SQL_DRIVERS as unknown as string[]}
      />
      <TextField
        label="DSN"
        required
        value={value.dsn}
        onChange={(v) => onChange({ ...value, dsn: v })}
        error={errors?.dsn}
      />
      <TextField
        label="Table"
        required
        value={value.table}
        onChange={(v) => onChange({ ...value, table: v })}
        error={errors?.table}
      />
      <ArrayField
        label="Columns"
        description="A list of columns to query."
        required
        value={value.columns}
        onChange={(v) => onChange({ ...value, columns: v })}
      />
      <CodeField
        label="Where"
        description="An optional where clause to add."
        language="sql"
        value={value.where}
        onChange={(v) => onChange({ ...value, where: v })}
      />
      <CodeField
        label="Args Mapping"
        description="A Bloblang mapping evaluating to an array of values matching where-clause placeholders."
        value={value.args_mapping}
        onChange={(v) => onChange({ ...value, args_mapping: v })}
      />
      <TextField
        label="Prefix"
        description="Optional prefix to prepend to the query (before SELECT)."
        value={value.prefix}
        onChange={(v) => onChange({ ...value, prefix: v })}
      />
      <TextField
        label="Suffix"
        description="Optional suffix to append to the select query."
        value={value.suffix}
        onChange={(v) => onChange({ ...value, suffix: v })}
      />
      <CodeField
        label="Init Statement"
        description="Optional SQL executed once on first connection."
        language="sql"
        value={value.init_statement}
        onChange={(v) => onChange({ ...value, init_statement: v })}
      />
      <TextField
        label="Conn Max Idle Time"
        value={value.conn_max_idle_time}
        onChange={(v) => onChange({ ...value, conn_max_idle_time: v })}
      />
      <TextField
        label="Conn Max Life Time"
        value={value.conn_max_life_time}
        onChange={(v) => onChange({ ...value, conn_max_life_time: v })}
      />
    </div>
  );
}
