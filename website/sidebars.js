/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docsSidebar: [
    "intro",
    {
      type: "category",
      label: "Getting Started",
      items: [
        "getting-started/installation",
        "getting-started/quickstart",
        "getting-started/configuration",
        "getting-started/authentication",
      ],
    },
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/architecture",
        "concepts/flows",
        "concepts/components",
        "concepts/files",
        "concepts/validation",
        "concepts/testing-flows",
      ],
    },
    {
      type: "category",
      label: "Components",
      items: [
        {
          type: "category",
          label: "Inputs",
          link: { type: "doc", id: "components/inputs/index" },
          items: [
            "components/inputs/generate",
            "components/inputs/http-client",
            "components/inputs/http-server",
            "components/inputs/kafka",
            "components/inputs/amqp-0-9",
            "components/inputs/broker",
            "components/inputs/cdc-mysql",
            "components/inputs/mcp-tool",
            "components/inputs/shopify",
          ],
        },
        {
          type: "category",
          label: "Processors",
          link: { type: "doc", id: "components/processors/index" },
          items: [
            "components/processors/ai-gateway",
            "components/processors/google-calendar",
            "components/processors/google-drive",
            "components/processors/google-sheets",
            "components/processors/shopify",
            "components/processors/mapping",
            "components/processors/command",
            "components/processors/python",
            "components/processors/json-schema",
            "components/processors/catch",
            "components/processors/switch",
            "components/processors/schema-registry-decode",
            "components/processors/http-client",
            "components/processors/sql-raw",
            "components/processors/sql-select",
            "components/processors/sql-insert",
          ],
        },
        {
          type: "category",
          label: "Outputs",
          link: { type: "doc", id: "components/outputs/index" },
          items: [
            "components/outputs/http-client",
            "components/outputs/kafka",
            "components/outputs/amqp-0-9",
            "components/outputs/sql-insert",
            "components/outputs/sync-response",
            "components/outputs/switch",
            "components/outputs/broker",
          ],
        },
        "components/caches",
      ],
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/google-oauth-setup",
        "guides/mcp-server",
        {
          type: "category",
          label: "MCP Tool Packs",
          link: { type: "doc", id: "guides/mcp-tool-packs/index" },
          items: [
            "guides/mcp-tool-packs/google-calendar",
            "guides/mcp-tool-packs/google-drive",
            "guides/mcp-tool-packs/google-sheets",
          ],
        },
        "guides/scaling-workers",
        "guides/keycloak-authentication",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/cli",
        "reference/environment-variables",
      ],
    },
  ],
};

module.exports = sidebars;
