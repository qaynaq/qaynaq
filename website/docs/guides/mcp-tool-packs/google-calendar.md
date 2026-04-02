---
sidebar_position: 1
---

# Google Calendar

Deploys up to 13 MCP tools for managing Google Calendar events and calendars.

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| OAuth Connection | Yes | Google OAuth connection. Set up in the [Connections](/docs/guides/google-oauth-setup) page first. Make sure to enable the **Google Calendar API** and add the `auth/calendar` scope. |

## Included Tools

| Tool | Description |
|------|-------------|
| `google_calendar_create_event` | Create a new event with title, time, location, and attendees |
| `google_calendar_find_events` | Search for events by text query or time range |
| `google_calendar_get_event` | Get detailed information about a specific event |
| `google_calendar_update_event` | Update an existing event's details |
| `google_calendar_delete_event` | Delete an event |
| `google_calendar_quick_add_event` | Create an event from natural language text |
| `google_calendar_find_busy_periods` | Find busy and free time periods |
| `google_calendar_add_attendees` | Add attendees to an existing event |
| `google_calendar_move_event` | Move an event to a different calendar |
| `google_calendar_find_calendars` | List all available calendars |
| `google_calendar_get_calendar` | Get details about a specific calendar |
| `google_calendar_create_calendar` | Create a new calendar |
| `google_calendar_find_or_create_event` | Find an existing event or create one if not found |
