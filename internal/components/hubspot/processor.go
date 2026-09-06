package hubspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	err := service.RegisterProcessor(
		"hubspot", ProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewProcessorFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

const baseURL = "https://api.hubapi.com"

type Processor struct {
	oauthConnection string
	action          string

	limit          *service.InterpolatedString
	after          *service.InterpolatedString
	objectID       *service.InterpolatedString
	properties     *service.InterpolatedString
	query          *service.InterpolatedString
	filters        *service.InterpolatedString
	propertiesJSON *service.InterpolatedString

	clientOnce sync.Once
	clientErr  error
	httpClient httpDoer

	mgr    *service.Resources
	logger *service.Logger
}

func NewProcessorFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	oauthConnection, err := conf.FieldString(hsfOAuthConnection)
	if err != nil {
		return nil, err
	}
	if oauthConnection == "" {
		return nil, fmt.Errorf("oauth_connection is required")
	}

	action, err := conf.FieldString(hsfAction)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		oauthConnection: oauthConnection,
		action:          action,
		mgr:             mgr,
		logger:          mgr.Logger(),
	}

	if p.limit, err = conf.FieldInterpolatedString(hsfLimit); err != nil {
		return nil, err
	}
	if conf.Contains(hsfAfter) {
		if p.after, err = conf.FieldInterpolatedString(hsfAfter); err != nil {
			return nil, err
		}
	}
	if conf.Contains(hsfObjectID) {
		if p.objectID, err = conf.FieldInterpolatedString(hsfObjectID); err != nil {
			return nil, err
		}
	}
	if conf.Contains(hsfProperties) {
		if p.properties, err = conf.FieldInterpolatedString(hsfProperties); err != nil {
			return nil, err
		}
	}
	if conf.Contains(hsfQuery) {
		if p.query, err = conf.FieldInterpolatedString(hsfQuery); err != nil {
			return nil, err
		}
	}
	if conf.Contains(hsfFilters) {
		if p.filters, err = conf.FieldInterpolatedString(hsfFilters); err != nil {
			return nil, err
		}
	}
	if conf.Contains(hsfPropertiesJSON) {
		if p.propertiesJSON, err = conf.FieldInterpolatedString(hsfPropertiesJSON); err != nil {
			return nil, err
		}
	}

	return p, nil
}

type resolvedFields struct {
	limitRaw       string
	after          string
	objectID       string
	properties     string
	query          string
	filters        string
	propertiesJSON string
}

// limitOr parses the raw limit, falling back to 10 and clamping to [1, maxLimit].
func (f *resolvedFields) limitOr(maxLimit int) int {
	n, err := strconv.Atoi(f.limitRaw)
	if err != nil || n < 1 {
		return 10
	}
	if n > maxLimit {
		return maxLimit
	}
	return n
}

func (p *Processor) resolveFields(msg *service.Message) (*resolvedFields, error) {
	r := &resolvedFields{}

	var err error
	if r.limitRaw, err = p.limit.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate limit: %w", err)
	}
	if p.after != nil {
		if r.after, err = p.after.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate after: %w", err)
		}
	}
	if p.objectID != nil {
		if r.objectID, err = p.objectID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate object_id: %w", err)
		}
	}
	if p.properties != nil {
		if r.properties, err = p.properties.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate properties: %w", err)
		}
	}
	if p.query != nil {
		if r.query, err = p.query.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate query: %w", err)
		}
	}
	if p.filters != nil {
		if r.filters, err = p.filters.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate filters: %w", err)
		}
	}
	if p.propertiesJSON != nil {
		if r.propertiesJSON, err = p.propertiesJSON.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate properties_json: %w", err)
		}
	}

	return r, nil
}

// objectForAction maps each action to its HubSpot CRM object path segment.
var objectForAction = map[string]string{
	actionListContacts:    "contacts",
	actionGetContact:      "contacts",
	actionSearchContacts:  "contacts",
	actionCreateContact:   "contacts",
	actionUpdateContact:   "contacts",
	actionDeleteContact:   "contacts",
	actionListCompanies:   "companies",
	actionGetCompany:      "companies",
	actionSearchCompanies: "companies",
	actionCreateCompany:   "companies",
	actionUpdateCompany:   "companies",
	actionDeleteCompany:   "companies",
	actionListDeals:       "deals",
	actionGetDeal:         "deals",
	actionSearchDeals:     "deals",
	actionCreateDeal:      "deals",
	actionUpdateDeal:      "deals",
	actionDeleteDeal:      "deals",
	actionListTickets:     "tickets",
	actionGetTicket:       "tickets",
	actionSearchTickets:   "tickets",
	actionCreateTicket:    "tickets",
	actionUpdateTicket:    "tickets",
	actionDeleteTicket:    "tickets",
}

func isGetAction(action string) bool {
	switch action {
	case actionGetContact, actionGetCompany, actionGetDeal, actionGetTicket:
		return true
	default:
		return false
	}
}

