package mcp

import "sort"

// Maintainer tags catalog entries with their trust tier so users can see at
// a glance whether an MCP server is vendor-maintained or community-maintained.
type Maintainer string

const (
	MaintainerOfficial  Maintainer = "official"
	MaintainerCommunity Maintainer = "community"
)

// StdioCatalogEntry describes one allowlisted command-line MCP server.
// v1 is hardcoded; free-form commands are deliberately not supported.
type StdioCatalogEntry struct {
	ID           string         `json:"id"`
	DisplayName  string         `json:"display_name"`
	Description  string         `json:"description"`
	DocsURL      string         `json:"docs_url,omitempty"`
	Maintainer   Maintainer     `json:"maintainer"`
	Command      string         `json:"command"`
	ArgsTemplate []string       `json:"args"`
	EnvSpec      []StdioEnvSpec `json:"env_spec"`
}

type StdioEnvSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
	Advanced bool `json:"advanced"`
}

var stdioCatalog = map[string]StdioCatalogEntry{
	"filesystem": {
		ID:           "filesystem",
		DisplayName:  "Filesystem",
		Description:  "Read and write files in a sandboxed directory.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem",
		Maintainer:   MaintainerOfficial,
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-filesystem", "${ALLOWED_DIR}"},
		EnvSpec: []StdioEnvSpec{
			{Name: "ALLOWED_DIR", Description: "Absolute path the server is allowed to read/write.", Required: true},
		},
	},
	"slack": {
		ID:           "slack",
		DisplayName:  "Slack",
		Description:  "Read and post Slack messages.",
		DocsURL:      "https://github.com/korotovsky/slack-mcp-server",
		Maintainer:   MaintainerCommunity,
		Command:      "npx",
		ArgsTemplate: []string{"-y", "slack-mcp-server@^1.1.0"},
		EnvSpec: []StdioEnvSpec{
			{Name: "SLACK_MCP_XOXP_TOKEN", Description: "User token (xoxp-...) - see slack-mcp-server docs.", Required: true, Secret: true},
		},
	},
	"playwright": {
		ID:           "playwright",
		DisplayName:  "Playwright",
		Description:  "Browser automation, headless by default.",
		DocsURL:      "https://github.com/microsoft/playwright-mcp",
		Maintainer:   MaintainerOfficial,
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@playwright/mcp@latest"},
	},
	"sentry": {
		ID:           "sentry",
		DisplayName:  "Sentry",
		Description:  "Sentry issues, events, and performance lookup.",
		DocsURL:      "https://github.com/getsentry/sentry-mcp",
		Maintainer:   MaintainerOfficial,
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@sentry/mcp-server"},
		EnvSpec: []StdioEnvSpec{
			{Name: "SENTRY_AUTH_TOKEN", Description: "Sentry auth token (or ${SECRET_KEY}).", Required: true, Secret: true},
			{Name: "SENTRY_HOST", Description: "Sentry host (defaults to sentry.io).", Advanced: true},
		},
	},
	"redash": {
		ID:           "redash",
		DisplayName:  "Redash",
		Description:  "Run SQL queries, browse schemas, manage dashboards.",
		DocsURL:      "https://github.com/seob717/redash-mcp",
		Maintainer:   MaintainerCommunity,
		Command:      "npx",
		ArgsTemplate: []string{"-y", "redash-mcp@^3.0.0"},
		EnvSpec: []StdioEnvSpec{
			{Name: "REDASH_URL", Description: "Redash base URL (https://redash.example.com).", Required: true},
			{Name: "REDASH_API_KEY", Description: "User API key from Redash settings.", Required: true, Secret: true},
			{Name: "REDASH_SAFETY_MODE", Description: "off | warn (default) | strict - strict recommended.", Advanced: true},
			{Name: "REDASH_AUTO_LIMIT", Description: "Auto-append LIMIT N to unbounded SELECTs (e.g. 1000).", Advanced: true},
			{Name: "REDASH_DEFAULT_MAX_AGE", Description: "Cache age in seconds for query results (default 0).", Advanced: true},
			{Name: "REDASH_MCP_CACHE_TTL", Description: "MCP-side cache TTL in seconds (default 300).", Advanced: true},
		},
	},
}

func LookupCatalogEntry(id string) (StdioCatalogEntry, bool) {
	e, ok := stdioCatalog[id]
	return e, ok
}

func ListCatalogEntries() []StdioCatalogEntry {
	ids := make([]string, 0, len(stdioCatalog))
	for id := range stdioCatalog {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]StdioCatalogEntry, 0, len(ids))
	for _, id := range ids {
		out = append(out, stdioCatalog[id])
	}
	return out
}
