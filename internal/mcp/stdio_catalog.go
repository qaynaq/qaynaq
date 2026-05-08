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
	Advanced    bool   `json:"advanced"`
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
			{Name: "SLACK_MCP_XOXP_TOKEN", Description: "User OAuth token (xoxp-...). Pick one of: this, a bot token, or the stealth XOXC+XOXD pair.", Secret: true},
			{Name: "SLACK_MCP_XOXB_TOKEN", Description: "Bot OAuth token (xoxb-...). Alternative to user token.", Secret: true},
			{Name: "SLACK_MCP_ADD_MESSAGE_TOOL", Description: "Enable posting: 'true' for all channels, channel-id list, or '!'-prefixed exclusions. Off by default.", Secret: false},
			{Name: "SLACK_MCP_ENABLED_TOOLS", Description: "Comma-separated tool names to register (otherwise all).", Secret: false},
			{Name: "SLACK_MCP_XOXC_TOKEN", Description: "Browser session token for stealth mode (paired with XOXD).", Advanced: true, Secret: true},
			{Name: "SLACK_MCP_XOXD_TOKEN", Description: "Browser cookie 'd' for stealth mode (paired with XOXC).", Advanced: true, Secret: true},
			{Name: "SLACK_MCP_ADD_MESSAGE_MARK", Description: "Auto-mark posted messages as read.", Advanced: true},
			{Name: "SLACK_MCP_ADD_MESSAGE_UNFURLING", Description: "Enable link unfurling (or domain whitelist).", Advanced: true},
			{Name: "SLACK_MCP_USERS_CACHE", Description: "Path to users cache file.", Advanced: true},
			{Name: "SLACK_MCP_CHANNELS_CACHE", Description: "Path to channels cache file.", Advanced: true},
			{Name: "SLACK_MCP_LOG_LEVEL", Description: "debug | info (default) | warn | error.", Advanced: true},
			{Name: "SLACK_MCP_PROXY", Description: "Outbound HTTP/HTTPS proxy URL.", Advanced: true},
			{Name: "SLACK_MCP_USER_AGENT", Description: "Custom user-agent (Enterprise stealth).", Advanced: true},
			{Name: "SLACK_MCP_GOVSLACK", Description: "Enable GovSlack mode.", Advanced: true},
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
		EnvSpec: []StdioEnvSpec{
			{Name: "PLAYWRIGHT_MCP_BROWSER", Description: "chrome | firefox | webkit | msedge (default chromium)."},
			{Name: "PLAYWRIGHT_MCP_HEADLESS", Description: "true | false (default headed)."},
			{Name: "PLAYWRIGHT_MCP_VIEWPORT_SIZE", Description: "Viewport WxH, e.g. 1280x720."},
			{Name: "PLAYWRIGHT_MCP_USER_DATA_DIR", Description: "Browser profile directory (default temp)."},
			{Name: "PLAYWRIGHT_MCP_ISOLATED", Description: "Use an in-memory profile."},
			{Name: "PLAYWRIGHT_MCP_DEVICE", Description: "Device emulation, e.g. 'iPhone 15'."},
			{Name: "PLAYWRIGHT_MCP_CAPS", Description: "Extra capabilities: vision, pdf, devtools."},
			{Name: "PLAYWRIGHT_MCP_EXECUTABLE_PATH", Description: "Custom browser binary path.", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_ALLOWED_HOSTS", Description: "Server-side allow-list of hosts.", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_BLOCKED_ORIGINS", Description: "Block-list of origins.", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_PROXY_SERVER", Description: "Outbound proxy URL.", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_IGNORE_HTTPS_ERRORS", Description: "Skip TLS validation.", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_TIMEOUT_NAVIGATION", Description: "Navigation timeout in ms (default 60000).", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_TIMEOUT_ACTION", Description: "Action timeout in ms (default 5000).", Advanced: true},
			{Name: "PLAYWRIGHT_MCP_USER_AGENT", Description: "Custom User-Agent header.", Advanced: true},
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
