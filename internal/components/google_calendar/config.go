package google_calendar

import "github.com/warpstreamlabs/bento/public/service"

const (
	gcfServiceAccountJSON    = "service_account_json"
	gcfDelegateTo            = "delegate_to"
	gcfAction                = "action"
	gcfCalendarID            = "calendar_id"
	gcfEventID               = "event_id"
	gcfDestinationCalendarID = "destination_calendar_id"
	gcfSummary               = "summary"
	gcfDescription           = "description"
	gcfLocation              = "location"
	gcfStartTime             = "start_time"
	gcfEndTime               = "end_time"
	gcfTimeZone              = "time_zone"
	gcfAttendees             = "attendees"
	gcfQuickAddText          = "quick_add_text"
	gcfQuery                 = "query"
	gcfMaxResults            = "max_results"
	gcfSendUpdates           = "send_updates"
	gcfRecurrence            = "recurrence"
	gcfVisibility            = "visibility"
	gcfAddConference         = "add_conference"
	gcfCalendarSummary       = "calendar_summary"
	gcfOAuthConnection       = "oauth_connection"
)

const (
	actionAddAttendees     = "add_attendees"
	actionCreateCalendar   = "create_calendar"
	actionCreateEvent      = "create_event"
	actionDeleteEvent      = "delete_event"
	actionFindBusyPeriods  = "find_busy_periods"
	actionFindCalendars    = "find_calendars"
	actionFindEvents       = "find_events"
	actionFindOrCreateEvent = "find_or_create_event"
	actionGetCalendar      = "get_calendar"
	actionGetEvent         = "get_event"
	actionMoveEvent        = "move_event"
	actionQuickAddEvent    = "quick_add_event"
	actionUpdateEvent      = "update_event"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Integration").
		Summary("Performs Google Calendar operations - create, read, update, and delete events and calendars.").
		Description(`
This processor interacts with the Google Calendar API using service account authentication.
It supports creating, reading, updating, and deleting events and calendars, as well as
querying availability and searching for events.

Store your Google service account JSON as a secret in Settings > Secrets, then reference
it in the Service Account JSON field. For accessing other users' calendars in a Google
Workspace domain, enable Domain-Wide Delegation and set the Delegate To field.

Most fields support interpolation functions, allowing dynamic values from message content
using the ` + "`${!this.field_name}`" + ` syntax.`).
		Field(service.NewStringField(gcfServiceAccountJSON).
			Description("Google service account credentials JSON. Store as a secret and reference via ${SECRET_NAME}.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gcfOAuthConnection).
			Description("OAuth connection for user authentication. Set up in Settings > Connections, then reference via ${CONN_NAME}. Alternative to Service Account JSON.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gcfDelegateTo).
			Description("Email address to impersonate via Domain-Wide Delegation. Required for accessing other users' calendars in Google Workspace.").
			Default("").
			Optional()).
		Field(service.NewStringEnumField(gcfAction,
			actionAddAttendees,
			actionCreateCalendar,
			actionCreateEvent,
			actionDeleteEvent,
			actionFindBusyPeriods,
			actionFindCalendars,
			actionFindEvents,
			actionFindOrCreateEvent,
			actionGetCalendar,
			actionGetEvent,
			actionMoveEvent,
			actionQuickAddEvent,
			actionUpdateEvent,
		).Description("The calendar operation to perform.")).
		Field(service.NewInterpolatedStringField(gcfCalendarID).
			Description("Target calendar ID. Use 'primary' for the authenticated user's main calendar.").
			Default("primary")).
		Field(service.NewInterpolatedStringField(gcfEventID).
			Description("The event identifier. Required for: get_event, delete_event, update_event, add_attendees, move_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfDestinationCalendarID).
			Description("Target calendar to move the event to. Required for: move_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfSummary).
			Description("Event title. Required for: create_event, find_or_create_event. Optional for: update_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfDescription).
			Description("Event description text. Used by: create_event, update_event, find_or_create_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfLocation).
			Description("Event location. Used by: create_event, update_event, find_or_create_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfStartTime).
			Description("Start time in RFC3339 format (e.g. '2025-01-15T09:00:00-05:00'). Required for: create_event, find_or_create_event, find_busy_periods.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfEndTime).
			Description("End time in RFC3339 format (e.g. '2025-01-15T10:00:00-05:00'). Required for: create_event, find_or_create_event, find_busy_periods.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfTimeZone).
			Description("IANA time zone (e.g. 'America/New_York'). Defaults to UTC. Used by: create_event, update_event, find_or_create_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfAttendees).
			Description("Comma-separated attendee email addresses. Supports interpolation (e.g. '${!this.email}'). Used by: create_event, update_event, add_attendees, find_or_create_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfQuickAddText).
			Description("Natural language event description that Google will parse (e.g. 'Meeting with John tomorrow at 3pm'). Required for: quick_add_event.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gcfQuery).
			Description("Free text search terms to filter events. Used by: find_events, find_or_create_event.").
			Default("").
			Optional()).
		Field(service.NewIntField(gcfMaxResults).
			Description("Maximum number of results. Used by: find_events (max 2500), find_calendars (max 250).").
			Default(25)).
		Field(service.NewInterpolatedStringField(gcfSendUpdates).
			Description("Notification policy for attendees: all, externalOnly, none. Used by: create_event, update_event, delete_event, add_attendees, find_or_create_event.").
			Default("none").
			Advanced()).
		Field(service.NewInterpolatedStringField(gcfRecurrence).
			Description("Comma-separated RRULE recurrence rules (e.g. 'RRULE:FREQ=WEEKLY;COUNT=5'). Used by: create_event, update_event, find_or_create_event.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gcfVisibility).
			Description("Event visibility: default, public, private, confidential. Used by: create_event, update_event, find_or_create_event.").
			Default("default").
			Advanced()).
		Field(service.NewBoolField(gcfAddConference).
			Description("Auto-generate a Google Meet conference link. Used by: create_event, update_event, find_or_create_event.").
			Default(false).
			Advanced()).
		Field(service.NewInterpolatedStringField(gcfCalendarSummary).
			Description("Name for the new calendar. Required for: create_calendar.").
			Default("").
			Optional()).
		Version("1.0.0")
}
