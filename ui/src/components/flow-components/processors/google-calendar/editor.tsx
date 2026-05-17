import {
  TextField,
  TextAreaField,
  NumberField,
  CheckboxField,
  SelectField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import { CALENDAR_ACTIONS } from ".";
import type { EditorProps } from "../../types";

interface Config {
  service_account_json: string;
  oauth_connection: string;
  delegate_to: string;
  action: (typeof CALENDAR_ACTIONS)[number];
  calendar_id: string;
  event_id: string;
  destination_calendar_id: string;
  summary: string;
  description: string;
  location: string;
  start_time: string;
  end_time: string;
  time_zone: string;
  attendees: string;
  quick_add_text: string;
  query: string;
  max_results: number;
  send_updates: string;
  recurrence: string;
  visibility: string;
  add_conference: boolean;
  calendar_summary: string;
}

export default function GoogleCalendarEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ConnectionPickerField label="Service Account JSON" source="secrets" value={value.service_account_json} onChange={(v) => set("service_account_json", v)} />
      <ConnectionPickerField label="OAuth Connection" source="connections" value={value.oauth_connection} onChange={(v) => set("oauth_connection", v)} />
      <TextField label="Delegate To" description="Email to impersonate via Domain-Wide Delegation." value={value.delegate_to} onChange={(v) => set("delegate_to", v)} />
      <SelectField label="Action" required value={value.action} onChange={(v) => set("action", v as Config["action"])} options={CALENDAR_ACTIONS as unknown as string[]} />
      <TextField label="Calendar ID" description="Use 'primary' for the authenticated user's main calendar." required value={value.calendar_id} onChange={(v) => set("calendar_id", v)} error={errors?.calendar_id} />
      <TextField label="Event ID" value={value.event_id} onChange={(v) => set("event_id", v)} />
      <TextField label="Destination Calendar ID" value={value.destination_calendar_id} onChange={(v) => set("destination_calendar_id", v)} />
      <TextField label="Summary" value={value.summary} onChange={(v) => set("summary", v)} />
      <TextAreaField label="Description" rows={2} value={value.description} onChange={(v) => set("description", v)} />
      <TextField label="Location" value={value.location} onChange={(v) => set("location", v)} />
      <TextField label="Start Time" description="RFC3339 (e.g. 2025-01-15T09:00:00-05:00)." value={value.start_time} onChange={(v) => set("start_time", v)} />
      <TextField label="End Time" description="RFC3339 (e.g. 2025-01-15T10:00:00-05:00)." value={value.end_time} onChange={(v) => set("end_time", v)} />
      <TextField label="Time Zone" description="IANA time zone (e.g. America/New_York)." value={value.time_zone} onChange={(v) => set("time_zone", v)} />
      <TextField label="Attendees" description="Comma-separated emails (supports interpolation)." value={value.attendees} onChange={(v) => set("attendees", v)} />
      <TextField label="Quick Add Text" value={value.quick_add_text} onChange={(v) => set("quick_add_text", v)} />
      <TextField label="Query" value={value.query} onChange={(v) => set("query", v)} />
      <NumberField label="Max Results" min={1} value={value.max_results} onChange={(v) => set("max_results", v)} />
      <TextField label="Send Updates" description="all | externalOnly | none." value={value.send_updates} onChange={(v) => set("send_updates", v)} />
      <TextField label="Recurrence" description="RRULE rules (e.g. RRULE:FREQ=WEEKLY;COUNT=5)." value={value.recurrence} onChange={(v) => set("recurrence", v)} />
      <TextField label="Visibility" description="default | public | private | confidential." value={value.visibility} onChange={(v) => set("visibility", v)} />
      <CheckboxField label="Add Google Meet" checked={value.add_conference} onChange={(c) => set("add_conference", c)} />
      <TextField label="Calendar Name" description="Name for a new calendar (create_calendar only)." value={value.calendar_summary} onChange={(v) => set("calendar_summary", v)} />
    </div>
  );
}
