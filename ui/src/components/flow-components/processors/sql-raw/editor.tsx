import {
  TextField,
  SelectField,
  CodeField,
  CheckboxField,
  NumberField,
} from "@/components/form-primitives";
import { SQL_DRIVERS } from ".";
import type { EditorProps } from "../../types";

interface Config {
  driver: (typeof SQL_DRIVERS)[number];
  dsn: string;
  query: string;
  unsafe_dynamic_query: boolean;
  args_mapping: string;
  exec_only: boolean;
  init_statement: string;
  conn_max_idle_time: string;
  conn_max_life_time: string;
  conn_max_idle: number;
  conn_max_open: number;
}

export default function SqlRawEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <SelectField
        label="Driver"
        description="A database driver to use."
        required
        value={value.driver}
        onChange={(v) =>
          onChange({ ...value, driver: v as Config["driver"] })
        }
        options={SQL_DRIVERS as unknown as string[]}
      />
      <TextField
        label="DSN"
        description="A Data Source Name to identify the target database."
        required
        value={value.dsn}
        onChange={(v) => onChange({ ...value, dsn: v })}
        error={errors?.dsn}
      />
      <CodeField
        label="Query"
        description="The query to execute. Placeholder arguments are populated with the args_mapping field."
        required
        language="sql"
        value={value.query}
        onChange={(v) => onChange({ ...value, query: v })}
        error={errors?.query}
      />
      <CheckboxField
        label="Unsafe Dynamic Query"
        description="Enable interpolation in the query. WARNING: may be susceptible to SQL injection."
        checked={value.unsafe_dynamic_query}
        onChange={(c) => onChange({ ...value, unsafe_dynamic_query: c })}
      />
      <CodeField
        label="Args Mapping"
        description="A Bloblang mapping evaluating to an array of values matching the query placeholders."
        value={value.args_mapping}
        onChange={(v) => onChange({ ...value, args_mapping: v })}
      />
      <CheckboxField
        label="Exec Only"
        description="Discard the query result. Useful for INSERT, UPDATE, DELETE."
        checked={value.exec_only}
        onChange={(c) => onChange({ ...value, exec_only: c })}
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
      <NumberField
        label="Conn Max Idle"
        min={0}
        value={value.conn_max_idle}
        onChange={(v) => onChange({ ...value, conn_max_idle: v })}
      />
      <NumberField
        label="Conn Max Open"
        min={0}
        value={value.conn_max_open}
        onChange={(v) => onChange({ ...value, conn_max_open: v })}
      />
    </div>
  );
}
