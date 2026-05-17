import {
  TextField,
  TextAreaField,
  CheckboxField,
  SelectField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import { DRIVE_ACTIONS, DRIVE_ROLES, DRIVE_PERM_TYPES } from ".";
import type { EditorProps } from "../../types";

interface Config {
  service_account_json: string;
  oauth_connection: string;
  delegate_to: string;
  action: (typeof DRIVE_ACTIONS)[number];
  file_id: string;
  file_name: string;
  folder_id: string;
  destination_folder_id: string;
  mime_type: string;
  content: string;
  file_url: string;
  description: string;
  email: string;
  role: (typeof DRIVE_ROLES)[number];
  permission_type: (typeof DRIVE_PERM_TYPES)[number];
  permission_id: string;
  query: string;
  max_results: string;
  starred: boolean;
  folder_color: string;
  custom_properties: string;
  shared_drive_name: string;
  target_file_id: string;
  send_notification: boolean;
}

export default function GoogleDriveEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ConnectionPickerField label="Service Account JSON" source="secrets" value={value.service_account_json} onChange={(v) => set("service_account_json", v)} />
      <ConnectionPickerField label="OAuth Connection" source="connections" value={value.oauth_connection} onChange={(v) => set("oauth_connection", v)} />
      <TextField label="Delegate To" value={value.delegate_to} onChange={(v) => set("delegate_to", v)} />
      <SelectField label="Action" required value={value.action} onChange={(v) => set("action", v as Config["action"])} options={DRIVE_ACTIONS as unknown as string[]} />
      <TextField label="File ID" value={value.file_id} onChange={(v) => set("file_id", v)} />
      <TextField label="File Name" value={value.file_name} onChange={(v) => set("file_name", v)} />
      <TextField label="Folder ID" value={value.folder_id} onChange={(v) => set("folder_id", v)} />
      <TextField label="Destination Folder ID" value={value.destination_folder_id} onChange={(v) => set("destination_folder_id", v)} />
      <TextField label="MIME Type" value={value.mime_type} onChange={(v) => set("mime_type", v)} />
      <TextAreaField label="Content" rows={3} value={value.content} onChange={(v) => set("content", v)} />
      <TextField label="File URL" value={value.file_url} onChange={(v) => set("file_url", v)} />
      <TextField label="Description" value={value.description} onChange={(v) => set("description", v)} />
      <TextField label="Email" value={value.email} onChange={(v) => set("email", v)} />
      <SelectField label="Role" value={value.role} onChange={(v) => set("role", v as Config["role"])} options={DRIVE_ROLES as unknown as string[]} />
      <SelectField label="Permission Type" value={value.permission_type} onChange={(v) => set("permission_type", v as Config["permission_type"])} options={DRIVE_PERM_TYPES as unknown as string[]} />
      <TextField label="Permission ID" value={value.permission_id} onChange={(v) => set("permission_id", v)} />
      <TextField label="Query" description="Drive query syntax (list_files)." value={value.query} onChange={(v) => set("query", v)} />
      <TextField label="Max Results" value={value.max_results} onChange={(v) => set("max_results", v)} />
      <CheckboxField label="Starred" checked={value.starred} onChange={(c) => set("starred", c)} />
      <TextField label="Folder Color" description="Hex (e.g. #FF0000)." value={value.folder_color} onChange={(v) => set("folder_color", v)} />
      <TextField label="Custom Properties" description='JSON object (e.g. {"key":"value"}).' value={value.custom_properties} onChange={(v) => set("custom_properties", v)} />
      <TextField label="Shared Drive Name" value={value.shared_drive_name} onChange={(v) => set("shared_drive_name", v)} />
      <TextField label="Target File ID" value={value.target_file_id} onChange={(v) => set("target_file_id", v)} />
      <CheckboxField label="Send Notification" checked={value.send_notification} onChange={(c) => set("send_notification", c)} />
    </div>
  );
}
