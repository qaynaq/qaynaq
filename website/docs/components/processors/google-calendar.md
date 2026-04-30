# Google Calendar

Performs Google Calendar operations - create, read, update, and delete events and calendars.

:::tip MCP Tool Pack Available
Want to expose Google Calendar actions as MCP tools for AI assistants? Use the [Google Calendar MCP Tool Pack](/docs/guides/mcp-tool-packs/google-calendar) to deploy all 13 tools in one step - no manual configuration needed.
:::

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Service Account JSON | secret | - | Google service account credentials JSON (required) |
| Delegate To | string | - | Email to impersonate via Domain-Wide Delegation |
| Action | select | - | The calendar operation to perform (required) |
| Calendar ID | string | `primary` | Target calendar ID (required) |
| Event ID | string | - | Event identifier for event-specific actions |
| Destination Calendar ID | string | - | Target calendar for move_event |
| Summary | string | - | Event title |
| Description | string | - | Event description text |
| Location | string | - | Event location |
| Start Time | string | - | Start time in RFC3339 format |
| End Time | string | - | End time in RFC3339 format |
| Time Zone | string | - | IANA time zone (e.g. `America/New_York`) |
| Attendees | string | - | Comma-separated attendee email addresses |
| Quick Add Text | string | - | Natural language text for quick_add_event |
| Query | string | - | Free text search for find_events |
| Max Results | integer | `25` | Maximum results to return |
| Send Updates | string | `none` | Notification policy: `all`, `externalOnly`, `none` |
| Recurrence | string | - | Comma-separated RRULE recurrence rules |
| Visibility | string | `default` | Event visibility: `default`, `public`, `private`, `confidential` |
| Add Google Meet | boolean | `false` | Auto-generate a Google Meet conference link |
| Calendar Name | string | - | Name for create_calendar |

## Authentication

This processor supports two authentication methods: **OAuth Connection** (recommended for personal accounts) and **Service Account** (for server-to-server automation).

### Option A: OAuth Connection (recommended)

OAuth lets you authorize Qaynaq to act as your Google account. Events are owned by you, using your quota. You authorize once and it works indefinitely.

Follow the [Google OAuth Setup](/docs/guides/google-oauth-setup) guide to create a connection. Make sure to enable the **Google Calendar API** and add the `auth/calendar` scope. Then select your connection from the **OAuth Connection** dropdown in the processor configuration.

### Option B: Service Account

Service accounts are best for Google Workspace organizations with Domain-Wide Delegation.

#### Step 1: Create a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select an existing one
3. Navigate to **IAM & Admin** > **Service Accounts**
4. Click **Create Service Account**, give it a name, and click **Done**
5. Click the service account, go to the **Keys** tab
6. Click **Add Key** > **Create new key** > **JSON**
7. Download the JSON key file

#### Step 2: Enable the Calendar API

1. In Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for **Google Calendar API** and click **Enable**

#### Step 3: Store the Credentials

1. Open the downloaded JSON key file and copy its entire contents
2. In Qaynaq, go to **Secrets**
3. Create a new secret (e.g. key: `GOOGLE_CALENDAR_SA`) and paste the JSON as the value
4. In the Google Calendar processor, select `GOOGLE_CALENDAR_SA` from the Service Account JSON dropdown

#### Domain-Wide Delegation (Google Workspace only)

To access other users' calendars in a Google Workspace domain:

1. In Google Cloud Console, go to the service account details
2. Enable **Domain-Wide Delegation** and note the Client ID
3. In Google Workspace Admin Console, go to **Security** > **API Controls** > **Domain-wide Delegation**
4. Add the Client ID with the scope `https://www.googleapis.com/auth/calendar`
5. Set the **Delegate To** field to the email of the user whose calendar you want to access

:::tip
Without Domain-Wide Delegation, the service account can only access calendars that have been explicitly shared with the service account's email address (found in the `client_email` field of the JSON key).
:::

## Actions

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><line x1="12" y1="14" x2="12" y2="18"/><line x1="10" y1="16" x2="14" y2="16"/></svg>
<div class="action-card-content">
<h4>Create Detailed Event</h4>
<p>Create an event by defining each field.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><path d="M9 16l2 2 4-4"/></svg>
<div class="action-card-content">
<h4>Quick Add Event</h4>
<p>Create an event from a piece of text. Google parses the text for date, time, and description info.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><circle cx="12" cy="15" r="1"/></svg>
<div class="action-card-content">
<h4>Retrieve Event by ID</h4>
<p>Finds a specific event by its ID in your calendar.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><path d="M10 14l4 4m0-4l-4 4"/></svg>
<div class="action-card-content">
<h4>Delete Event</h4>
<p>Deletes an event.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><path d="M9 14h6"/></svg>
<div class="action-card-content">
<h4>Update Event</h4>
<p>Updates an event. Only filled fields are updated.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="19" y1="8" x2="19" y2="14"/><line x1="22" y1="11" x2="16" y2="11"/></svg>
<div class="action-card-content">
<h4>Add Attendee(s) to Event</h4>
<p>Invites one or more person to an existing event.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
<div class="action-card-content">
<h4>Find Events</h4>
<p>Finds events in your calendar. Returns up to 25 matching events.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
<div class="action-card-content">
<h4>Find or Create Events</h4>
<p>Finds or creates events.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><path d="M8 14h.01M12 14h.01M16 14h.01M8 18h.01M12 18h.01"/></svg>
<div class="action-card-content">
<h4>Find Busy Periods in Calendar</h4>
<p>Finds busy time periods in your calendar for a specific timeframe.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
<div class="action-card-content">
<h4>Move Event to Another Calendar</h4>
<p>Move an event from one calendar to another calendar.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
<div class="action-card-content">
<h4>Create Calendar</h4>
<p>Creates a new calendar.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
<div class="action-card-content">
<h4>Get Calendar Information</h4>
<p>Retrieve calendar properties including timezone, access permissions, default settings, and metadata.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
<div class="action-card-content">
<h4>Find Calendars</h4>
<p>Get comprehensive list of calendars accessible to the user with their properties and access levels. Returns up to 250 matching calendars.</p>
</div>
</div>
</div>

:::info
**Create Calendar** and **Move Event to Another Calendar** are only available with Domain-Wide Delegation on Google Workspace enterprise accounts.
:::

## Dynamic Fields

All action parameter fields support Bento interpolation functions, allowing dynamic values from message content using `${!this.field_name}` syntax. This enables processing batches of events or reacting to upstream data. For example, setting Event ID to `${!this.event_id}` will read the event ID from the incoming message.

Comma-separated fields (Attendees, Recurrence) also support interpolation. For example, set Attendees to `${!this.email}` for a single attendee from the message, or `${!this.attendees.join(",")}` for an array field.

Static fields (not interpolated): Service Account JSON, Delegate To, Action, Max Results, Add Google Meet.

## Output Format

All actions return a structured JSON object:

- **Event actions** return an `event` object containing the full Google Calendar event data (id, summary, start, end, attendees, etc.)
- **List actions** (find_events, find_calendars, find_busy_periods) return an array and a `count` field
- **find_or_create_event** includes a `created` boolean field
- **delete_event** returns `{deleted: true, event_id: "..."}`

:::tip
Use a Mapping processor after Google Calendar to extract specific fields from the response, such as the event ID or the list of attendee responses.
:::
