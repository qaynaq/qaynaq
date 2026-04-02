// Component schemas with nested configuration support
export const componentSchemas = {
  input: {
    generate: {
      title: "Generate",
      properties: {
        mapping: {
          type: "code",
          title: "Mapping",
          description: "A bloblang mapping to use for generating messages.",
          required: true,
        },
        interval: {
          type: "input",
          title: "Interval",
          description:
            "The time interval at which messages should be generated. E.g. 1s, 1m, 1h, @every 1s, 0,30 */2 * * * *",
          default: "1s",
        },
        count: {
          type: "number",
          title: "Count",
          description:
            "An optional number of messages to generate, if set above 0 the specified number of messages is generated and then the input will shut down.",
          default: 0,
        },
        batch_size: {
          type: "number",
          title: "Batch size",
          description:
            "The number of generated messages that should be accumulated into each batch flushed at the specified interval.",
          default: 1,
        },
        auto_replay_nacks: {
          type: "bool",
          title: "Auto replay",
          description:
            "Whether messages that are rejected (nacked) at the output level should be automatically replayed indefinitely",
          default: true,
        },
      },
    },
    http_client: {
      title: "HTTP Client",
      properties: {
        url: {
          type: "input",
          title: "URL",
          description: "The URL to connect to.",
          required: true,
        },
        verb: {
          type: "select",
          title: "HTTP Verb",
          description: "The HTTP verb to use.",
          options: ["GET", "POST", "PUT", "DELETE"],
          default: "GET",
        },
        headers: {
          type: "key_value",
          title: "Headers",
          description: "A map of headers to add to the request.",
          default: {},
        },
        metadata: {
          type: "object",
          title: "Metadata",
          description: "Metadata configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include metadata prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include metadata patterns",
              default: [],
            },
          },
        },
        dump_request_log_level: {
          type: "input",
          title: "Dump Request Log Level",
          description: "Log level to dump request details",
          default: "",
        },
        oauth: {
          type: "object",
          title: "OAuth",
          description: "OAuth configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth authentication",
              default: false,
            },
            consumer_key: {
              type: "input",
              title: "Consumer Key",
              description: "OAuth consumer key",
              default: "",
            },
            consumer_secret: {
              type: "input",
              title: "Consumer Secret",
              description: "OAuth consumer secret",
              default: "",
            },
            access_token: {
              type: "input",
              title: "Access Token",
              description: "OAuth access token",
              default: "",
            },
            access_token_secret: {
              type: "input",
              title: "Access Token Secret",
              description: "OAuth access token secret",
              default: "",
            },
          },
        },
        oauth2: {
          type: "object",
          title: "OAuth2",
          description: "OAuth2 configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth2 authentication",
              default: false,
            },
            client_key: {
              type: "input",
              title: "Client Key",
              description: "OAuth2 client key",
              default: "",
            },
            client_secret: {
              type: "input",
              title: "Client Secret",
              description: "OAuth2 client secret",
              default: "",
            },
            token_url: {
              type: "input",
              title: "Token URL",
              description: "OAuth2 token URL",
              default: "",
            },
            scopes: {
              type: "array",
              title: "Scopes",
              description: "OAuth2 scopes",
              default: [],
            },
            endpoint_params: {
              type: "key_value",
              title: "Endpoint Parameters",
              description: "OAuth2 endpoint parameters",
              default: {},
            },
          },
        },
        basic_auth: {
          type: "object",
          title: "Basic Auth",
          description: "Basic authentication configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable basic authentication",
              default: false,
            },
            username: {
              type: "input",
              title: "Username",
              description: "Basic auth username",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "Basic auth password",
              default: "",
            },
          },
        },
        jwt: {
          type: "object",
          title: "JWT",
          description: "JWT configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable JWT authentication",
              default: false,
            },
            private_key_file: {
              type: "input",
              title: "Private Key File",
              description: "JWT private key file",
              default: "",
            },
            signing_method: {
              type: "input",
              title: "Signing Method",
              description: "JWT signing method",
              default: "",
            },
            claims: {
              type: "key_value",
              title: "Claims",
              description: "JWT claims",
              default: {},
            },
            headers: {
              type: "key_value",
              title: "Headers",
              description: "JWT headers",
              default: {},
            },
          },
        },
        tls: {
          type: "object",
          title: "TLS",
          description: "TLS configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable TLS",
              default: false,
            },
            skip_cert_verify: {
              type: "bool",
              title: "Skip Certificate Verification",
              description: "Skip TLS certificate verification",
              default: false,
            },
            enable_renegotiation: {
              type: "bool",
              title: "Enable Renegotiation",
              description: "Enable TLS renegotiation",
              default: false,
            },
            root_cas: {
              type: "input",
              title: "Root CAs",
              description: "TLS root CAs",
              default: "",
            },
            root_cas_file: {
              type: "input",
              title: "Root CAs File",
              description: "TLS root CAs file",
              default: "",
            },
            client_certs: {
              type: "array",
              title: "Client Certificates",
              description: "TLS client certificates",
              default: [],
            },
          },
        },
        extract_headers: {
          type: "object",
          title: "Extract Headers",
          description: "Extract headers configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include header prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include header patterns",
              default: [],
            },
          },
        },
        rate_limit: {
          type: "input",
          title: "Rate Limit",
          description: "An optional rate limit to apply to the requests.",
          default: "",
        },
        timeout: {
          type: "input",
          title: "Timeout",
          description: "An optional timeout for the request.",
          default: "5s",
        },
        retry_period: {
          type: "input",
          title: "Retry Period",
          description: "Period to wait between retries",
          default: "1s",
        },
        max_retry_backoff: {
          type: "input",
          title: "Max Retry Backoff",
          description: "Maximum backoff time between retries",
          default: "300s",
        },
        retries: {
          type: "number",
          title: "Retries",
          description: "Number of retries",
          default: 3,
        },
        backoff_on: {
          type: "array",
          title: "Backoff On",
          description: "Status codes to backoff on",
          default: [429],
        },
        drop_on: {
          type: "array",
          title: "Drop On",
          description: "Status codes to drop on",
          default: [],
        },
        successful_on: {
          type: "array",
          title: "Successful On",
          description: "Status codes to consider successful",
          default: [],
        },
        proxy_url: {
          type: "input",
          title: "Proxy URL",
          description: "Proxy URL",
          default: "",
        },
      },
    },
    kafka: {
      title: "Kafka",
      properties: {
        addresses: {
          type: "array",
          title: "Addresses",
          description: "A list of broker addresses to connect to.",
          default: [],
        },
        topics: {
          type: "array",
          title: "Topics",
          description: "A list of topics to consume from.",
          default: [],
        },
        consumer_group: {
          type: "input",
          title: "Consumer Group",
          description: "A consumer group ID to use when consuming topics.",
        },
        client_id: {
          type: "input",
          title: "Client ID",
          description: "An identifier for the client connection.",
          default: "",
        },
        rack_id: {
          type: "input",
          title: "Rack ID",
          description: "A rack identifier for this client.",
          default: "",
        },
        start_from_oldest: {
          type: "bool",
          title: "Start From Oldest",
          description:
            "If true, the consumer group will begin at the oldest available offset.",
          default: false,
        },
        checkpoint_limit: {
          type: "number",
          title: "Checkpoint Limit",
          description:
            "The maximum number of messages to process before checkpointing the current offset.",
          default: 1000,
        },
        commit_period: {
          type: "input",
          title: "Commit Period",
          description: "The period of time between offset commits.",
          default: "1s",
        },
        max_processing_period: {
          type: "input",
          title: "Max Processing Period",
          description:
            "A maximum estimate for how long a message takes to be processed.",
          default: "100ms",
        },
        sasl: {
          type: "object",
          title: "SASL",
          description: "SASL authentication configuration.",
          properties: {
            mechanism: {
              type: "select",
              title: "Mechanism",
              description: "The SASL mechanism to use.",
              options: ["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"],
              default: "none",
            },
            user: {
              type: "input",
              title: "Username",
              description: "The SASL username.",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "The SASL password.",
              default: "",
            },
          },
        },
        target_version: {
          type: "input",
          title: "Target Version",
          description: "The version of the Kafka protocol to use.",
          default: "2.0.0",
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Allows you to configure a batching policy.",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description:
                "A number of messages at which the batch should be flushed. If 0 disables count based batching.",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description:
                "An amount of bytes at which the batch should be flushed. If 0 disables size based batching.",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description:
                "A period in which an incomplete batch should be flushed regardless of its size.",
              default: "",
            },
            check: {
              type: "input",
              title: "Check",
              description:
                "A Bloblang query that should return a boolean value indicating whether a message should end a batch.",
              default: "",
            },
          },
        },
      },
    },
    http_server: {
      title: "HTTP Server",
      properties: {
        path: {
          type: "input",
          title: "Path",
          description: "The endpoint path to listen for POST requests.",
          default: "/post",
        },
        allowed_verbs: {
          type: "array",
          title: "Allowed Verbs",
          description:
            "An array of verbs that are allowed for the path endpoint.",
          default: ["POST"],
        },
        timeout: {
          type: "input",
          title: "Timeout",
          description:
            "Timeout for requests. If a consumed messages takes longer than this to be delivered the connection is closed, but the message may still be delivered.",
          default: "5s",
        },
        sync_response: {
          type: "object",
          title: "Synchronous responses",
          description: "Customise messages returned via synchronous responses.",
          properties: {
            status: {
              type: "input",
              title: "HTTP Status code",
              description:
                "Specify the status code to return with synchronous responses",
              default: "200",
            },
            headers: {
              type: "key_value",
              title: "Headers",
              description: "A map of headers to add to the response.",
              default: {},
            },
          },
        },
      },
    },
    mcp_tool: {
      title: "MCP Tool",
      properties: {
        name: {
          type: "input",
          title: "Tool Name",
          description:
            "Unique identifier for the MCP tool. This is the name AI assistants will use to call it.",
          required: true,
          pattern: "[a-zA-Z0-9_-]*",
          patternMessage: "Only letters, numbers, underscores and hyphens allowed",
        },
        description: {
          type: "input",
          title: "Description",
          description:
            "Human-readable description of what the tool does. This is shown to AI assistants.",
          required: true,
        },
        input_schema: {
          type: "property_list",
          title: "Input Parameters",
          description:
            "Define the parameters that AI assistants will pass when calling this tool.",
        },
      },
    },
    broker: {
      title: "Broker",
      description:
        "Allows you to combine multiple inputs into a single flow using a range of input brokers.",
      properties: {
        copies: {
          type: "number",
          title: "Copies",
          description:
            "The number of copies of each configured input to spawn.",
          default: 1,
        },
        inputs: {
          type: "input_list",
          title: "Inputs",
          description: "A list of child inputs to broker.",
          required: true,
          default: [],
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Allows you to configure a batching policy.",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description:
                "A number of messages at which the batch should be flushed. If 0 disables count based batching.",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description:
                "An amount of bytes at which the batch should be flushed. If 0 disables size based batching.",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description:
                "A period in which an incomplete batch should be flushed regardless of its size.",
              default: "",
            },
            jitter: {
              type: "number",
              title: "Jitter",
              description:
                "A non-negative factor that adds random delay to batch flush intervals.",
              default: 0,
            },
            check: {
              type: "input",
              title: "Check",
              description:
                "A Bloblang query that should return a boolean value indicating whether a message should end a batch.",
              default: "",
            },
            processors: {
              type: "processor_list",
              title: "Processors",
              description:
                "A list of processors to apply to a batch as it is flushed.",
              default: [],
            },
          },
        },
      },
    },
    cdc_mysql: {
      title: "MySQL (MariaDB) CDC (Experimental)",
      properties: {
        host: {
          type: "input",
          title: "Host",
          description: "MySQL server hostname or IP address.",
          required: true,
        },
        port: {
          type: "number",
          title: "Port",
          description: "MySQL server port.",
          default: 3306,
        },
        user: {
          type: "input",
          title: "User",
          description: "MySQL username for connection.",
          required: true,
        },
        password: {
          type: "input",
          title: "Password",
          description: "MySQL password for connection.",
          default: "",
          required: true,
        },
        server_id: {
          type: "input",
          title: "Server ID",
          description:
            "Unique server ID for this binlog consumer. Corresponds to MySQL's server_id. Can be an integer or string.",
          default: "1000",
        },
        position_cache: {
          type: "dynamic_select",
          title: "Position Cache",
          description: "Cache resource to use for position tracking.",
          dataSource: "caches",
          required: true,
        },
        position_cache_key: {
          type: "input",
          title: "Position Cache Key",
          description: "Cache key to use for storing position information.",
          required: true,
        },
        position_mode: {
          type: "select",
          title: "Position Mode",
          description: "Position tracking mode: 'gtid' (default) or 'file'.",
          options: ["gtid", "file"],
          default: "gtid",
        },
        flavor: {
          type: "select",
          title: "Flavor",
          description: "Database flavor: 'mysql' (default) or 'mariadb'.",
          options: ["mysql", "mariadb"],
          default: "mysql",
        },
        cache_save_interval: {
          type: "input",
          title: "Cache Save Interval",
          description:
            "Interval for saving binlog position to cache (e.g., '10s', '1m'). Set to '0s' to save immediately.",
          default: "30s",
        },
        include_tables: {
          type: "array",
          title: "Include Tables",
          description:
            "List of tables to monitor in format 'schema_name.table_name'. If empty, all tables are monitored.",
          default: [],
        },
        exclude_tables: {
          type: "array",
          title: "Exclude Tables",
          description:
            "List of tables to exclude in format 'schema_name.table_name'.",
          default: [],
        },
        use_schema_cache: {
          type: "bool",
          title: "Use Schema Cache",
          description:
            "Enable schema caching to get column names without requiring --binlog-row-metadata=FULL.",
          default: true,
        },
        schema_cache_ttl: {
          type: "input",
          title: "Schema Cache TTL",
          description:
            "TTL for schema cache entries (e.g., '5m', '1h'). Schema is refreshed when cache expires.",
          default: "5m",
        },
        max_batch_size: {
          type: "number",
          title: "Max Batch Size",
          description:
            "Maximum number of messages per batch. Prevents unbounded memory growth for large transactions.",
          default: 1000,
        },
        max_pending_checkpoints: {
          type: "number",
          title: "Max Pending Checkpoints",
          description:
            "Maximum number of pending checkpoints. Provides backpressure when downstream is slow.",
          default: 100,
        },
        retry_initial_interval: {
          type: "input",
          title: "Retry Initial Interval",
          description:
            "Initial wait time before first retry attempt (e.g., '1s', '500ms').",
          default: "1s",
        },
        retry_max_interval: {
          type: "input",
          title: "Retry Max Interval",
          description:
            "Maximum wait time between retry attempts (e.g., '30s', '1m').",
          default: "30s",
        },
        retry_multiplier: {
          type: "number",
          title: "Retry Multiplier",
          description:
            "Multiplier for exponential backoff (e.g., 2.0 for doubling).",
          default: 2.0,
        },
      },
    },
    shopify: {
      title: "Shopify",
      properties: {
        shop_name: {
          type: "input",
          title: "Shop Name",
          description: "Shopify store name (without .myshopify.com).",
          required: true,
        },
        api_key: {
          type: "input",
          title: "API Key",
          description: "Shopify API key for authentication (Private App).",
          required: true,
        },
        api_access_token: {
          type: "input",
          title: "API Access Token",
          description:
            "Shopify API Access Token for authentication (Private App).",
          required: true,
        },
        shop_resource: {
          type: "select",
          title: "Shop Resource",
          description: "The Shopify resource type to fetch.",
          options: [
            "products",
            "orders",
            "customers",
            "inventory_items",
            "locations",
          ],
          default: "products",
        },
        limit: {
          type: "number",
          title: "Limit",
          description:
            "Maximum number of items to fetch per API request (max 250).",
          default: 50,
        },
        api_version: {
          type: "input",
          title: "API Version",
          description:
            "Shopify API version to use (e.g., '2024-01'). If not specified, uses the default version.",
          default: "",
        },
        cache_resource: {
          type: "dynamic_select",
          title: "Cache Resource",
          description:
            "Optional cache resource name for storing the last updated_at timestamp. When configured, resumes fetching from items updated after that timestamp.",
          dataSource: "caches",
          default: "",
        },
        rate_limit: {
          type: "dynamic_select",
          title: "Rate Limit",
          description:
            "Rate limit resource to use for Shopify API requests. Uses shop name as the rate limit key.",
          dataSource: "rate_limits",
          default: "",
        },
      },
    },
    amqp_0_9: {
      title: "AMQP 0.9",
      properties: {
        urls: {
          type: "array",
          title: "URLs",
          description:
            "A list of URLs to connect to. The first URL to successfully establish a connection will be used.",
          default: [],
          required: true,
        },
        queue: {
          type: "input",
          title: "Queue",
          description: "An AMQP queue to consume from.",
          required: true,
        },
        consumer_tag: {
          type: "input",
          title: "Consumer Tag",
          description: "An identifier for the consumer connection.",
          default: "",
        },
        auto_ack: {
          type: "bool",
          title: "Auto Ack",
          description:
            "Acknowledge messages automatically on receipt, skipping downstream acknowledgment.",
          default: false,
        },
        prefetch_count: {
          type: "number",
          title: "Prefetch Count",
          description: "The maximum number of pending messages to have consumed at a time.",
          default: 10,
        },
        nack_reject_patterns: {
          type: "array",
          title: "Nack Reject Patterns",
          description:
            "A list of regular expressions. If a message fails processing and its error matches any pattern, it will be dropped (rejected) instead of requeued.",
          default: [],
        },
        queue_declare: {
          type: "object",
          title: "Queue Declare",
          description: "Optionally declare the queue during connection.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to declare the queue during connection.",
              default: false,
            },
            durable: {
              type: "bool",
              title: "Durable",
              description: "Whether the declared queue is durable.",
              default: true,
            },
            auto_delete: {
              type: "bool",
              title: "Auto Delete",
              description: "Whether the declared queue will auto-delete when there are no consumers.",
              default: false,
            },
          },
        },
      },
    },
  },
  pipeline: {
    ai_gateway: {
      title: "AI Gateway",
      description:
        "Calls an AI chat completion API and maps the response into the message.",
      properties: {
        provider: {
          type: "select",
          title: "Provider",
          description: "The AI provider to use for chat completions.",
          options: ["openai", "anthropic"],
          required: true,
        },
        model: {
          type: "input",
          title: "Model",
          description:
            "The model identifier to use (e.g., 'gpt-4o', 'claude-sonnet-4-6').",
          required: true,
        },
        api_key: {
          type: "input",
          title: "API Key",
          description:
            "API key for authenticating with the AI provider.",
          required: true,
        },
        base_url: {
          type: "input",
          title: "Base URL",
          description:
            "Custom base URL for the API endpoint. When empty, uses the provider's default URL.",
          default: "",
        },
        system_prompt: {
          type: "code",
          title: "System Prompt",
          description:
            "An optional system prompt to set the behavior of the AI model.",
          default: "",
        },
        prompt: {
          type: "code",
          title: "Prompt",
          description:
            "The user prompt template. Use %v placeholders for values provided by args_mapping.",
          required: true,
        },
        args_mapping: {
          type: "code",
          title: "Args Mapping",
          description:
            "A Bloblang mapping which should evaluate to an array of values matching the number of %v placeholders in the prompt.",
        },
        unsafe_dynamic_prompt: {
          type: "bool",
          title: "Unsafe Dynamic Prompt",
          description:
            "When enabled, the prompt and system_prompt fields support interpolation functions like ${!this.field_name}.",
          default: false,
        },
        max_tokens: {
          type: "number",
          title: "Max Tokens",
          description:
            "Maximum number of tokens to generate in the response.",
          default: 1024,
        },
        temperature: {
          type: "number",
          title: "Temperature",
          description:
            "Sampling temperature for the model. Higher values produce more random output.",
          default: 1.0,
        },
        mcp_tools: {
          type: "bool",
          title: "MCP Tools",
          description:
            "When enabled, the AI model can discover and call MCP tools registered in Qaynaq. The model will automatically have access to all active MCP tool flows.",
          default: true,
        },
        mcp_url: {
          type: "input",
          title: "MCP URL",
          description:
            "URL of the Qaynaq MCP endpoint for tool discovery and execution.",
          default: "http://localhost:8080/mcp",
        },
        max_tool_rounds: {
          type: "number",
          title: "Max Tool Rounds",
          description:
            "Maximum number of tool calling rounds before forcing a final response.",
          default: 5,
        },
      },
    },
    google_calendar: {
      title: "Google Calendar",
      description:
        "Performs Google Calendar operations - create, read, update, and delete events and calendars.",
      properties: {
        service_account_json: {
          type: "dynamic_select",
          title: "Service Account JSON",
          description:
            "Google service account credentials JSON. Store as a secret in Settings > Secrets, then select it here.",
          dataSource: "secrets",
        },
        oauth_connection: {
          type: "dynamic_select",
          title: "OAuth Connection",
          description:
            "OAuth connection for user authentication. Set up in Connections page. Alternative to Service Account JSON.",
          dataSource: "connections",
        },
        delegate_to: {
          type: "input",
          title: "Delegate To",
          description:
            "Email address to impersonate via Domain-Wide Delegation. Required for accessing other users' calendars in Google Workspace.",
          default: "",
        },
        action: {
          type: "select",
          title: "Action",
          description: "The calendar operation to perform.",
          options: [
            "add_attendees",
            "create_calendar",
            "create_event",
            "delete_event",
            "find_busy_periods",
            "find_calendars",
            "find_events",
            "find_or_create_event",
            "get_calendar",
            "get_event",
            "move_event",
            "quick_add_event",
            "update_event",
          ],
          required: true,
        },
        calendar_id: {
          type: "input",
          title: "Calendar ID",
          description:
            "Target calendar ID. Use 'primary' for the authenticated user's main calendar.",
          default: "primary",
          required: true,
        },
        event_id: {
          type: "input",
          title: "Event ID",
          description:
            "The event identifier. Required for: get_event, delete_event, update_event, add_attendees, move_event.",
          default: "",
        },
        destination_calendar_id: {
          type: "input",
          title: "Destination Calendar ID",
          description:
            "Target calendar to move the event to. Required for: move_event.",
          default: "",
        },
        summary: {
          type: "input",
          title: "Summary",
          description:
            "Event title. Required for: create_event, find_or_create_event. Optional for: update_event.",
          default: "",
        },
        description: {
          type: "input",
          title: "Description",
          description:
            "Event description text. Used by: create_event, update_event, find_or_create_event.",
          default: "",
        },
        location: {
          type: "input",
          title: "Location",
          description:
            "Event location. Used by: create_event, update_event, find_or_create_event.",
          default: "",
        },
        start_time: {
          type: "input",
          title: "Start Time",
          description:
            "Start time in RFC3339 format (e.g. '2025-01-15T09:00:00-05:00'). Required for: create_event, find_or_create_event, find_busy_periods.",
          default: "",
        },
        end_time: {
          type: "input",
          title: "End Time",
          description:
            "End time in RFC3339 format (e.g. '2025-01-15T10:00:00-05:00'). Required for: create_event, find_or_create_event, find_busy_periods.",
          default: "",
        },
        time_zone: {
          type: "input",
          title: "Time Zone",
          description:
            "IANA time zone (e.g. 'America/New_York'). Defaults to UTC. Used by: create_event, update_event, find_or_create_event.",
          default: "",
        },
        attendees: {
          type: "input",
          title: "Attendees",
          description:
            "Comma-separated attendee email addresses. Supports interpolation (e.g. '${!this.email}'). Used by: create_event, update_event, add_attendees, find_or_create_event.",
          default: "",
        },
        quick_add_text: {
          type: "input",
          title: "Quick Add Text",
          description:
            "Natural language event description that Google will parse (e.g. 'Meeting with John tomorrow at 3pm'). Required for: quick_add_event.",
          default: "",
        },
        query: {
          type: "input",
          title: "Query",
          description:
            "Free text search terms to filter events. Used by: find_events, find_or_create_event.",
          default: "",
        },
        max_results: {
          type: "number",
          title: "Max Results",
          description:
            "Maximum number of results. Used by: find_events (max 2500), find_calendars (max 250).",
          default: 25,
        },
        send_updates: {
          type: "input",
          title: "Send Updates",
          description:
            "Notification policy for attendees: all, externalOnly, none. Used by: create_event, update_event, delete_event, add_attendees, find_or_create_event.",
          default: "none",
        },
        recurrence: {
          type: "input",
          title: "Recurrence",
          description:
            "Comma-separated RRULE recurrence rules (e.g. 'RRULE:FREQ=WEEKLY;COUNT=5'). Used by: create_event, update_event, find_or_create_event.",
          default: "",
        },
        visibility: {
          type: "input",
          title: "Visibility",
          description:
            "Event visibility: default, public, private, confidential. Used by: create_event, update_event, find_or_create_event.",
          default: "default",
        },
        add_conference: {
          type: "bool",
          title: "Add Google Meet",
          description:
            "Auto-generate a Google Meet conference link. Used by: create_event, update_event, find_or_create_event.",
          default: false,
        },
        calendar_summary: {
          type: "input",
          title: "Calendar Name",
          description:
            "Name for the new calendar. Required for: create_calendar.",
          default: "",
        },
      },
    },
    google_drive: {
      title: "Google Drive",
      description:
        "Performs Google Drive operations - manage files, folders, permissions, and shared drives.",
      properties: {
        service_account_json: {
          type: "dynamic_select",
          title: "Service Account JSON",
          description:
            "Google service account credentials JSON. Store as a secret in Settings > Secrets, then select it here.",
          dataSource: "secrets",
        },
        oauth_connection: {
          type: "dynamic_select",
          title: "OAuth Connection",
          description:
            "OAuth connection for user authentication. Set up in Connections page. Alternative to Service Account JSON.",
          dataSource: "connections",
        },
        delegate_to: {
          type: "input",
          title: "Delegate To",
          description:
            "Email address to impersonate via Domain-Wide Delegation. Required for accessing files owned by other users in Google Workspace.",
          default: "",
        },
        action: {
          type: "select",
          title: "Action",
          description: "The Google Drive operation to perform.",
          options: [
            "copy_file",
            "delete_file_permanent",
            "export_file",
            "upload_file",
            "create_folder",
            "move_file",
            "create_file_from_text",
            "remove_file_permission",
            "replace_file",
            "create_shared_drive",
            "add_file_sharing",
            "create_shortcut",
            "update_metadata",
            "rename",
            "delete_file",
            "list_files",
            "get_file",
            "get_permissions",
            "find_file",
            "find_folder",
            "find_or_create_file",
            "find_or_create_folder",
          ],
          required: true,
        },
        file_id: {
          type: "input",
          title: "File ID",
          description:
            "The file or folder ID. Required for most file-specific operations.",
          default: "",
        },
        file_name: {
          type: "input",
          title: "File Name",
          description:
            "File or folder name. Used by: create_folder, create_file_from_text, rename, find_file, find_folder, copy_file, upload_file, create_shortcut, find_or_create_file, find_or_create_folder.",
          default: "",
        },
        folder_id: {
          type: "input",
          title: "Folder ID",
          description:
            "Parent folder ID. Defaults to root.",
          default: "",
        },
        destination_folder_id: {
          type: "input",
          title: "Destination Folder ID",
          description:
            "Target folder ID for move operations. Required for: move_file.",
          default: "",
        },
        mime_type: {
          type: "input",
          title: "MIME Type",
          description:
            "MIME type. Used by: export_file (target format), upload_file, create_file_from_text, find_or_create_file.",
          default: "",
        },
        content: {
          type: "input",
          title: "Content",
          description:
            "Text content for file creation. Used by: create_file_from_text, find_or_create_file.",
          default: "",
        },
        file_url: {
          type: "input",
          title: "File URL",
          description:
            "URL to download file content from. Used by: upload_file, replace_file.",
          default: "",
        },
        description: {
          type: "input",
          title: "Description",
          description:
            "File description. Used by: update_metadata, create_file_from_text, upload_file.",
          default: "",
        },
        email: {
          type: "input",
          title: "Email",
          description:
            "Email address for permission operations. Used by: add_file_sharing, remove_file_permission.",
          default: "",
        },
        role: {
          type: "select",
          title: "Role",
          description:
            "Permission role. Used by: add_file_sharing.",
          options: [
            "reader",
            "writer",
            "commenter",
            "owner",
            "organizer",
            "fileOrganizer",
          ],
          default: "reader",
        },
        permission_type: {
          type: "select",
          title: "Permission Type",
          description:
            "Permission type. Used by: add_file_sharing.",
          options: ["user", "group", "domain", "anyone"],
          default: "user",
        },
        permission_id: {
          type: "input",
          title: "Permission ID",
          description:
            "Permission ID for removal. Used by: remove_file_permission (alternative to email).",
          default: "",
        },
        query: {
          type: "input",
          title: "Query",
          description:
            "Search query using Drive query syntax. Used by: list_files.",
          default: "",
        },
        max_results: {
          type: "input",
          inputType: "number",
          title: "Max Results",
          description:
            "Maximum number of results to return.",
          default: "100",
        },
        starred: {
          type: "boolean",
          title: "Starred",
          description: "Whether the file is starred. Used by: update_metadata.",
          default: false,
        },
        folder_color: {
          type: "input",
          title: "Folder Color",
          description:
            "Folder color as hex (e.g. #FF0000). Used by: update_metadata.",
          default: "",
        },
        custom_properties: {
          type: "input",
          title: "Custom Properties",
          description:
            'JSON object of custom properties (e.g. {"key":"value"}). Used by: update_metadata.',
          default: "",
        },
        shared_drive_name: {
          type: "input",
          title: "Shared Drive Name",
          description:
            "Name for the new shared drive. Required for: create_shared_drive.",
          default: "",
        },
        target_file_id: {
          type: "input",
          title: "Target File ID",
          description:
            "Target file ID for creating a shortcut. Required for: create_shortcut.",
          default: "",
        },
        send_notification: {
          type: "boolean",
          title: "Send Notification",
          description:
            "Send notification email when sharing. Used by: add_file_sharing.",
          default: true,
        },
      },
    },
    google_sheets: {
      title: "Google Sheets",
      description:
        "Performs Google Sheets operations - create, read, update, and format spreadsheets, worksheets, rows, and cells.",
      properties: {
        service_account_json: {
          type: "dynamic_select",
          title: "Service Account JSON",
          description:
            "Google service account credentials JSON. Store as a secret in Settings > Secrets, then select it here.",
          dataSource: "secrets",
        },
        oauth_connection: {
          type: "dynamic_select",
          title: "OAuth Connection",
          description:
            "OAuth connection for user authentication. Set up in Connections page. Alternative to Service Account JSON.",
          dataSource: "connections",
        },
        delegate_to: {
          type: "input",
          title: "Delegate To",
          description:
            "Email address to impersonate via Domain-Wide Delegation. Required for accessing spreadsheets owned by other users in Google Workspace.",
          default: "",
        },
        action: {
          type: "select",
          title: "Action",
          description: "The spreadsheet operation to perform.",
          options: [
            "create_spreadsheet",
            "get_spreadsheet",
            "create_worksheet",
            "find_worksheet",
            "find_or_create_worksheet",
            "delete_worksheet",
            "rename_worksheet",
            "copy_worksheet",
            "change_sheet_properties",
            "create_row",
            "create_rows",
            "create_row_at_top",
            "get_row",
            "get_rows",
            "get_data_range",
            "lookup_row",
            "lookup_rows",
            "update_row",
            "update_rows",
            "clear_rows",
            "delete_rows",
            "create_column",
            "copy_range",
            "sort_range",
            "format_cells",
            "format_row",
            "set_data_validation",
            "create_conditional_formatting",
          ],
          required: true,
        },
        spreadsheet_id: {
          type: "input",
          title: "Spreadsheet ID",
          description:
            "The spreadsheet identifier. Required for all actions.",
          default: "",
        },
        sheet_name: {
          type: "input",
          title: "Sheet Name",
          description: "The worksheet/tab name.",
          default: "Sheet1",
        },
        title: {
          type: "input",
          title: "Title",
          description:
            "Name for new worksheets. Used by: create_worksheet, find_or_create_worksheet.",
          default: "",
        },
        range: {
          type: "input",
          title: "Range",
          description:
            "A1 notation range (e.g. A1:D10). Used by: get_data_range, format_cells, sort_range, set_data_validation, create_conditional_formatting.",
          default: "",
        },
        row_number: {
          type: "input",
          inputType: "number",
          title: "Row Number",
          description: "Specific row number (1-based).",
          default: "0",
        },
        end_row_number: {
          type: "input",
          inputType: "number",
          title: "End Row Number",
          description:
            "End row number for range operations (1-based, inclusive).",
          default: "0",
        },
        column_name: {
          type: "input",
          title: "Column Name",
          description: "Column header name for lookup operations.",
          default: "",
        },
        lookup_value: {
          type: "input",
          title: "Lookup Value",
          description: "Value to search for in the specified column.",
          default: "",
        },
        values: {
          type: "input",
          title: "Values",
          description:
            'JSON array of values for a single row (e.g. ["val1","val2"]).',
          default: "",
        },
        rows: {
          type: "input",
          title: "Rows",
          description: "JSON array of arrays for multiple rows.",
          default: "",
        },
        max_results: {
          type: "input",
          inputType: "number",
          title: "Max Results",
          description: "Maximum number of rows to return.",
          default: "100",
        },
        new_name: {
          type: "input",
          title: "New Name",
          description: "New name for renaming a worksheet.",
          default: "",
        },
        destination_spreadsheet_id: {
          type: "input",
          title: "Destination Spreadsheet ID",
          description:
            "Target spreadsheet ID for copying a worksheet.",
          default: "",
        },
        source_range: {
          type: "input",
          title: "Source Range",
          description:
            "Source range in A1 notation for copy operations.",
          default: "",
        },
        destination_range: {
          type: "input",
          title: "Destination Range",
          description:
            "Destination range in A1 notation for copy operations.",
          default: "",
        },
        paste_type: {
          type: "select",
          title: "Paste Type",
          description: "Paste type for copy_range.",
          options: [
            "PASTE_NORMAL",
            "PASTE_VALUES",
            "PASTE_FORMAT",
            "PASTE_NO_BORDERS",
            "PASTE_FORMULA",
            "PASTE_DATA_VALIDATION",
            "PASTE_CONDITIONAL_FORMATTING",
          ],
          default: "PASTE_NORMAL",
        },
        sort_column_index: {
          type: "input",
          inputType: "number",
          title: "Sort Column Index",
          description: "0-based column index to sort by.",
          default: "0",
        },
        sort_order: {
          type: "select",
          title: "Sort Order",
          description: "Sort direction.",
          options: ["ASCENDING", "DESCENDING"],
          default: "ASCENDING",
        },
        bold: {
          type: "boolean",
          title: "Bold",
          description: "Apply bold formatting.",
          default: false,
        },
        italic: {
          type: "boolean",
          title: "Italic",
          description: "Apply italic formatting.",
          default: false,
        },
        strikethrough: {
          type: "boolean",
          title: "Strikethrough",
          description: "Apply strikethrough formatting.",
          default: false,
        },
        background_color: {
          type: "input",
          title: "Background Color",
          description: "Background color as hex (e.g. #FF0000).",
          default: "",
        },
        foreground_color: {
          type: "input",
          title: "Foreground Color",
          description: "Text color as hex.",
          default: "",
        },
        number_format: {
          type: "input",
          title: "Number Format",
          description: "Number format pattern (e.g. #,##0.00).",
          default: "",
        },
        frozen_rows: {
          type: "input",
          inputType: "number",
          title: "Frozen Rows",
          description: "Number of rows to freeze.",
          default: "0",
        },
        frozen_columns: {
          type: "input",
          inputType: "number",
          title: "Frozen Columns",
          description: "Number of columns to freeze.",
          default: "0",
        },
        hidden: {
          type: "boolean",
          title: "Hidden",
          description: "Whether the sheet is hidden.",
          default: false,
        },
        sheet_position: {
          type: "input",
          inputType: "number",
          title: "Sheet Position",
          description: "Sheet tab position (0-based).",
          default: "0",
        },
        validation_type: {
          type: "input",
          title: "Validation Type",
          description:
            "Data validation type: ONE_OF_LIST, ONE_OF_RANGE, NUMBER_BETWEEN, etc.",
          default: "",
        },
        validation_values: {
          type: "input",
          title: "Validation Values",
          description: "Comma-separated validation values.",
          default: "",
        },
        condition_type: {
          type: "input",
          title: "Condition Type",
          description:
            "Conditional format type: NUMBER_GREATER, TEXT_CONTAINS, CUSTOM_FORMULA, etc.",
          default: "",
        },
        condition_value: {
          type: "input",
          title: "Condition Value",
          description: "Condition value or formula.",
          default: "",
        },
        condition_background_color: {
          type: "input",
          title: "Condition Background Color",
          description:
            "Background color as hex for conditional formatting.",
          default: "",
        },
        include_headers: {
          type: "boolean",
          title: "Include Headers",
          description: "Include header row in get operations.",
          default: true,
        },
      },
    },
    mapping: {
      title: "Mapping",
      flat: true,
      properties: {
        mapping: {
          type: "code",
          title: "Mapping",
          description: "A bloblang mapping to apply to messages.",
          required: true,
        },
      },
    },
    json_schema: {
      title: "JSON Schema",
      properties: {
        schema_path: {
          type: "file",
          title: "Schema File",
          description:
            "Select a file from the file manager containing the JSON schema.",
        },
        schema: {
          type: "code",
          title: "Inline Schema",
          description:
            "A JSON schema to validate messages against. Used if no schema file is selected.",
        },
      },
    },
    branch: {
      title: "Branch",
      description:
        "Executes a list of processors on a copy of the message, optionally transforming the input and mapping the result back.",
      properties: {
        request_map: {
          type: "code",
          title: "Request Map",
          description:
            "A Bloblang mapping that transforms the message before sending it to the branch processors. If empty, the original message is used as-is.",
          default: "",
        },
        processors: {
          type: "processor_list",
          title: "Processors",
          description:
            "A list of processors to execute on the mapped request.",
          default: [],
        },
        result_map: {
          type: "code",
          title: "Result Map",
          description:
            "A Bloblang mapping that merges the branch processor results back into the original message. In this mapping, 'this' refers to the branch result and 'root' refers to the original message.",
          default: "",
        },
      },
    },
    catch: {
      title: "Catch",
      description:
        "Applies a list of child processors only when a previous processing step has failed.",
      flat: true,
      properties: {
        processors: {
          type: "processor_list",
          title: "Processors",
          description:
            "A list of processors to apply when a message fails processing.",
          default: [],
        },
      },
    },
    switch: {
      title: "Switch",
      description: "Conditionally processes messages based on their contents.",
      flat: true,
      properties: {
        switch: {
          type: "processor_cases",
          title: "Cases",
          description:
            "A list of switch cases with conditions and processors to execute.",
          required: true,
          default: [],
        },
      },
    },
    schema_registry_decode: {
      title: "Schema Registry Decode",
      description:
        "Automatically decodes and validates messages with schemas from a Confluent Schema Registry service.",
      properties: {
        url: {
          type: "input",
          title: "URL",
          description: "The base URL of the schema registry service.",
          required: true,
        },
        avro_raw_json: {
          type: "bool",
          title: "Avro Raw JSON",
          description:
            "Whether Avro messages should be decoded into normal JSON rather than Avro JSON.",
          default: false,
        },
        avro_nested_schemas: {
          type: "bool",
          title: "Avro Nested Schemas",
          description:
            "Whether Avro Schemas are nested. If true bento will resolve schema references.",
          default: false,
        },
        oauth: {
          type: "object",
          title: "OAuth",
          description:
            "Allows you to specify open authentication via OAuth version 1.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to use OAuth version 1 in requests.",
              default: false,
            },
            consumer_key: {
              type: "input",
              title: "Consumer Key",
              description:
                "A value used to identify the client to the service provider.",
              default: "",
            },
            consumer_secret: {
              type: "input",
              title: "Consumer Secret",
              description:
                "A secret used to establish ownership of the consumer key.",
              default: "",
            },
            access_token: {
              type: "input",
              title: "Access Token",
              description:
                "A value used to gain access to the protected resources on behalf of the user.",
              default: "",
            },
            access_token_secret: {
              type: "input",
              title: "Access Token Secret",
              description:
                "A secret provided in order to establish ownership of a given access token.",
              default: "",
            },
          },
        },
        basic_auth: {
          type: "object",
          title: "Basic Auth",
          description: "Allows you to specify basic authentication.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to use basic authentication in requests.",
              default: false,
            },
            username: {
              type: "input",
              title: "Username",
              description: "A username to authenticate as.",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "A password to authenticate with.",
              default: "",
            },
          },
        },
        jwt: {
          type: "object",
          title: "JWT",
          description: "Allows you to specify JWT authentication.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to use JWT authentication in requests.",
              default: false,
            },
            private_key_file: {
              type: "input",
              title: "Private Key File",
              description:
                "A file with the PEM encoded via PKCS1 or PKCS8 as private key.",
              default: "",
            },
            signing_method: {
              type: "input",
              title: "Signing Method",
              description:
                "A method used to sign the token such as RS256, RS384, RS512 or EdDSA.",
              default: "",
            },
            claims: {
              type: "key_value",
              title: "Claims",
              description:
                "A value used to identify the claims that issued the JWT.",
              default: {},
            },
            headers: {
              type: "key_value",
              title: "Headers",
              description: "Add optional key/value headers to the JWT.",
              default: {},
            },
          },
        },
        tls: {
          type: "object",
          title: "TLS",
          description:
            "Custom TLS settings can be used to override system defaults.",
          properties: {
            skip_cert_verify: {
              type: "bool",
              title: "Skip Certificate Verification",
              description:
                "Whether to skip server side certificate verification.",
              default: false,
            },
            enable_renegotiation: {
              type: "bool",
              title: "Enable Renegotiation",
              description:
                "Whether to allow the remote server to repeatedly request renegotiation.",
              default: false,
            },
            root_cas: {
              type: "input",
              title: "Root CAs",
              description:
                "An optional root certificate authority to use. This is a string, representing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.",
              default: "",
            },
            root_cas_file: {
              type: "input",
              title: "Root CAs File",
              description:
                "An optional path of a root certificate authority file to use. This is a file, often with a .pem extension, containing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.",
              default: "",
            },
            client_certs: {
              type: "array",
              title: "Client Certificates",
              description:
                "A list of client certificates to use. For each certificate either the fields cert and key, or cert_file and key_file should be specified, but not both.",
              default: [],
            },
          },
        },
      },
    },
    sync_response: {
      title: "Sync Response",
      description:
        "Returns the current message payload as a synchronous response to the input source. Useful with http_server input to send custom responses.",
      properties: {},
    },
    http: {
      title: "HTTP",
      description:
        "Performs an HTTP request using a message batch as the request body, and replaces the original message parts with the body of the response.",
      properties: {
        url: {
          type: "input",
          title: "URL",
          description: "The URL to connect to.",
          required: true,
        },
        verb: {
          type: "select",
          title: "HTTP Verb",
          description: "The HTTP verb to use.",
          options: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"],
          default: "POST",
        },
        headers: {
          type: "key_value",
          title: "Headers",
          description: "A map of headers to add to the request.",
          default: {},
        },
        metadata: {
          type: "object",
          title: "Metadata",
          description: "Metadata configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include metadata prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include metadata patterns",
              default: [],
            },
          },
        },
        oauth: {
          type: "object",
          title: "OAuth",
          description: "OAuth configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth authentication",
              default: false,
            },
            consumer_key: {
              type: "input",
              title: "Consumer Key",
              description: "OAuth consumer key",
              default: "",
            },
            consumer_secret: {
              type: "input",
              title: "Consumer Secret",
              description: "OAuth consumer secret",
              default: "",
            },
            access_token: {
              type: "input",
              title: "Access Token",
              description: "OAuth access token",
              default: "",
            },
            access_token_secret: {
              type: "input",
              title: "Access Token Secret",
              description: "OAuth access token secret",
              default: "",
            },
          },
        },
        oauth2: {
          type: "object",
          title: "OAuth2",
          description: "OAuth2 configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth2 authentication",
              default: false,
            },
            client_key: {
              type: "input",
              title: "Client Key",
              description: "OAuth2 client key",
              default: "",
            },
            client_secret: {
              type: "input",
              title: "Client Secret",
              description: "OAuth2 client secret",
              default: "",
            },
            token_url: {
              type: "input",
              title: "Token URL",
              description: "OAuth2 token URL",
              default: "",
            },
            scopes: {
              type: "array",
              title: "Scopes",
              description: "OAuth2 scopes",
              default: [],
            },
            endpoint_params: {
              type: "key_value",
              title: "Endpoint Parameters",
              description: "OAuth2 endpoint parameters",
              default: {},
            },
          },
        },
        basic_auth: {
          type: "object",
          title: "Basic Auth",
          description: "Basic authentication configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable basic authentication",
              default: false,
            },
            username: {
              type: "input",
              title: "Username",
              description: "Basic auth username",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "Basic auth password",
              default: "",
            },
          },
        },
        jwt: {
          type: "object",
          title: "JWT",
          description: "JWT configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable JWT authentication",
              default: false,
            },
            private_key_file: {
              type: "input",
              title: "Private Key File",
              description: "JWT private key file",
              default: "",
            },
            signing_method: {
              type: "input",
              title: "Signing Method",
              description: "JWT signing method",
              default: "",
            },
            claims: {
              type: "key_value",
              title: "Claims",
              description: "JWT claims",
              default: {},
            },
            headers: {
              type: "key_value",
              title: "Headers",
              description: "JWT headers",
              default: {},
            },
          },
        },
        extract_headers: {
          type: "object",
          title: "Extract Headers",
          description: "Extract headers configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include header prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include header patterns",
              default: [],
            },
          },
        },
        rate_limit: {
          type: "input",
          title: "Rate Limit",
          description: "An optional rate limit to apply to the requests.",
          default: "",
        },
        timeout: {
          type: "input",
          title: "Timeout",
          description: "An optional timeout for the request.",
          default: "5s",
        },
        retry_period: {
          type: "input",
          title: "Retry Period",
          description: "Period to wait between retries",
          default: "1s",
        },
        max_retry_backoff: {
          type: "input",
          title: "Max Retry Backoff",
          description: "Maximum backoff time between retries",
          default: "300s",
        },
        retries: {
          type: "number",
          title: "Retries",
          description: "Number of retries",
          default: 3,
        },
        backoff_on: {
          type: "array",
          title: "Backoff On",
          description: "Status codes to backoff on",
          default: [429],
        },
        drop_on: {
          type: "array",
          title: "Drop On",
          description: "Status codes to drop on",
          default: [],
        },
        successful_on: {
          type: "array",
          title: "Successful On",
          description: "Status codes to consider successful",
          default: [],
        },
        proxy_url: {
          type: "input",
          title: "Proxy URL",
          description: "Proxy URL",
          default: "",
        },
        payload: {
          type: "code",
          title: "Payload",
          description:
            "An optional Bloblang mapping to create the request payload. If not specified, the raw message payload is used.",
          default: "",
        },
        batch_as_multipart: {
          type: "bool",
          title: "Batch as Multipart",
          description: "Whether to send batched messages as multipart requests",
          default: false,
        },
        parallel: {
          type: "bool",
          title: "Parallel",
          description:
            "When processing batched messages, whether to send messages in parallel",
          default: false,
        },
      },
    },
    sql_raw: {
      title: "SQL Raw",
      description:
        "Runs an arbitrary SQL query against a database and replaces the message with the result.",
      properties: {
        driver: {
          type: "select",
          title: "Driver",
          description: "A database driver to use.",
          options: [
            "mysql",
            "postgres",
            "clickhouse",
            "mssql",
            "sqlite",
            "oracle",
            "snowflake",
            "trino",
            "gocosmos",
            "spanner",
          ],
          required: true,
        },
        dsn: {
          type: "input",
          title: "DSN",
          description: "A Data Source Name to identify the target database.",
          required: true,
        },
        query: {
          type: "code",
          title: "Query",
          description:
            "The query to execute. Placeholder arguments are populated with the args_mapping field.",
          required: true,
        },
        unsafe_dynamic_query: {
          type: "bool",
          title: "Unsafe Dynamic Query",
          description:
            "Whether to enable interpolation functions in the query. WARNING: may be susceptible to SQL injection attacks.",
          default: false,
        },
        args_mapping: {
          type: "code",
          title: "Args Mapping",
          description:
            "A Bloblang mapping which should evaluate to an array of values matching placeholder arguments in the query.",
        },
        exec_only: {
          type: "bool",
          title: "Exec Only",
          description:
            "Whether the query result should be discarded. When set to true the message contents will remain unchanged. This is useful for INSERT, UPDATE, or DELETE statements.",
          default: false,
        },
        init_statement: {
          type: "code",
          title: "Init Statement",
          description:
            "An optional SQL statement to execute immediately upon the first connection to the target database. This is a useful way to initialize tables.",
        },
        conn_max_idle_time: {
          type: "input",
          title: "Conn Max Idle Time",
          description:
            "An optional maximum amount of time a connection may be idle. Expired connections may be closed lazily before reuse. If empty connections are not closed due to idle time.",
        },
        conn_max_life_time: {
          type: "input",
          title: "Conn Max Life Time",
          description:
            "An optional maximum amount of time a connection may be reused. Expired connections may be closed lazily before reuse. If empty connections are not closed due to their age.",
        },
        conn_max_idle: {
          type: "number",
          title: "Conn Max Idle",
          description:
            "An optional maximum number of connections in the idle connection pool. If conn_max_open is greater than 0 but less than the new conn_max_idle, then the new conn_max_idle will be reduced to match the conn_max_open limit. If 0 no idle connections are retained.",
          default: 2,
        },
        conn_max_open: {
          type: "number",
          title: "Conn Max Open",
          description:
            "An optional maximum number of open connections to the database. If 0 there is no limit on the number of open connections.",
          default: 0,
        },
      },
    },
    sql_select: {
      title: "SQL Select",
      description:
        "Runs a select query against a database and replaces the message with the rows returned.",
      properties: {
        driver: {
          type: "select",
          title: "Driver",
          description: "A database driver to use.",
          options: [
            "mysql",
            "postgres",
            "clickhouse",
            "mssql",
            "sqlite",
            "oracle",
            "snowflake",
            "trino",
            "gocosmos",
            "spanner",
          ],
          required: true,
        },
        dsn: {
          type: "input",
          title: "DSN",
          description: "A Data Source Name to identify the target database.",
          required: true,
        },
        table: {
          type: "input",
          title: "Table",
          description: "The table to query.",
          required: true,
        },
        columns: {
          type: "array",
          title: "Columns",
          description: "A list of columns to query.",
          required: true,
          default: [],
        },
        where: {
          type: "code",
          title: "Where",
          description:
            "An optional where clause to add. Placeholder arguments are populated with the args_mapping field.",
        },
        args_mapping: {
          type: "code",
          title: "Args Mapping",
          description:
            "A Bloblang mapping which should evaluate to an array of values matching in size to the number of placeholder arguments in the where clause.",
        },
        prefix: {
          type: "input",
          title: "Prefix",
          description:
            "An optional prefix to prepend to the query (before SELECT).",
        },
        suffix: {
          type: "input",
          title: "Suffix",
          description: "An optional suffix to append to the select query.",
        },
        init_statement: {
          type: "code",
          title: "Init Statement",
          description:
            "An optional SQL statement to execute immediately upon the first connection to the target database.",
        },
        conn_max_idle_time: {
          type: "input",
          title: "Conn Max Idle Time",
          description:
            "An optional maximum amount of time a connection may be idle.",
        },
        conn_max_life_time: {
          type: "input",
          title: "Conn Max Life Time",
          description:
            "An optional maximum amount of time a connection may be reused.",
        },
        conn_max_idle: {
          type: "number",
          title: "Conn Max Idle",
          description:
            "An optional maximum number of connections in the idle connection pool.",
          default: 2,
        },
        conn_max_open: {
          type: "number",
          title: "Conn Max Open",
          description:
            "An optional maximum number of open connections to the database. If 0 there is no limit.",
          default: 0,
        },
      },
    },
    sql_insert: {
      title: "SQL Insert",
      description:
        "Inserts a row into an SQL database for each message.",
      properties: {
        driver: {
          type: "select",
          title: "Driver",
          description: "A database driver to use.",
          options: [
            "mysql",
            "postgres",
            "clickhouse",
            "mssql",
            "sqlite",
            "oracle",
            "snowflake",
            "trino",
            "gocosmos",
            "spanner",
          ],
          required: true,
        },
        dsn: {
          type: "input",
          title: "DSN",
          description: "A Data Source Name to identify the target database.",
          required: true,
        },
        table: {
          type: "input",
          title: "Table",
          description: "The table to insert into.",
          required: true,
        },
        columns: {
          type: "array",
          title: "Columns",
          description: "A list of columns to insert.",
          required: true,
          default: [],
        },
        args_mapping: {
          type: "code",
          title: "Args Mapping",
          description:
            "A Bloblang mapping which should evaluate to an array of values matching in size to the number of columns specified.",
          required: true,
        },
        prefix: {
          type: "input",
          title: "Prefix",
          description:
            "An optional prefix to prepend to the insert query (before INSERT).",
        },
        suffix: {
          type: "code",
          title: "Suffix",
          description:
            "An optional suffix to add to the insert query. For example: ON CONFLICT (name) DO NOTHING for PostgreSQL.",
        },
        init_statement: {
          type: "code",
          title: "Init Statement",
          description:
            "An optional SQL statement to execute immediately upon the first connection to the target database.",
        },
        conn_max_idle_time: {
          type: "input",
          title: "Conn Max Idle Time",
          description:
            "An optional maximum amount of time a connection may be idle.",
        },
        conn_max_life_time: {
          type: "input",
          title: "Conn Max Life Time",
          description:
            "An optional maximum amount of time a connection may be reused.",
        },
        conn_max_idle: {
          type: "number",
          title: "Conn Max Idle",
          description:
            "An optional maximum number of connections in the idle connection pool.",
          default: 2,
        },
        conn_max_open: {
          type: "number",
          title: "Conn Max Open",
          description:
            "An optional maximum number of open connections to the database. If 0 there is no limit.",
          default: 0,
        },
      },
    },
  },
  output: {
    http_client: {
      title: "HTTP Client",
      properties: {
        url: {
          type: "input",
          title: "URL",
          description: "The URL to send messages to.",
          required: true,
        },
        verb: {
          type: "select",
          title: "HTTP Verb",
          description: "The HTTP verb to use.",
          options: ["GET", "POST", "PUT", "DELETE"],
          default: "POST",
        },
        headers: {
          type: "key_value",
          title: "Headers",
          description: "A map of headers to add to the request.",
          default: {},
        },
        metadata: {
          type: "object",
          title: "Metadata",
          description: "Metadata configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include metadata prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include metadata patterns",
              default: [],
            },
          },
        },
        dump_request_log_level: {
          type: "input",
          title: "Dump Request Log Level",
          description: "Log level to dump request details",
          default: "",
        },
        oauth: {
          type: "object",
          title: "OAuth",
          description: "OAuth configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth authentication",
              default: false,
            },
            consumer_key: {
              type: "input",
              title: "Consumer Key",
              description: "OAuth consumer key",
              default: "",
            },
            consumer_secret: {
              type: "input",
              title: "Consumer Secret",
              description: "OAuth consumer secret",
              default: "",
            },
            access_token: {
              type: "input",
              title: "Access Token",
              description: "OAuth access token",
              default: "",
            },
            access_token_secret: {
              type: "input",
              title: "Access Token Secret",
              description: "OAuth access token secret",
              default: "",
            },
          },
        },
        oauth2: {
          type: "object",
          title: "OAuth2",
          description: "OAuth2 configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable OAuth2 authentication",
              default: false,
            },
            client_key: {
              type: "input",
              title: "Client Key",
              description: "OAuth2 client key",
              default: "",
            },
            client_secret: {
              type: "input",
              title: "Client Secret",
              description: "OAuth2 client secret",
              default: "",
            },
            token_url: {
              type: "input",
              title: "Token URL",
              description: "OAuth2 token URL",
              default: "",
            },
            scopes: {
              type: "array",
              title: "Scopes",
              description: "OAuth2 scopes",
              default: [],
            },
            endpoint_params: {
              type: "key_value",
              title: "Endpoint Parameters",
              description: "OAuth2 endpoint parameters",
              default: {},
            },
          },
        },
        basic_auth: {
          type: "object",
          title: "Basic Auth",
          description: "Basic authentication configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable basic authentication",
              default: false,
            },
            username: {
              type: "input",
              title: "Username",
              description: "Basic auth username",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "Basic auth password",
              default: "",
            },
          },
        },
        jwt: {
          type: "object",
          title: "JWT",
          description: "JWT configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable JWT authentication",
              default: false,
            },
            private_key_file: {
              type: "input",
              title: "Private Key File",
              description: "JWT private key file",
              default: "",
            },
            signing_method: {
              type: "input",
              title: "Signing Method",
              description: "JWT signing method",
              default: "",
            },
            claims: {
              type: "key_value",
              title: "Claims",
              description: "JWT claims",
              default: {},
            },
            headers: {
              type: "key_value",
              title: "Headers",
              description: "JWT headers",
              default: {},
            },
          },
        },
        tls: {
          type: "object",
          title: "TLS",
          description: "TLS configuration",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Enable TLS",
              default: false,
            },
            skip_cert_verify: {
              type: "bool",
              title: "Skip Certificate Verification",
              description: "Skip TLS certificate verification",
              default: false,
            },
            enable_renegotiation: {
              type: "bool",
              title: "Enable Renegotiation",
              description: "Enable TLS renegotiation",
              default: false,
            },
            root_cas: {
              type: "input",
              title: "Root CAs",
              description: "TLS root CAs",
              default: "",
            },
            root_cas_file: {
              type: "input",
              title: "Root CAs File",
              description: "TLS root CAs file",
              default: "",
            },
            client_certs: {
              type: "array",
              title: "Client Certificates",
              description: "TLS client certificates",
              default: [],
            },
          },
        },
        extract_headers: {
          type: "object",
          title: "Extract Headers",
          description: "Extract headers configuration",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include header prefixes",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include header patterns",
              default: [],
            },
          },
        },
        rate_limit: {
          type: "input",
          title: "Rate Limit",
          description: "An optional rate limit to apply to the requests.",
          default: "",
        },
        timeout: {
          type: "input",
          title: "Timeout",
          description: "An optional timeout for the request.",
          default: "5s",
        },
        retry_period: {
          type: "input",
          title: "Retry Period",
          description: "Period to wait between retries",
          default: "1s",
        },
        max_retry_backoff: {
          type: "input",
          title: "Max Retry Backoff",
          description: "Maximum backoff time between retries",
          default: "300s",
        },
        retries: {
          type: "number",
          title: "Retries",
          description: "Number of retries",
          default: 3,
        },
        backoff_on: {
          type: "array",
          title: "Backoff On",
          description: "Status codes to backoff on",
          default: [429],
        },
        drop_on: {
          type: "array",
          title: "Drop On",
          description: "Status codes to drop on",
          default: [],
        },
        successful_on: {
          type: "array",
          title: "Successful On",
          description: "Status codes to consider successful",
          default: [],
        },
        proxy_url: {
          type: "input",
          title: "Proxy URL",
          description: "Proxy URL",
          default: "",
        },
        batch_as_multipart: {
          type: "bool",
          title: "Batch as Multipart",
          description: "Whether to send batched messages as multipart requests",
          default: false,
        },
        propagate_response: {
          type: "bool",
          title: "Propagate Response",
          description:
            "Whether to propagate the response from the HTTP request",
          default: false,
        },
        max_in_flight: {
          type: "number",
          title: "Max In Flight",
          description:
            "The maximum number of messages to have in flight at a given time",
          default: 64,
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Batching configuration",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description: "The number of messages to batch together",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description: "The size of messages to batch together",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description: "The period of time to batch messages together",
              default: "",
            },
            check: {
              type: "input",
              title: "Check",
              description: "A condition to check before batching messages",
              default: "",
            },
            processors: {
              type: "array",
              title: "Processors",
              description: "A list of processors to apply to batched messages",
              default: [],
            },
          },
        },
        multipart: {
          type: "array",
          title: "Multipart",
          description: "A list of multipart form fields",
          default: [],
        },
      },
    },
    kafka: {
      title: "Kafka",
      properties: {
        addresses: {
          type: "array",
          title: "Addresses",
          description: "A list of broker addresses to connect to.",
          default: [],
        },
        topic: {
          type: "input",
          title: "Topic",
          description: "The topic to produce to.",
        },
        key: {
          type: "input",
          title: "Key",
          description:
            "An optional key to set for each message (interpolated).",
          default: "",
        },
        client_id: {
          type: "input",
          title: "Client ID",
          description: "An identifier for the client connection.",
          default: "",
        },
        max_in_flight: {
          type: "number",
          title: "Max In Flight",
          description:
            "The maximum number of messages to have in flight at a given time.",
          default: 1000,
        },
        ack_replicas: {
          type: "bool",
          title: "Ack Replicas",
          description:
            "Whether to wait for all replicas to acknowledge messages.",
          default: false,
        },
        compression: {
          type: "select",
          title: "Compression",
          description: "The compression algorithm to use.",
          options: ["none", "gzip", "snappy", "lz4", "zstd"],
          default: "none",
        },
        max_message_bytes: {
          type: "number",
          title: "Max Message Bytes",
          description: "The maximum permitted size of a message.",
          default: 1000000,
        },
        target_version: {
          type: "input",
          title: "Target Version",
          description: "The version of the Kafka protocol to use.",
          default: "2.0.0",
        },
        timeout: {
          type: "input",
          title: "Timeout",
          description:
            "The maximum amount of time to wait for an ack before retrying a send.",
          default: "5s",
        },
        metadata: {
          type: "object",
          title: "Metadata",
          description:
            "Specify criteria for which metadata values are added to messages as headers.",
          properties: {
            include_prefixes: {
              type: "array",
              title: "Include Prefixes",
              description: "Include metadata prefixes as message headers.",
              default: [],
            },
            include_patterns: {
              type: "array",
              title: "Include Patterns",
              description: "Include metadata patterns as message headers.",
              default: [],
            },
          },
        },
        tls: {
          type: "object",
          title: "TLS",
          description: "TLS configuration for secure connections.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to use TLS for secure connections.",
              default: false,
            },
            skip_cert_verify: {
              type: "bool",
              title: "Skip Certificate Verification",
              description:
                "Whether to skip verification of the server's certificate chain and host name.",
              default: false,
            },
            enable_renegotiation: {
              type: "bool",
              title: "Enable Renegotiation",
              description:
                "Whether to allow the remote server to change its certificates during client connection.",
              default: false,
            },
            root_cas: {
              type: "input",
              title: "Root CAs",
              description: "A list of paths to root certificate authorities.",
              default: "",
            },
            root_cas_file: {
              type: "input",
              title: "Root CAs File",
              description:
                "Path to a file containing root certificate authorities.",
              default: "",
            },
            client_certs: {
              type: "array",
              title: "Client Certificates",
              description: "A list of client certificates to use.",
              default: [],
            },
          },
        },
        sasl: {
          type: "object",
          title: "SASL",
          description: "SASL authentication configuration.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to use SASL authentication.",
              default: false,
            },
            mechanism: {
              type: "select",
              title: "Mechanism",
              description: "The SASL mechanism to use.",
              options: ["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"],
              default: "none",
            },
            username: {
              type: "input",
              title: "Username",
              description: "The SASL username.",
              default: "",
            },
            password: {
              type: "input",
              title: "Password",
              description: "The SASL password.",
              default: "",
            },
          },
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Batching configuration for message production.",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description: "The number of messages to batch together.",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description: "The size of messages to batch together.",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description: "The period of time to batch messages together.",
              default: "",
            },
            jitter: {
              type: "number",
              title: "Jitter",
              description: "The jitter to apply to the batching period.",
              default: 0,
            },
            check: {
              type: "input",
              title: "Check",
              description: "A condition to check before batching messages.",
              default: "",
            },
            processors: {
              type: "array",
              title: "Processors",
              description: "A list of processors to apply to batched messages.",
              default: [],
            },
          },
        },
      },
    },
    sync_response: {
      title: "Sync Response",
      properties: {},
    },
    switch: {
      title: "Switch",
      description:
        "Route messages to different outputs based on their contents.",
      properties: {
        retry_until_success: {
          type: "bool",
          title: "Retry Until Success",
          description:
            "If a selected output fails to send a message this field determines whether it is reattempted indefinitely.",
          default: false,
        },
        strict_mode: {
          type: "bool",
          title: "Strict Mode",
          description:
            "Whether an error should be reported if no condition is met. If set to true, an error is propagated back to the input level.",
          default: false,
        },
        cases: {
          type: "output_cases",
          title: "Cases",
          description:
            "A list of switch cases, outlining outputs that can be routed to.",
          required: true,
          default: [],
        },
      },
    },
    broker: {
      title: "Broker",
      description:
        "Route messages to multiple child outputs using a range of brokering patterns.",
      properties: {
        copies: {
          type: "number",
          title: "Copies",
          description:
            "The number of copies of each configured output to spawn.",
          default: 1,
        },
        pattern: {
          type: "select",
          title: "Pattern",
          description: "The brokering pattern to use.",
          options: [
            "fan_out",
            "fan_out_fail_fast",
            "fan_out_sequential",
            "fan_out_sequential_fail_fast",
            "round_robin",
            "greedy",
          ],
          default: "fan_out",
        },
        outputs: {
          type: "output_list",
          title: "Outputs",
          description: "A list of child outputs to broker.",
          required: true,
          default: [],
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Allows you to configure a batching policy.",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description:
                "A number of messages at which the batch should be flushed. If 0 disables count based batching.",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description:
                "An amount of bytes at which the batch should be flushed. If 0 disables size based batching.",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description:
                "A period in which an incomplete batch should be flushed regardless of its size.",
              default: "",
            },
            jitter: {
              type: "number",
              title: "Jitter",
              description:
                "A non-negative factor that adds random delay to batch flush intervals.",
              default: 0,
            },
            check: {
              type: "input",
              title: "Check",
              description:
                "A Bloblang query that should return a boolean value indicating whether a message should end a batch.",
              default: "",
            },
            processors: {
              type: "processor_list",
              title: "Processors",
              description:
                "A list of processors to apply to a batch as it is flushed.",
              default: [],
            },
          },
        },
      },
    },
    sql_insert: {
      title: "SQL Insert",
      description: "Inserts a row into an SQL database for each message.",
      properties: {
        driver: {
          type: "select",
          title: "Driver",
          description: "A database driver to use.",
          options: [
            "mysql",
            "postgres",
            "clickhouse",
            "mssql",
            "sqlite",
            "oracle",
            "snowflake",
            "trino",
            "gocosmos",
            "spanner",
          ],
          required: true,
        },
        dsn: {
          type: "input",
          title: "DSN",
          description: "A Data Source Name to identify the target database.",
          required: true,
        },
        table: {
          type: "input",
          title: "Table",
          description: "The table to insert into.",
          required: true,
        },
        columns: {
          type: "array",
          title: "Columns",
          description: "A list of columns to insert.",
          required: true,
          default: [],
        },
        args_mapping: {
          type: "code",
          title: "Args Mapping",
          description:
            "A Bloblang mapping which should evaluate to an array of values matching in size to the number of columns specified.",
          required: true,
        },
        suffix: {
          type: "code",
          title: "Suffix",
          description:
            "An optional suffix to add to the insert query. For example: ON CONFLICT (name) DO NOTHING for PostgreSQL.",
          default: "",
        },
        max_in_flight: {
          type: "number",
          title: "Max In Flight",
          description: "The maximum number of inserts to run in parallel.",
          default: 64,
        },
        batching: {
          type: "object",
          title: "Batching",
          description: "Allows you to configure a batching policy.",
          properties: {
            count: {
              type: "number",
              title: "Count",
              description:
                "A number of messages at which the batch should be flushed. If 0 disables count based batching.",
              default: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description:
                "An amount of bytes at which the batch should be flushed. If 0 disables size based batching.",
              default: 0,
            },
            period: {
              type: "input",
              title: "Period",
              description:
                "A period in which an incomplete batch should be flushed regardless of its size.",
              default: "",
            },
            jitter: {
              type: "number",
              title: "Jitter",
              description:
                "A non-negative factor that adds random delay to batch flush intervals.",
              default: 0,
            },
            check: {
              type: "input",
              title: "Check",
              description:
                "A Bloblang query that should return a boolean value indicating whether a message should end a batch.",
              default: "",
            },
          },
        },
      },
    },
    amqp_0_9: {
      title: "AMQP 0.9",
      properties: {
        urls: {
          type: "array",
          title: "URLs",
          description:
            "A list of URLs to connect to. The first URL to successfully establish a connection will be used.",
          default: [],
          required: true,
        },
        exchange: {
          type: "input",
          title: "Exchange",
          description: "An AMQP exchange to publish to. Supports interpolation functions.",
          required: true,
        },
        key: {
          type: "input",
          title: "Routing Key",
          description: "The binding key for each message. Supports interpolation functions.",
          default: "",
        },
        type: {
          type: "input",
          title: "Type",
          description: "The type property for each message. Supports interpolation functions.",
          default: "",
        },
        content_type: {
          type: "input",
          title: "Content Type",
          description: "The content type attribute for published messages.",
          default: "application/octet-stream",
        },
        persistent: {
          type: "bool",
          title: "Persistent",
          description: "Whether messages should be sent with persistent delivery mode.",
          default: false,
        },
        max_in_flight: {
          type: "number",
          title: "Max In Flight",
          description: "The maximum number of messages to have in flight at a given time.",
          default: 64,
        },
        exchange_declare: {
          type: "object",
          title: "Exchange Declare",
          description: "Optionally declare the target exchange during connection.",
          properties: {
            enabled: {
              type: "bool",
              title: "Enabled",
              description: "Whether to declare the exchange during connection.",
              default: false,
            },
            type: {
              type: "select",
              title: "Type",
              description: "The type of exchange to declare.",
              options: ["direct", "fanout", "topic", "x-custom"],
              default: "direct",
            },
            durable: {
              type: "bool",
              title: "Durable",
              description: "Whether the declared exchange is durable.",
              default: true,
            },
          },
        },
      },
    },
  },
  cache: {
    memory: {
      title: "Memory",
      properties: {
        default_ttl: {
          type: "string",
          title: "Default TTL",
          description:
            "The default TTL of each item. After this period an item will be eligible for removal during the next compaction.",
          default: "5m",
        },
        compaction_interval: {
          type: "string",
          title: "Compaction Interval",
          description:
            "The period of time to wait before each compaction, at which point expired items are removed. Set to empty string to disable compactions/expiry.",
          default: "60s",
        },
        init_values: {
          type: "object",
          title: "Init Values",
          description:
            "A table of key/value pairs that should be present in the cache on initialization.",
          default: "{}",
        },
        shards: {
          type: "number",
          title: "Shards",
          description:
            "A number of logical shards to spread keys across, increasing the shards can have a performance benefit.",
          default: 1,
        },
      },
    },
    redis: {
      title: "Redis",
      properties: {
        url: {
          type: "string",
          title: "URL",
          description:
            "The URL of the target Redis server. Database is optional and is supplied as the URL path.",
          required: true,
        },
        kind: {
          type: "string",
          title: "Kind",
          description:
            "Specifies a simple, cluster-aware, or failover-aware redis client.",
          options: ["simple", "cluster", "failover"],
          default: "simple",
        },
        master: {
          type: "string",
          title: "Master",
          description: "Name of the redis master when kind is failover",
          default: "",
        },
        tls: {
          type: "object",
          title: "TLS",
          description:
            "Custom TLS settings can be used to override system defaults.",
          properties: {
            enabled: {
              type: "boolean",
              title: "Enabled",
              description: "Whether custom TLS settings are enabled.",
              default: false,
            },
            skip_cert_verify: {
              type: "boolean",
              title: "Skip Cert Verify",
              description:
                "Whether to skip server side certificate verification.",
              default: false,
            },
            enable_renegotiation: {
              type: "boolean",
              title: "Enable Renegotiation",
              description:
                "Whether to allow the remote server to repeatedly request renegotiation.",
              default: false,
            },
            root_cas: {
              type: "string",
              title: "Root CAs",
              description: "An optional root certificate authority to use.",
              default: "",
            },
            root_cas_file: {
              type: "string",
              title: "Root CAs File",
              description:
                "An optional path of a root certificate authority file to use.",
              default: "",
            },
            client_certs: {
              type: "array",
              title: "Client Certs",
              description: "A list of client certificates to use.",
              default: "[]",
            },
          },
        },
        prefix: {
          type: "string",
          title: "Prefix",
          description:
            "An optional string to prefix item keys with in order to prevent collisions with similar services.",
          default: "",
        },
        default_ttl: {
          type: "string",
          title: "Default TTL",
          description:
            "An optional default TTL to set for items, calculated from the moment the item is cached.",
          default: "",
        },
        retries: {
          type: "object",
          title: "Retries",
          description:
            "Determine time intervals and cut offs for retry attempts.",
          properties: {
            initial_interval: {
              type: "string",
              title: "Initial Interval",
              description: "The initial period to wait between retry attempts.",
              default: "500ms",
            },
            max_interval: {
              type: "string",
              title: "Max Interval",
              description: "The maximum period to wait between retry attempts.",
              default: "1s",
            },
            max_elapsed_time: {
              type: "string",
              title: "Max Elapsed Time",
              description:
                "The maximum overall period of time to spend on retry attempts before the request is aborted.",
              default: "5s",
            },
          },
        },
      },
    },
    memcached: {
      title: "Memcached",
      properties: {
        addresses: {
          type: "array",
          title: "Addresses",
          description: "A list of addresses of memcached servers to use.",
          required: true,
          default: "[]",
        },
        prefix: {
          type: "string",
          title: "Prefix",
          description:
            "An optional string to prefix item keys with in order to prevent collisions with similar services.",
          default: "",
        },
        default_ttl: {
          type: "string",
          title: "Default TTL",
          description:
            "A default TTL to set for items, calculated from the moment the item is cached.",
          default: "300s",
        },
        retries: {
          type: "object",
          title: "Retries",
          description:
            "Determine time intervals and cut offs for retry attempts.",
          properties: {
            initial_interval: {
              type: "string",
              title: "Initial Interval",
              description: "The initial period to wait between retry attempts.",
              default: "1s",
            },
            max_interval: {
              type: "string",
              title: "Max Interval",
              description: "The maximum period to wait between retry attempts.",
              default: "5s",
            },
            max_elapsed_time: {
              type: "string",
              title: "Max Elapsed Time",
              description:
                "The maximum overall period of time to spend on retry attempts before the request is aborted.",
              default: "30s",
            },
          },
        },
      },
    },
    file: {
      title: "File",
      properties: {
        directory: {
          type: "string",
          title: "Directory",
          description: "The directory within which to store items as files.",
          required: true,
        },
      },
    },
    lru: {
      title: "LRU",
      properties: {
        cap: {
          type: "number",
          title: "Capacity",
          description: "The maximum number of items to store in the cache.",
          required: true,
          default: 1024,
        },
        init_values: {
          type: "object",
          title: "Init Values",
          description:
            "A table of key/value pairs that should be present in the cache on initialization.",
          default: "{}",
        },
      },
    },
    ttlru: {
      title: "TTLRU",
      properties: {
        cap: {
          type: "number",
          title: "Capacity",
          description: "The maximum number of items to store in the cache.",
          required: true,
          default: 1024,
        },
        default_ttl: {
          type: "string",
          title: "Default TTL",
          description: "The default TTL of each item.",
          default: "5m",
        },
        init_values: {
          type: "object",
          title: "Init Values",
          description:
            "A table of key/value pairs that should be present in the cache on initialization.",
          default: "{}",
        },
      },
    },
    ristretto: {
      title: "Ristretto",
      properties: {
        default_ttl: {
          type: "string",
          title: "Default TTL",
          description: "The default TTL of each item.",
          default: "5m",
        },
        max_cost: {
          type: "number",
          title: "Max Cost",
          description: "The maximum size of the cache in bytes.",
          default: 1073741824,
        },
        num_counters: {
          type: "number",
          title: "Num Counters",
          description:
            "The number of 4-bit access counters to keep for admission and eviction.",
          default: 10000000,
        },
      },
    },
    noop: {
      title: "Noop",
      properties: {},
    },
  },
  rate_limit: {
    coordinator: {
      title: "Coordinator",
      properties: {
        count: {
          type: "number",
          title: "Count",
          description: "Number of requests allowed per interval.",
          required: true,
          default: 10,
          min: 1,
        },
        interval: {
          type: "select",
          title: "Interval",
          description: "Time interval for rate limiting.",
          options: [
            "1s",
            "5s",
            "10s",
            "30s",
            "1m",
            "5m",
            "10m",
            "30m",
            "1h",
            "2h",
            "6h",
            "12h",
            "24h",
          ],
          default: "1s",
          required: true,
        },
        burst: {
          type: "number",
          title: "Burst",
          description: "Additional burst capacity for handling traffic spikes.",
          default: 0,
          min: 0,
        },
      },
    },
  },
  buffer: {
    memory: {
      title: "Memory",
      description:
        "Stores consumed messages in memory and acknowledges them at the input level. During graceful shutdown, the buffer flushes remaining messages. Delivery is not guaranteed in case of crashes.",
      properties: {
        limit: {
          type: "number",
          title: "Limit",
          description:
            "The maximum buffer size (in bytes) to allow before applying backpressure upstream.",
          required: true,
          default: 524288000,
          min: 1,
        },
        spillover: {
          type: "select",
          title: "Spillover",
          description:
            "Whether to drop incoming messages that will exceed the buffer limit.",
          options: ["true", "false"],
          default: "false",
        },
        batch_policy: {
          type: "object",
          title: "Batch Policy",
          description:
            "Optionally configure a policy to flush buffered messages in batches.",
          properties: {
            enabled: {
              type: "select",
              title: "Enabled",
              description: "Whether to batch messages as they are flushed.",
              options: ["true", "false"],
              default: "false",
            },
            count: {
              type: "number",
              title: "Count",
              description:
                "Number of messages at which the batch should be flushed. If 0 disables count based batching.",
              default: 0,
              min: 0,
            },
            byte_size: {
              type: "number",
              title: "Byte Size",
              description:
                "Amount of bytes at which the batch should be flushed. If 0 disables size based batching.",
              default: 0,
              min: 0,
            },
            period: {
              type: "string",
              title: "Period",
              description:
                "A period in which an incomplete batch should be flushed (e.g., 1s, 1m, 500ms).",
              default: "",
            },
            jitter: {
              type: "number",
              title: "Jitter",
              description:
                "A non-negative factor that adds random delay to batch flush intervals (0-1).",
              default: 0,
              min: 0,
            },
            check: {
              type: "textarea",
              title: "Check",
              description:
                "A Bloblang query that should return a boolean value indicating whether a message should end a batch.",
              default: "",
            },
            processors: {
              type: "textarea",
              title: "Processors",
              description:
                "A list of processors to apply to a batch as it is flushed (YAML array).",
              default: "",
            },
          },
        },
      },
    },
    sqlite: {
      title: "SQLite",
      description:
        "Persists messages to a SQLite database file. Messages are not acknowledged until written to disk, and are not removed until successfully delivered. Preserves at-least-once delivery across restarts.",
      properties: {
        path: {
          type: "string",
          title: "Path",
          description:
            "The path of the database file, which will be created if it does not already exist.",
          required: true,
        },
      },
    },
    system_window: {
      title: "System Window",
      description:
        "Groups messages into time-based windows using system clocks. Supports tumbling, sliding, and hopping windows with configurable lateness handling.",
      properties: {
        timestamp_mapping: {
          type: "textarea",
          title: "Timestamp Mapping",
          description:
            "A Bloblang mapping that provides the timestamp to use for allocating messages to windows.",
          default: "root = now()",
        },
        size: {
          type: "string",
          title: "Size",
          description:
            "A duration string describing the size of each window (e.g., 30s, 10m, 1h).",
          required: true,
        },
        slide: {
          type: "string",
          title: "Slide",
          description:
            "An optional duration for sliding windows. Must be smaller than the size (e.g., 30s, 10m).",
          default: "",
        },
        offset: {
          type: "string",
          title: "Offset",
          description:
            "An optional duration to offset the beginning of each window (e.g., -6h, 30m).",
          default: "",
        },
        allowed_lateness: {
          type: "string",
          title: "Allowed Lateness",
          description:
            "Length of time to wait after a window ends before flushing it (e.g., 10s, 1m).",
          default: "",
        },
      },
    },
  },
};

// Component lists for each type
export const componentLists = {
  input: [
    "generate",
    "http_client",
    "http_server",
    "mcp_tool",
    "kafka",
    "amqp_0_9",
    "broker",
    "cdc_mysql",
    "shopify",
  ],
  pipeline: [
    "ai_gateway",
    "google_calendar",
    "google_drive",
    "google_sheets",
    "branch",
    "mapping",
    "json_schema",
    "catch",
    "switch",
    "schema_registry_decode",
    "sync_response",
    "http",
    "sql_raw",
    "sql_select",
    "sql_insert",
  ],
  output: [
    "http_client",
    "kafka",
    "amqp_0_9",
    "sync_response",
    "switch",
    "broker",
    "sql_insert",
  ],
  cache: [
    "memory",
    "redis",
    "memcached",
    "file",
    "lru",
    "ttlru",
    "ristretto",
    "noop",
  ],
  rate_limit: ["coordinator"],
  buffer: ["memory", "sqlite", "system_window"],
};
