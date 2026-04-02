package google_calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

func (p *Processor) addAttendees(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.eventID == "" {
		return nil, fmt.Errorf("event_id is required for add_attendees action")
	}
	if len(f.attendees) == 0 {
		return nil, fmt.Errorf("attendees is required for add_attendees action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	existing, err := svc.Events.Get(f.calendarID, f.eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event for adding attendees: %w", err)
	}

	attendees := existing.Attendees
	for _, email := range f.attendees {
		attendees = append(attendees, &calendar.EventAttendee{Email: email})
	}

	patched, err := svc.Events.Patch(f.calendarID, f.eventID, &calendar.Event{
		Attendees: attendees,
	}).SendUpdates(f.sendUpdates).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add attendees: %w", err)
	}

	return map[string]any{"event": eventToMap(patched)}, nil
}

func (p *Processor) createCalendar(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.calendarSummary == "" {
		return nil, fmt.Errorf("calendar_summary is required for create_calendar action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	cal, err := svc.Calendars.Insert(&calendar.Calendar{
		Summary: f.calendarSummary,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar: %w", err)
	}

	return map[string]any{"calendar": calendarToMap(cal)}, nil
}

func (p *Processor) createEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.summary == "" {
		return nil, fmt.Errorf("summary is required for create_event action")
	}
	if f.startTime == "" {
		return nil, fmt.Errorf("start_time is required for create_event action")
	}
	if f.endTime == "" {
		return nil, fmt.Errorf("end_time is required for create_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	event := p.buildEvent(f)

	call := svc.Events.Insert(f.calendarID, event).SendUpdates(f.sendUpdates)
	if p.addConference {
		call = call.ConferenceDataVersion(1)
	}
	created, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return map[string]any{"event": eventToMap(created)}, nil
}

func (p *Processor) deleteEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.eventID == "" {
		return nil, fmt.Errorf("event_id is required for delete_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	err = svc.Events.Delete(f.calendarID, f.eventID).
		SendUpdates(f.sendUpdates).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete event: %w", err)
	}

	return map[string]any{
		"deleted":  true,
		"event_id": f.eventID,
	}, nil
}

func (p *Processor) findBusyPeriods(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.startTime == "" {
		return nil, fmt.Errorf("start_time is required for find_busy_periods action")
	}
	if f.endTime == "" {
		return nil, fmt.Errorf("end_time is required for find_busy_periods action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Freebusy.Query(&calendar.FreeBusyRequest{
		TimeMin: f.startTime,
		TimeMax: f.endTime,
		Items: []*calendar.FreeBusyRequestItem{
			{Id: f.calendarID},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query freebusy: %w", err)
	}

	var periods []map[string]any
	if cal, ok := resp.Calendars[f.calendarID]; ok {
		for _, busy := range cal.Busy {
			periods = append(periods, map[string]any{
				"start": busy.Start,
				"end":   busy.End,
			})
		}
	}

	if periods == nil {
		periods = []map[string]any{}
	}

	return map[string]any{
		"busy_periods": periods,
		"count":        len(periods),
	}, nil
}

func (p *Processor) findCalendars(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	maxResults := int64(p.maxResults)
	if maxResults > 250 {
		maxResults = 250
	}

	resp, err := svc.CalendarList.List().
		MaxResults(maxResults).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]map[string]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		calendars = append(calendars, map[string]any{
			"id":           item.Id,
			"summary":      item.Summary,
			"description":  item.Description,
			"time_zone":    item.TimeZone,
			"access_role":  item.AccessRole,
			"primary":      item.Primary,
			"hidden":       item.Hidden,
			"selected":     item.Selected,
			"color_id":     item.ColorId,
			"background_color": item.BackgroundColor,
			"foreground_color": item.ForegroundColor,
		})
	}

	return map[string]any{
		"calendars": calendars,
		"count":     len(calendars),
	}, nil
}

func (p *Processor) findEvents(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	call := svc.Events.List(f.calendarID).
		MaxResults(int64(p.maxResults)).
		SingleEvents(true).
		OrderBy("startTime")

	if f.query != "" {
		call = call.Q(f.query)
	}
	if f.startTime != "" {
		call = call.TimeMin(f.startTime)
	} else {
		call = call.TimeMin(time.Now().Format(time.RFC3339))
	}
	if f.endTime != "" {
		call = call.TimeMax(f.endTime)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]map[string]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		events = append(events, eventToMap(item))
	}

	return map[string]any{
		"events": events,
		"count":  len(events),
	}, nil
}

func (p *Processor) findOrCreateEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.summary == "" {
		return nil, fmt.Errorf("summary is required for find_or_create_event action")
	}
	if f.startTime == "" {
		return nil, fmt.Errorf("start_time is required for find_or_create_event action")
	}
	if f.endTime == "" {
		return nil, fmt.Errorf("end_time is required for find_or_create_event action")
	}

	searchResult, err := p.findEvents(ctx, &resolvedFields{
		calendarID: f.calendarID,
		query:      f.summary,
		startTime:  f.startTime,
		endTime:    f.endTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search for existing events: %w", err)
	}

	events := searchResult["events"].([]map[string]any)
	if len(events) > 0 {
		return map[string]any{
			"event":   events[0],
			"created": false,
		}, nil
	}

	result, err := p.createEvent(ctx, f)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"event":   result["event"],
		"created": true,
	}, nil
}

