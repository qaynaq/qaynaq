package mcp

import "sort"

// StdioCatalogEntry describes one allowlisted command-line MCP server.
// v1 is hardcoded; free-form commands are deliberately not supported.
type StdioCatalogEntry struct {
	ID           string         `json:"id"`
	DisplayName  string         `json:"display_name"`
	Description  string         `json:"description"`
	DocsURL      string         `json:"docs_url,omitempty"`
	Command      string         `json:"command"`
	ArgsTemplate []string       `json:"args"`
	EnvSpec      []StdioEnvSpec `json:"env_spec"`
}

type StdioEnvSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
}

var stdioCatalog = map[string]StdioCatalogEntry{
	"filesystem": {
		ID:           "filesystem",
		DisplayName:  "Filesystem",
		Description:  "Read and write files in a sandboxed directory.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem",
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-filesystem@^0.6.0", "${ALLOWED_DIR}"},
		EnvSpec: []StdioEnvSpec{
			{Name: "ALLOWED_DIR", Description: "Absolute path the server is allowed to read/write.", Required: true},
		},
	},
	"git": {
		ID:           "git",
		DisplayName:  "Git",
		Description:  "Read and search git repositories.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/git",
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-git@^0.6.0", "--repository", "${REPO_PATH}"},
		EnvSpec: []StdioEnvSpec{
			{Name: "REPO_PATH", Description: "Absolute path to the git repository.", Required: true},
		},
	},
	"github": {
		ID:           "github",
		DisplayName:  "GitHub",
		Description:  "Interact with GitHub repos, issues, and PRs.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/github",
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-github@^0.6.0"},
		EnvSpec: []StdioEnvSpec{
			{Name: "GITHUB_PERSONAL_ACCESS_TOKEN", Description: "Personal access token (or ${SECRET_KEY}).", Required: true, Secret: true},
		},
	},
	"slack": {
		ID:           "slack",
		DisplayName:  "Slack",
		Description:  "Read and post Slack messages.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/slack",
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-slack@^0.6.0"},
		EnvSpec: []StdioEnvSpec{
			{Name: "SLACK_BOT_TOKEN", Description: "Bot user OAuth token (xoxb-...).", Required: true, Secret: true},
			{Name: "SLACK_TEAM_ID", Description: "Team ID (T...).", Required: true},
		},
	},
	"postgres": {
		ID:           "postgres",
		DisplayName:  "Postgres",
		Description:  "Read-only access to a Postgres database.",
		DocsURL:      "https://github.com/modelcontextprotocol/servers/tree/main/src/postgres",
		Command:      "npx",
		ArgsTemplate: []string{"-y", "@modelcontextprotocol/server-postgres@^0.6.0", "${POSTGRES_URL}"},
		EnvSpec: []StdioEnvSpec{
			{Name: "POSTGRES_URL", Description: "Connection URL (postgres://...).", Required: true, Secret: true},
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
