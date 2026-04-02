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
  tokens: APIToken[];
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
