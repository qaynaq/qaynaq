package mcp_call

import "github.com/warpstreamlabs/bento/public/service"

const (
	fieldServerURL   = "server_url"
	fieldTool        = "tool"
	fieldAuthHeader  = "auth_header"
	fieldAuthValue   = "auth_value"
	fieldArgsMapping = "args_mapping"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("AI", "MCP").
		Summary("Calls a tool on an external MCP server.").
		Description(`
This processor connects to an external MCP server and calls a specific tool,
mapping the incoming message to tool arguments via a Bloblang mapping.

The tool response replaces the message content. Metadata fields ` + "`mcp_server`" + `,
` + "`mcp_tool`" + `, and ` + "`mcp_latency_ms`" + ` are added to the output message.

The connection is established lazily on the first call and reused for subsequent calls.
If the connection fails, it will be retried on the next message (not permanently broken).`).
		Field(service.NewStringField(fieldServerURL).
			Description("URL of the external MCP server to connect to.")).
		Field(service.NewStringField(fieldTool).
			Description("Name of the tool to call on the MCP server.")).
		Field(service.NewStringField(fieldAuthHeader).
			Description("HTTP header name for authentication (e.g., 'Authorization', 'X-API-Key'). Leave empty for no auth.").
			Default("").
			Optional()).
		Field(service.NewStringField(fieldAuthValue).
			Description("Value for the authentication header (e.g., 'Bearer sk-...'). Leave empty for no auth.").
			Default("").
			Secret().
			Optional()).
		Field(service.NewBloblangField(fieldArgsMapping).
			Description("A Bloblang mapping that transforms the incoming message into a JSON object of tool arguments.").
			Optional()).
		Version("1.0.0")
}
