package google_calendar

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/warpstreamlabs/bento/public/service"
	"google.golang.org/api/calendar/v3"
)

func init() {
	err := service.RegisterProcessor(
		"google_calendar", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	serviceAccountJSON string
	oauthConnection    string
	delegateTo         string
	action             string
	calendarID         *service.InterpolatedString
	eventID            *service.InterpolatedString
	destCalendarID     *service.InterpolatedString
	summary            *service.InterpolatedString
	description        *service.InterpolatedString
	location           *service.InterpolatedString
	startTime          *service.InterpolatedString
	endTime            *service.InterpolatedString
	timeZone           *service.InterpolatedString
	attendees          *service.InterpolatedString
	quickAddText       *service.InterpolatedString
	query              *service.InterpolatedString
	maxResults         int
	sendUpdates        *service.InterpolatedString
	recurrence         *service.InterpolatedString
	visibility         *service.InterpolatedString
	addConference      bool
	calendarSummary    *service.InterpolatedString

	calendarService *calendar.Service
	serviceOnce     sync.Once
	serviceInitErr  error
	logger          *service.Logger
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	serviceAccountJSON, err := conf.FieldString(gcfServiceAccountJSON)
	if err != nil {
		return nil, err
	}

	oauthConnection, err := conf.FieldString(gcfOAuthConnection)
	if err != nil {
		return nil, err
	}

	if serviceAccountJSON == "" && oauthConnection == "" {
		return nil, fmt.Errorf("either service_account_json or oauth_connection must be provided")
	}

	action, err := conf.FieldString(gcfAction)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		serviceAccountJSON: serviceAccountJSON,
		oauthConnection:    oauthConnection,
		action:             action,
		logger:             mgr.Logger(),
	}

	if conf.Contains(gcfDelegateTo) {
		if p.delegateTo, err = conf.FieldString(gcfDelegateTo); err != nil {
			return nil, err
		}
	}

	if p.calendarID, err = conf.FieldInterpolatedString(gcfCalendarID); err != nil {
		return nil, err
	}

	if conf.Contains(gcfEventID) {
		if p.eventID, err = conf.FieldInterpolatedString(gcfEventID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfDestinationCalendarID) {
		if p.destCalendarID, err = conf.FieldInterpolatedString(gcfDestinationCalendarID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfSummary) {
		if p.summary, err = conf.FieldInterpolatedString(gcfSummary); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfDescription) {
		if p.description, err = conf.FieldInterpolatedString(gcfDescription); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfLocation) {
		if p.location, err = conf.FieldInterpolatedString(gcfLocation); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfStartTime) {
		if p.startTime, err = conf.FieldInterpolatedString(gcfStartTime); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfEndTime) {
		if p.endTime, err = conf.FieldInterpolatedString(gcfEndTime); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfTimeZone) {
		if p.timeZone, err = conf.FieldInterpolatedString(gcfTimeZone); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfAttendees) {
		if p.attendees, err = conf.FieldInterpolatedString(gcfAttendees); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfQuickAddText) {
		if p.quickAddText, err = conf.FieldInterpolatedString(gcfQuickAddText); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gcfQuery) {
		if p.query, err = conf.FieldInterpolatedString(gcfQuery); err != nil {
			return nil, err
		}
	}

	if p.maxResults, err = conf.FieldInt(gcfMaxResults); err != nil {
		return nil, err
	}

	if p.sendUpdates, err = conf.FieldInterpolatedString(gcfSendUpdates); err != nil {
		return nil, err
	}

	if conf.Contains(gcfRecurrence) {
		if p.recurrence, err = conf.FieldInterpolatedString(gcfRecurrence); err != nil {
			return nil, err
		}
	}

	if p.visibility, err = conf.FieldInterpolatedString(gcfVisibility); err != nil {
		return nil, err
	}

	if p.addConference, err = conf.FieldBool(gcfAddConference); err != nil {
		return nil, err
	}

	if conf.Contains(gcfCalendarSummary) {
		if p.calendarSummary, err = conf.FieldInterpolatedString(gcfCalendarSummary); err != nil {
			return nil, err
		}
	}

	return p, nil
}

type resolvedFields struct {
	calendarID      string
	eventID         string
	destCalendar    string
	summary         string
	description     string
	location        string
	startTime       string
	endTime         string
	timeZone        string
	attendees       []string
	quickAddText    string
	query           string
	sendUpdates     string
	recurrence      []string
	visibility      string
	calendarSummary string
}

func (p *Processor) resolveFields(msg *service.Message) (*resolvedFields, error) {
	r := &resolvedFields{}
	var err error

	if r.calendarID, err = p.calendarID.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate calendar_id: %w", err)
	}

	if p.eventID != nil {
		if r.eventID, err = p.eventID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate event_id: %w", err)
		}
	}

	if p.destCalendarID != nil {
		if r.destCalendar, err = p.destCalendarID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate destination_calendar_id: %w", err)
		}
	}

	if p.summary != nil {
		if r.summary, err = p.summary.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate summary: %w", err)
		}
	}

	if p.description != nil {
		if r.description, err = p.description.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate description: %w", err)
		}
	}

	if p.location != nil {
		if r.location, err = p.location.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate location: %w", err)
		}
	}

	if p.startTime != nil {
		if r.startTime, err = p.startTime.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate start_time: %w", err)
		}
	}

	if p.endTime != nil {
		if r.endTime, err = p.endTime.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate end_time: %w", err)
		}
	}

	if p.timeZone != nil {
		if r.timeZone, err = p.timeZone.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate time_zone: %w", err)
		}
	}

	if p.attendees != nil {
		raw, err := p.attendees.TryString(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate attendees: %w", err)
		}
		r.attendees = splitCSV(raw)
	}

	if p.quickAddText != nil {
		if r.quickAddText, err = p.quickAddText.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate quick_add_text: %w", err)
		}
	}

	if p.query != nil {
		if r.query, err = p.query.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate query: %w", err)
		}
	}

	if r.sendUpdates, err = p.sendUpdates.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate send_updates: %w", err)
	}

	if p.recurrence != nil {
		raw, err := p.recurrence.TryString(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate recurrence: %w", err)
		}
		r.recurrence = splitCSV(raw)
	}

	if r.visibility, err = p.visibility.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate visibility: %w", err)
	}

	if p.calendarSummary != nil {
		if r.calendarSummary, err = p.calendarSummary.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate calendar_summary: %w", err)
		}
	}

	return r, nil
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	fields, err := p.resolveFields(msg)
	if err != nil {
		return nil, classifyError(err)
	}

	var result map[string]any

	switch p.action {
	case actionAddAttendees:
		result, err = p.addAttendees(ctx, fields)
	case actionCreateCalendar:
		result, err = p.createCalendar(ctx, fields)
	case actionCreateEvent:
		result, err = p.createEvent(ctx, fields)
	case actionDeleteEvent:
		result, err = p.deleteEvent(ctx, fields)
	case actionFindBusyPeriods:
		result, err = p.findBusyPeriods(ctx, fields)
	case actionFindCalendars:
		result, err = p.findCalendars(ctx, fields)
	case actionFindEvents:
		result, err = p.findEvents(ctx, fields)
	case actionFindOrCreateEvent:
		result, err = p.findOrCreateEvent(ctx, fields)
	case actionGetCalendar:
		result, err = p.getCalendar(ctx, fields)
	case actionGetEvent:
		result, err = p.getEvent(ctx, fields)
	case actionMoveEvent:
		result, err = p.moveEvent(ctx, fields)
	case actionQuickAddEvent:
		result, err = p.quickAddEvent(ctx, fields)
	case actionUpdateEvent:
		result, err = p.updateEvent(ctx, fields)
	default:
		err = fmt.Errorf("unsupported action: %s", p.action)
	}

	if err != nil {
		return nil, classifyError(err)
	}

	outMsg := msg.Copy()
	outMsg.SetStructured(result)
	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	return nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}