func (p *Processor) getCalendar(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	cal, err := svc.Calendars.Get(f.calendarID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	return map[string]any{"calendar": calendarToMap(cal)}, nil
}

func (p *Processor) getEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.eventID == "" {
		return nil, fmt.Errorf("event_id is required for get_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	event, err := svc.Events.Get(f.calendarID, f.eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return map[string]any{"event": eventToMap(event)}, nil
}

func (p *Processor) moveEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.eventID == "" {
		return nil, fmt.Errorf("event_id is required for move_event action")
	}
	if f.destCalendar == "" {
		return nil, fmt.Errorf("destination_calendar_id is required for move_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	moved, err := svc.Events.Move(f.calendarID, f.eventID, f.destCalendar).
		SendUpdates(f.sendUpdates).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to move event: %w", err)
	}

	return map[string]any{"event": eventToMap(moved)}, nil
}

func (p *Processor) quickAddEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.quickAddText == "" {
		return nil, fmt.Errorf("quick_add_text is required for quick_add_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	event, err := svc.Events.QuickAdd(f.calendarID, f.quickAddText).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to quick add event: %w", err)
	}

	return map[string]any{"event": eventToMap(event)}, nil
}

func (p *Processor) updateEvent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.eventID == "" {
		return nil, fmt.Errorf("event_id is required for update_event action")
	}

	svc, err := p.initCalendarService()
	if err != nil {
		return nil, err
	}

	patch := &calendar.Event{}
	if f.summary != "" {
		patch.Summary = f.summary
	}
	if f.description != "" {
		patch.Description = f.description
	}
	if f.location != "" {
		patch.Location = f.location
	}
	if f.startTime != "" {
		patch.Start = &calendar.EventDateTime{
			DateTime: f.startTime,
			TimeZone: f.timeZone,
		}
	}
	if f.endTime != "" {
		patch.End = &calendar.EventDateTime{
			DateTime: f.endTime,
			TimeZone: f.timeZone,
		}
	}
	if len(f.attendees) > 0 {
		for _, email := range f.attendees {
			patch.Attendees = append(patch.Attendees, &calendar.EventAttendee{Email: email})
		}
	}
	if len(f.recurrence) > 0 {
		patch.Recurrence = f.recurrence
	}
	if f.visibility != "default" {
		patch.Visibility = f.visibility
	}
	if p.addConference {
		patch.ConferenceData = &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: fmt.Sprintf("qaynaq-%d", time.Now().UnixNano()),
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		}
	}

	call := svc.Events.Patch(f.calendarID, f.eventID, patch).SendUpdates(f.sendUpdates)
	if p.addConference {
		call = call.ConferenceDataVersion(1)
	}
	updated, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return map[string]any{"event": eventToMap(updated)}, nil
}

func (p *Processor) buildEvent(f *resolvedFields) *calendar.Event {
	event := &calendar.Event{
		Summary: f.summary,
		Start: &calendar.EventDateTime{
			DateTime: f.startTime,
			TimeZone: f.timeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: f.endTime,
			TimeZone: f.timeZone,
		},
	}

	if f.description != "" {
		event.Description = f.description
	}
	if f.location != "" {
		event.Location = f.location
	}
	if len(f.attendees) > 0 {
		for _, email := range f.attendees {
			event.Attendees = append(event.Attendees, &calendar.EventAttendee{Email: email})
		}
	}
	if len(f.recurrence) > 0 {
		event.Recurrence = f.recurrence
	}
	if f.visibility != "default" {
		event.Visibility = f.visibility
	}
	if p.addConference {
		event.ConferenceData = &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: fmt.Sprintf("qaynaq-%d", time.Now().UnixNano()),
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		}
	}

	return event
}

func eventToMap(event *calendar.Event) map[string]any {
	data, _ := json.Marshal(event)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}

func calendarToMap(cal *calendar.Calendar) map[string]any {
	data, _ := json.Marshal(cal)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