func isSearchAction(action string) bool {
	switch action {
	case actionSearchContacts, actionSearchCompanies, actionSearchDeals, actionSearchTickets:
		return true
	default:
		return false
	}
}

func isCreateAction(action string) bool {
	switch action {
	case actionCreateContact, actionCreateCompany, actionCreateDeal, actionCreateTicket:
		return true
	default:
		return false
	}
}

func isUpdateAction(action string) bool {
	switch action {
	case actionUpdateContact, actionUpdateCompany, actionUpdateDeal, actionUpdateTicket:
		return true
	default:
		return false
	}
}

func isDeleteAction(action string) bool {
	switch action {
	case actionDeleteContact, actionDeleteCompany, actionDeleteDeal, actionDeleteTicket:
		return true
	default:
		return false
	}
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	fields, err := p.resolveFields(msg)
	if err != nil {
		return nil, classifyHubSpotError(err)
	}

	object, ok := objectForAction[p.action]
	if !ok {
		return nil, classifyHubSpotError(fmt.Errorf("unsupported action: %s", p.action))
	}

	var result map[string]any
	switch {
	case isGetAction(p.action):
		result, err = p.getObject(ctx, object, fields)
	case isSearchAction(p.action):
		result, err = p.searchObjects(ctx, object, fields)
	case isCreateAction(p.action):
		result, err = p.createObject(ctx, object, fields)
	case isUpdateAction(p.action):
		result, err = p.updateObject(ctx, object, fields)
	case isDeleteAction(p.action):
		result, err = p.deleteObject(ctx, object, fields)
	default:
		result, err = p.listObjects(ctx, object, fields)
	}
	if err != nil {
		return nil, classifyHubSpotError(err)
	}

	outMsg := msg.Copy()
	outMsg.SetStructured(result)
	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	return nil
}

func (p *Processor) listObjects(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(f.limitOr(100)))
	if f.after != "" {
		q.Set("after", f.after)
	}
	if f.properties != "" {
		q.Set("properties", f.properties)
	}
	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s?%s", baseURL, object, q.Encode())
	return p.doGet(ctx, endpoint)
}

func (p *Processor) searchObjects(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	body := map[string]any{
		"limit": f.limitOr(200),
	}
	if f.query != "" {
		body["query"] = f.query
	}
	if f.after != "" {
		body["after"] = f.after
	}
	if f.properties != "" {
		body["properties"] = splitCSV(f.properties)
	}
	if f.filters != "" {
		var groups []any
		if err := json.Unmarshal([]byte(f.filters), &groups); err != nil {
			return nil, fmt.Errorf("invalid filters JSON: %w", err)
		}
		body["filterGroups"] = groups
	}

	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s/search", baseURL, object)
	return p.doPost(ctx, endpoint, body)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (p *Processor) getObject(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	if f.objectID == "" {
		return nil, fmt.Errorf("object_id is required for get_%s action", singular(object))
	}
	q := url.Values{}
	if f.properties != "" {
		q.Set("properties", f.properties)
	}
	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s/%s", baseURL, object, url.PathEscape(f.objectID))
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	return p.doGet(ctx, endpoint)
}

func (p *Processor) createObject(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	props, err := parsePropertiesJSON(f.propertiesJSON, "create_"+singular(object))
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s", baseURL, object)
	return p.doPost(ctx, endpoint, map[string]any{"properties": props})
}

func (p *Processor) updateObject(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	if f.objectID == "" {
		return nil, fmt.Errorf("object_id is required for update_%s action", singular(object))
	}
	props, err := parsePropertiesJSON(f.propertiesJSON, "update_"+singular(object))
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s/%s", baseURL, object, url.PathEscape(f.objectID))
	return p.doPatch(ctx, endpoint, map[string]any{"properties": props})
}

func (p *Processor) deleteObject(ctx context.Context, object string, f *resolvedFields) (map[string]any, error) {
	if f.objectID == "" {
		return nil, fmt.Errorf("object_id is required for delete_%s action", singular(object))
	}
	endpoint := fmt.Sprintf("%s/crm/v3/objects/%s/%s", baseURL, object, url.PathEscape(f.objectID))
	if err := p.doDelete(ctx, endpoint); err != nil {
		return nil, err
	}
	return map[string]any{
		"deleted": true,
		"id":      f.objectID,
		"object":  object,
	}, nil
}

// parsePropertiesJSON decodes the properties_json field into a property map,
// requiring a non-empty JSON object for write actions.
func parsePropertiesJSON(raw, action string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("properties_json is required for %s action", action)
	}
	var props map[string]any
	if err := json.Unmarshal([]byte(raw), &props); err != nil {
		return nil, fmt.Errorf("invalid properties_json: %w", err)
	}
	if len(props) == 0 {
		return nil, fmt.Errorf("properties_json must contain at least one property for %s action", action)
	}
	return props, nil
}

func singular(object string) string {
	switch object {
	case "contacts":
		return "contact"
	case "companies":
		return "company"
	case "deals":
		return "deal"
	case "tickets":
		return "ticket"
	default:
		return object
	}
}
