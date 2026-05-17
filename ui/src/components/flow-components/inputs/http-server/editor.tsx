import {
  TextField,
  ArrayField,
  KeyValueField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  path: string;
  allowed_verbs: string[];
  timeout: string;
  sync_response: { status: string; headers: Record<string, string> };
}

export default function HttpServerEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setSync = (next: Config["sync_response"]) =>
    onChange({ ...value, sync_response: next });

  return (
    <div className="space-y-4">
      <TextField
        label="Path"
        description="The endpoint path to listen for POST requests."
        value={value.path}
        onChange={(v) => onChange({ ...value, path: v })}
        error={errors?.path}
      />
      <ArrayField
        label="Allowed Verbs"
        description="An array of verbs that are allowed for the path endpoint."
        value={value.allowed_verbs}
        onChange={(v) => onChange({ ...value, allowed_verbs: v })}
        placeholder="POST"
      />
      <TextField
        label="Timeout"
        description="Timeout for requests; the connection is closed if exceeded."
        value={value.timeout}
        onChange={(v) => onChange({ ...value, timeout: v })}
        error={errors?.timeout}
      />
      <section className="space-y-3 border-t pt-3">
        <h4 className="text-sm font-medium">Synchronous Responses</h4>
        <TextField
          label="HTTP Status Code"
          size="sm"
          value={value.sync_response.status}
          onChange={(s) => setSync({ ...value.sync_response, status: s })}
          error={errors?.["sync_response.status"]}
        />
        <KeyValueField
          label="Headers"
          description="A map of headers to add to the response."
          size="sm"
          value={value.sync_response.headers}
          onChange={(h) => setSync({ ...value.sync_response, headers: h })}
        />
      </section>
    </div>
  );
}
