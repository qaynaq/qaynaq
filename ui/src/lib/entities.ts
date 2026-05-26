export type Worker = {
  id: string;
  status: string;
  address: string;
  lastHeartbeat: string;
  activeFlows: number;
  createdAt: string;
};

export type Secret = {
  key: string;
  createdAt: string;
};

export type Connection = {
  name: string;
  provider: string;
  scopes: string[];
  clientId: string;
  clientSecretHint: string;
  // shop is the Shopify shop subdomain for shop-templated providers.
  shop?: string;
  // cloudId is the Atlassian Cloud ID for cloudid-required providers.
  cloudId?: string;
  lastError?: string;
  lastErrorAt?: string;
  firstFailedAt?: string;
  consecutiveFailures: number;
  createdAt: string;
  updatedAt: string;
};

export type Flow = {
  id: string;
  parentID?: string;
  name: string;
  status: string;
  input_label: string;
  input_component: string;
  input_config: string;
  output_label: string;
  output_component: string;
  output_config: string;
  buffer_id?: number;
  processors: Array<{
    label: string;
    component: string;
    config: string;
  }>;
  createdAt: string;
  is_http_server: boolean;
  is_mcp_tool: boolean;
  is_ready: boolean;
  builder_state?: string;
  managed_by?: string;

  // Legacy fields for backward compatibility
  inputLabel?: string;
  inputID?: number;
  input?: string;
  output?: string;
  outputID?: number;
  outputLabel?: string;
  isHttpServer?: boolean;
};

export type FlowProcessor = {
  processorID: number;
  label: string;
};

export type Cache = {
  id: string;
  parentID?: string;
  label: string;
  component: string;
  config: string;
  createdAt: string;
};

export type RateLimit = {
  id: string;
  parentID?: string;
  label: string;
  component: string;
  config: string;
  createdAt: string;
};

export type Buffer = {
  id: string;
  parentID?: string;
  label: string;
  component: string;
  config: string;
  createdAt: string;
};

export type FileEntry = {
  id: string;
  parentID?: string;
  key: string;
  content?: string;
  size: number;
  createdAt: string;
  updatedAt?: string;
};

export type FlowEvent = {
  id: number;
  worker_flow_id: number;
  trace_id: string;
  section: string;
  component_label: string;
  type: string;
  content: string;
  meta: Record<string, any>;
  created_at: string;
};

export type FlowStatusCount = {
  status: string;
  count: number;
};

export type ComponentCount = {
  component: string;
  count: number;
};

export type TimeSeriesPoint = {
  timestamp: string;
  input_events: number;
  output_events: number;
  error_events: number;
};

export type APIToken = {
  id: number;
  name: string;
  token?: string;
  scopes: string[];
  last_used_at?: string;
  created_at: string;
};

export type MCPSettings = {
  protected: boolean;
  auth_enabled: boolean;
  oauth_enabled: boolean;
  tokens: APIToken[];
};

export type OAuthClient = {
  id: string;
  name: string;
  redirect_uris: string[];
  created_at: string;
  last_used_at?: string;
  consented: boolean;
};

export type OAuthConsentRequest = {
  request_id: string;
  client_id: string;
  client_name: string;
  redirect_uri: string;
  scope: string;
  user_email: string;
};

export type OAuthClients = {
  oauth_enabled: boolean;
  clients: OAuthClient[];
};

export type OAuthSession = {
  id: number;
  client_id: string;
  client_name: string;
  user_email: string;
  created_at: string;
  expires_at: string;
};

export type OAuthSessions = {
  oauth_enabled: boolean;
  sessions: OAuthSession[];
};

export type MCPServer = {
  id: number;
  name: string;
  url: string;
  auth_type: string;
  auth_header: string;
  connection_name: string;
  status: string;
  tool_count: number;
  last_error: string;
  last_sync_at?: string;
  created_at: string;
  updated_at: string;
  transport: string;
  catalog_id: string;
  process_state: string;
};

export type MCPCatalogEnvSpec = {
  name: string;
  description: string;
  required: boolean;
  secret: boolean;
  advanced: boolean;
};

export type MCPCatalogEntry = {
  id: string;
  display_name: string;
  description: string;
  docs_url: string;
  maintainer: "official" | "community" | string;
  command: string;
  args: string[];
  env_spec: MCPCatalogEnvSpec[];
};

export type MCPServerLogs = {
  last_error: string;
  stderr: string;
  process_state: string;
};

export type Analytics = {
  total_flows: number;
  flows_by_status: FlowStatusCount[];
  total_input_events: number;
  total_output_events: number;
  total_processor_errors: number;
  active_workers: number;
  total_events: number;
  error_events: number;
  events_over_time: TimeSeriesPoint[];
  top_input_components: ComponentCount[];
  top_output_components: ComponentCount[];
};
