package hubspot

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyHubSpotError(t *testing.T) {
	tests := []struct {
		name    string
		err     string
		wantPfx string
	}{
		{"required field", "object_id is required for get_contact action", "[400]"},
		{"interpolation error", "failed to interpolate limit: bad value", "[400]"},
		{"unsupported action", "unsupported action: foo", "[400]"},
		{"auth error", "401 Unauthorized", "[401]"},
		{"not found", "404 Not Found", "[404]"},
		{"rate limit", "429 rate limit exceeded", "[429]"},
		{"unknown error", "something went wrong", "[500]"},
		{"already classified passes through", "[409] conflict", "[409]"},
		{"already classified 404 not rewrapped", "[404] {\"message\":\"not found\"}", "[404]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyHubSpotError(errors.New(tt.err))
			if !strings.HasPrefix(err.Error(), tt.wantPfx) {
				t.Errorf("classifyHubSpotError(%q) = %q, want prefix %q", tt.err, err.Error(), tt.wantPfx)
			}
		})
	}
}

func TestClassifyHubSpotErrorNoDoubleWrap(t *testing.T) {
	err := classifyHubSpotError(errors.New("[500] internal"))
	if err.Error() != "[500] internal" {
		t.Errorf("classified error should pass through unchanged, got %q", err.Error())
	}
}

func TestObjectForActionCoversAllActions(t *testing.T) {
	actions := []string{
		actionListContacts, actionGetContact, actionSearchContacts, actionCreateContact, actionUpdateContact, actionDeleteContact,
		actionListCompanies, actionGetCompany, actionSearchCompanies, actionCreateCompany, actionUpdateCompany, actionDeleteCompany,
		actionListDeals, actionGetDeal, actionSearchDeals, actionCreateDeal, actionUpdateDeal, actionDeleteDeal,
		actionListTickets, actionGetTicket, actionSearchTickets, actionCreateTicket, actionUpdateTicket, actionDeleteTicket,
	}
	for _, a := range actions {
		if _, ok := objectForAction[a]; !ok {
			t.Errorf("action %q has no object mapping", a)
		}
	}
}

func TestActionClassification(t *testing.T) {
	type cls struct {
		get, search, create, update, del bool
	}
	want := map[string]cls{
		actionGetContact:    {get: true},
		actionSearchDeals:   {search: true},
		actionCreateTicket:  {create: true},
		actionUpdateCompany: {update: true},
		actionDeleteContact: {del: true},
		actionListCompanies: {}, // list = none of the above
	}
	for action, w := range want {
		got := cls{
			get:    isGetAction(action),
			search: isSearchAction(action),
			create: isCreateAction(action),
			update: isUpdateAction(action),
			del:    isDeleteAction(action),
		}
		if got != w {
			t.Errorf("%q classified as %+v, want %+v", action, got, w)
		}
	}
}

func TestEveryActionHasExactlyOneClassification(t *testing.T) {
	var all []string
	for a := range objectForAction {
		all = append(all, a)
	}
	for _, a := range all {
		n := 0
		for _, fn := range []func(string) bool{isGetAction, isSearchAction, isCreateAction, isUpdateAction, isDeleteAction} {
			if fn(a) {
				n++
			}
		}
		// list actions match none (handled by the default branch); all others match exactly one.
		if n > 1 {
			t.Errorf("%q matches %d classifiers, want at most 1", a, n)
		}
	}
}

func TestParsePropertiesJSON(t *testing.T) {
	props, err := parsePropertiesJSON(`{"email":"x@y.com","firstname":"Sarah"}`, "create_contact")
	if err != nil {
		t.Fatalf("valid properties: %v", err)
	}
	if props["email"] != "x@y.com" || props["firstname"] != "Sarah" {
		t.Errorf("parsed props = %v", props)
	}

	cases := []struct{ name, raw string }{
		{"empty", ""},
		{"whitespace", "   "},
		{"empty object", "{}"},
		{"invalid json", "{not json}"},
	}
	for _, c := range cases {
		if _, err := parsePropertiesJSON(c.raw, "update_deal"); err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
		}
	}
}

func TestLimitOr(t *testing.T) {
	tests := []struct {
		raw  string
		max  int
		want int
	}{
		{"25", 100, 25},
		{"25", 200, 25},
		{"0", 100, 10},    // below 1 -> default
		{"", 100, 10},     // unparseable -> default
		{"abc", 100, 10},  // unparseable -> default
		{"500", 100, 100}, // clamp to list max
		{"500", 200, 200}, // clamp to search max
		{"150", 100, 100}, // over list max
		{"150", 200, 150}, // within search max
	}
	for _, tt := range tests {
		f := &resolvedFields{limitRaw: tt.raw}
		if got := f.limitOr(tt.max); got != tt.want {
			t.Errorf("limitOr(%q, max=%d) = %d, want %d", tt.raw, tt.max, got, tt.want)
		}
	}
}
