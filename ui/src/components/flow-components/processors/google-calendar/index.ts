import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const CALENDAR_ACTIONS = [
  "add_attendees",
  "create_calendar",
  "create_event",
  "delete_event",
  "find_busy_periods",
  "find_calendars",
  "find_events",
  "find_or_create_event",
  "get_calendar",
  "get_event",
  "move_event",
  "quick_add_event",
  "update_event",
] as const;

const configSchema = z.object({
  service_account_json: z.string(),
  oauth_connection: z.string(),
  delegate_to: z.string(),
  action: z.enum(CALENDAR_ACTIONS),
  calendar_id: z.string().min(1, "Required"),
  event_id: z.string(),
  destination_calendar_id: z.string(),
  summary: z.string(),
  description: z.string(),
  location: z.string(),
  start_time: z.string(),
  end_time: z.string(),
  time_zone: z.string(),
  attendees: z.string(),
  quick_add_text: z.string(),
  query: z.string(),
  max_results: z.number().int().min(1),
  send_updates: z.string(),
  recurrence: z.string(),
  visibility: z.string(),
  add_conference: z.boolean(),
  calendar_summary: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  service_account_json: "",
  oauth_connection: "",
  delegate_to: "",
  action: "create_event",
  calendar_id: "primary",
  event_id: "",
  destination_calendar_id: "",
  summary: "",
  description: "",
  location: "",
  start_time: "",
  end_time: "",
  time_zone: "",
  attendees: "",
  quick_add_text: "",
  query: "",
  max_results: 25,
  send_updates: "none",
  recurrence: "",
  visibility: "default",
  add_conference: false,
  calendar_summary: "",
};

const component: FlowComponent<Config> = {
  id: "google_calendar",
  name: "Google Calendar",
  category: "processor",
  description: "Performs Google Calendar operations.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
