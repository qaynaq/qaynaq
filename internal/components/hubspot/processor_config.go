package hubspot

import "github.com/warpstreamlabs/bento/public/service"

const (
	hsfOAuthConnection = "oauth_connection"
	hsfAction          = "action"
	hsfLimit           = "limit"
	hsfAfter           = "after"
	hsfObjectID        = "object_id"
	hsfProperties      = "properties"
	hsfQuery           = "query"
	hsfFilters         = "filters"
	hsfPropertiesJSON  = "properties_json"
)

const (
	actionListContacts    = "list_contacts"
	actionGetContact      = "get_contact"
	actionSearchContacts  = "search_contacts"
	actionCreateContact   = "create_contact"
	actionUpdateContact   = "update_contact"
	actionDeleteContact   = "delete_contact"
	actionListCompanies   = "list_companies"
	actionGetCompany      = "get_company"
	actionSearchCompanies = "search_companies"
	actionCreateCompany   = "create_company"
	actionUpdateCompany   = "update_company"
	actionDeleteCompany   = "delete_company"
	actionListDeals       = "list_deals"
	actionGetDeal         = "get_deal"
	actionSearchDeals     = "search_deals"
	actionCreateDeal      = "create_deal"
	actionUpdateDeal      = "update_deal"
	actionDeleteDeal      = "delete_deal"
	actionListTickets     = "list_tickets"
	actionGetTicket       = "get_ticket"
	actionSearchTickets   = "search_tickets"
	actionCreateTicket    = "create_ticket"
	actionUpdateTicket    = "update_ticket"
	actionDeleteTicket    = "delete_ticket"
)

func ProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Services").
		Summary("Performs on-demand HubSpot CRM API operations for MCP tools.").
		Description(`
This processor connects to the HubSpot CRM API using a Qaynaq OAuth connection and
performs on-demand read and write operations on contacts, companies, deals, and
tickets. It is designed for use as an MCP tool processor, where each incoming message
triggers a single API call.

Authentication uses an OAuth connection configured on the Connections page. The
connection's access token is injected automatically and refreshed when it expires
(HubSpot access tokens are short-lived), so no manual token handling is required.

The OAuth app needs scopes matching the actions you use. Read actions need the
.read scopes; create, update, and delete actions need the matching .write scopes:
- crm.objects.contacts.read / crm.objects.contacts.write
- crm.objects.companies.read / crm.objects.companies.write
- crm.objects.deals.read / crm.objects.deals.write
- crm.objects.tickets.read / crm.objects.tickets.write

Delete actions archive the record in HubSpot (recoverable from the HubSpot UI for a
limited window), they do not hard-delete.

Most action parameter fields support interpolation using the ` + "`${!this.field_name}`" + ` syntax,
allowing dynamic values from the incoming message.`).
		Field(service.NewStringField(hsfOAuthConnection).
			Description("HubSpot OAuth connection name. Set up on the Connections page first.")).
		Field(service.NewStringEnumField(hsfAction,
			actionListContacts,
			actionGetContact,
			actionSearchContacts,
			actionCreateContact,
			actionUpdateContact,
			actionDeleteContact,
			actionListCompanies,
			actionGetCompany,
			actionSearchCompanies,
			actionCreateCompany,
			actionUpdateCompany,
			actionDeleteCompany,
			actionListDeals,
			actionGetDeal,
			actionSearchDeals,
			actionCreateDeal,
			actionUpdateDeal,
			actionDeleteDeal,
			actionListTickets,
			actionGetTicket,
			actionSearchTickets,
			actionCreateTicket,
			actionUpdateTicket,
			actionDeleteTicket,
		).Description("The HubSpot CRM operation to perform.")).
		Field(service.NewInterpolatedStringField(hsfLimit).
			Description("Maximum records per page. List operations allow 1-100, search operations allow 1-200.").
			Default("10")).
		Field(service.NewInterpolatedStringField(hsfAfter).
			Description("Pagination cursor, taken from paging.next.after of a previous response. Used by list and search operations.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(hsfObjectID).
			Description("Record ID for get operations.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(hsfProperties).
			Description("Comma-separated list of properties to return. If empty, HubSpot returns its default property set.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(hsfQuery).
			Description("Free-text search string for search operations. Matches across the object's default searchable properties.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(hsfFilters).
			Description("Optional JSON array of filter groups for search operations, e.g. " +
				`[{"filters":[{"propertyName":"email","operator":"EQ","value":"x@y.com"}]}]` +
				". Combined with OR across groups, AND within a group. Operators: EQ, NEQ, LT, LTE, GT, GTE, BETWEEN, IN, NOT_IN, HAS_PROPERTY, NOT_HAS_PROPERTY, CONTAINS_TOKEN, NOT_CONTAINS_TOKEN.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(hsfPropertiesJSON).
			Description("JSON object of property values to write for create and update operations, e.g. " +
				`{"email":"x@y.com","firstname":"Sarah"}` +
				". Required for create and update actions.").
			Default("").
			Optional()).
		Version("0.3.0")
}
