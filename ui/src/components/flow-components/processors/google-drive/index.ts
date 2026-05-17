import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const DRIVE_ACTIONS = [
  "copy_file",
  "delete_file_permanent",
  "export_file",
  "upload_file",
  "create_folder",
  "move_file",
  "create_file_from_text",
  "remove_file_permission",
  "replace_file",
  "create_shared_drive",
  "add_file_sharing",
  "create_shortcut",
  "update_metadata",
  "rename",
  "delete_file",
  "list_files",
  "get_file",
  "get_permissions",
  "find_file",
  "find_folder",
  "find_or_create_file",
  "find_or_create_folder",
] as const;

export const DRIVE_ROLES = [
  "reader",
  "writer",
  "commenter",
  "owner",
  "organizer",
  "fileOrganizer",
] as const;

export const DRIVE_PERM_TYPES = ["user", "group", "domain", "anyone"] as const;

const configSchema = z.object({
  service_account_json: z.string(),
  oauth_connection: z.string(),
  delegate_to: z.string(),
  action: z.enum(DRIVE_ACTIONS),
  file_id: z.string(),
  file_name: z.string(),
  folder_id: z.string(),
  destination_folder_id: z.string(),
  mime_type: z.string(),
  content: z.string(),
  file_url: z.string(),
  description: z.string(),
  email: z.string(),
  role: z.enum(DRIVE_ROLES),
  permission_type: z.enum(DRIVE_PERM_TYPES),
  permission_id: z.string(),
  query: z.string(),
  max_results: z.string(),
  starred: z.boolean(),
  folder_color: z.string(),
  custom_properties: z.string(),
  shared_drive_name: z.string(),
  target_file_id: z.string(),
  send_notification: z.boolean(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  service_account_json: "",
  oauth_connection: "",
  delegate_to: "",
  action: "list_files",
  file_id: "",
  file_name: "",
  folder_id: "",
  destination_folder_id: "",
  mime_type: "",
  content: "",
  file_url: "",
  description: "",
  email: "",
  role: "reader",
  permission_type: "user",
  permission_id: "",
  query: "",
  max_results: "100",
  starred: false,
  folder_color: "",
  custom_properties: "",
  shared_drive_name: "",
  target_file_id: "",
  send_notification: true,
};

const component: FlowComponent<Config> = {
  id: "google_drive",
  name: "Google Drive",
  category: "processor",
  description: "Performs Google Drive operations.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
