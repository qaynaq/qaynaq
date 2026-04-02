import {
  Flow,
  FlowEvent,
  Worker,
  Secret,
  Cache,
  Connection,
  RateLimit,
  FileEntry,
  Analytics,
  MCPSettings,
  APIToken,
} from "./entities";
import * as yaml from "js-yaml";

// Placeholder API functions - adapt based on original logic and backend
const API_BASE_URL = "/api/v0";

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem("qaynaq_token");
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

async function handleResponse(response: Response): Promise<Response> {
  if (response.status === 401) {
    localStorage.removeItem("qaynaq_token");
    window.location.href = "/login?error=session_expired";
    throw new Error("Unauthorized - redirecting to login");
  }
  return response;
}

export async function fetchWorkers(): Promise<Worker[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/workers/all`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return data.data.map((worker: any) => ({
      id: worker.id,
      status: worker.status,
      address: worker.address || "N/A",
      lastHeartbeat: worker.last_heartbeat
        ? new Date(worker.last_heartbeat).toLocaleString()
        : "N/A",
      activeFlows: worker.active_flows || 0,
      createdAt: worker.created_at
        ? new Date(worker.created_at).toLocaleString()
        : "N/A",
    }));
  } catch (error) {
    console.error("Error fetching workers:", error);
    throw error;
  }
}

export async function fetchFlows(): Promise<Flow[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows?status=all`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return data?.data.map((flow: any) => ({
      id: flow.id,
      parentID: flow.parent_id || flow.id,
      name: flow.name,
      status: flow.status,
      input_label: flow.input_label,
      input_component: flow.input_component,
      input_config: flow.input_config || "",
      output_label: flow.output_label,
      output_component: flow.output_component,
      output_config: flow.output_config || "",
      buffer_id: flow.buffer_id || undefined,
      processors:
        flow.processors?.map((processor: any) => ({
          label: processor.label,
          component: processor.component,
          config: processor.config || "",
        })) || [],
      createdAt: new Date(flow.created_at).toLocaleString(),
      is_http_server: flow.is_http_server || false,
      is_mcp_tool: flow.is_mcp_tool || false,
      is_ready: flow.is_ready || false,
      builder_state: flow.builder_state || undefined,
      managed_by: flow.managed_by || undefined,
    }));
  } catch (error) {
    console.error("Error fetching flows:", error);
    throw error; // Re-throw error to be handled by the calling component
  }
}

export async function fetchStream(id: string): Promise<Flow> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/${id}`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return {
      id: data.data.id,
      parentID: data.data.parent_id || data.data.id,
      name: data.data.name,
      status: data.data.status,
      input_label: data.data.input_label,
      input_component: data.data.input_component,
      input_config: data.data.input_config || "",
      output_label: data.data.output_label,
      output_component: data.data.output_component,
      output_config: data.data.output_config || "",
      buffer_id: data.data.buffer_id || undefined,
      processors:
        data.data.processors?.map((processor: any) => ({
          label: processor.label,
          component: processor.component,
          config: processor.config || "",
        })) || [],
      createdAt: new Date(data.data.created_at).toLocaleString(),
      is_http_server: data.data.is_http_server || false,
      is_mcp_tool: data.data.is_mcp_tool || false,
      is_ready: data.data.is_ready || false,
      builder_state: data.data.builder_state || undefined,
    };
  } catch (error) {
    console.error("Error fetching flow:", error);
    throw error;
  }
}

export async function validateFlow(flow: {
  input_component: string;
  input_label: string;
  input_config: string;
  output_component: string;
  output_label: string;
  output_config: string;
  processors: Array<{ label: string; component: string; config: string }>;
}): Promise<{ valid: boolean; error?: string }> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/validate`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify(flow),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error("Error validating flow:", error);
    throw error;
  }
}

export async function tryFlow(data: {
  processors: Array<{ label: string; component: string; config: string }>;
  messages: Array<{ content: string }>;
}): Promise<{ outputs: Array<{ content: string }>; error?: string }> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/try`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify(data),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error("Error trying flow:", error);
    throw error;
  }
}

export async function createFlow(flow: {
  name: string;
  status: string;
  input_component: string;
  input_label: string;
  input_config: string;
  output_component: string;
  output_label: string;
  output_config: string;
  buffer_id?: number;
  is_ready?: boolean;
  builder_state?: string;
  managed_by?: string;
  processors: Array<{
    label: string;
    component: string;
    config: string;
  }>;
}): Promise<Flow> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          name: flow.name,
          status: flow.status,
          input_component: flow.input_component,
          input_label: flow.input_label,
          input_config: flow.input_config,
          output_component: flow.output_component,
          output_label: flow.output_label,
          output_config: flow.output_config,
          buffer_id: flow.buffer_id || undefined,
          is_ready: flow.is_ready ?? true,
          builder_state: flow.builder_state || "",
          managed_by: flow.managed_by || undefined,
          processors: flow.processors.map((processor) => ({
            label: processor.label,
            component: processor.component,
            config: processor.config,
          })),
        }),
      }),
    );
    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      throw new Error(errorData?.message || `HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return {
      id: data.data.id,
      parentID: data.data.parent_id || data.data.id,
      name: data.data.name,
      status: data.data.status,
      input_label: data.data.input_label,
      input_component: data.data.input_component,
      input_config: data.data.input_config,
      output_label: data.data.output_label,
      output_component: data.data.output_component,
      output_config: data.data.output_config,
      buffer_id: data.data.buffer_id || undefined,
      processors:
        data.data.processors?.map((processor: any) => ({
          label: processor.label,
          component: processor.component,
          config: processor.config || "",
        })) || [],
      createdAt: new Date(data.data.created_at).toLocaleString(),
      is_http_server: data.data.is_http_server || false,
      is_mcp_tool: data.data.is_mcp_tool || false,
      is_ready: data.data.is_ready || false,
      builder_state: data.data.builder_state || undefined,
      managed_by: data.data.managed_by || undefined,
    };
  } catch (error) {
    console.error("Error creating flow:", error);
    throw error;
  }
}

export async function updateFlow(
  id: string,
  flow: {
    name: string;
    status: string;
    input_component: string;
    input_label: string;
    input_config: string;
    output_component: string;
    output_label: string;
    output_config: string;
    buffer_id?: number;
    is_ready?: boolean;
    builder_state?: string;
    managed_by?: string;
    processors: Array<{
      label: string;
      component: string;
      config: string;
    }>;
  },
): Promise<Flow> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          name: flow.name,
          status: flow.status,
          input_component: flow.input_component,
          input_label: flow.input_label,
          input_config: flow.input_config,
          output_component: flow.output_component,
          output_label: flow.output_label,
          output_config: flow.output_config,
          buffer_id: flow.buffer_id || undefined,
          is_ready: flow.is_ready ?? true,
          builder_state: flow.builder_state || "",
          managed_by: flow.managed_by || undefined,
          processors: flow.processors.map((processor) => ({
            label: processor.label,
            component: processor.component,
            config: processor.config,
          })),
        }),
      }),
    );
    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      throw new Error(errorData?.message || `HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return {
      id: data.data.id,
      parentID: data.data.parent_id || data.data.id,
      name: data.data.name,
      status: data.data.status,
      input_label: data.data.input_label,
      input_component: data.data.input_component,
      input_config: data.data.input_config,
      output_label: data.data.output_label,
      output_component: data.data.output_component,
      output_config: data.data.output_config,
      buffer_id: data.data.buffer_id || undefined,
      processors:
        data.data.processors?.map((processor: any) => ({
          label: processor.label,
          component: processor.component,
          config: processor.config,
        })) || [],
      createdAt: new Date(data.data.created_at).toLocaleString(),
      is_http_server: data.data.is_http_server || false,
      is_mcp_tool: data.data.is_mcp_tool || false,
      is_ready: data.data.is_ready || false,
      builder_state: data.data.builder_state || undefined,
      managed_by: data.data.managed_by || undefined,
    };
  } catch (error) {
    console.error("Error updating flow:", error);
    throw error;
  }
}

export async function deleteFlow(id: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/${id}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting flow:", error);
    throw error;
  }
}

export async function updateFlowStatus(
  id: string,
  status: string,
): Promise<Flow> {
  try {
    const flow = await fetchStream(id);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/flows/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          name: flow.name,
          status: status,
          input_component: flow.input_component,
          input_label: flow.input_label,
          input_config: flow.input_config,
          output_component: flow.output_component,
          output_label: flow.output_label,
          output_config: flow.output_config,
          buffer_id: flow.buffer_id || undefined,
          is_ready: flow.is_ready,
          builder_state: flow.builder_state,
          processors: flow.processors.map((p) => ({
            label: p.label,
            component: p.component,
            config: p.config,
          })),
        }),
      }),
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      const message =
        errorData?.message || `HTTP error! status: ${response.status}`;
      throw new Error(message);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      parentID: data.data.parent_id || data.data.id,
      name: data.data.name,
      status: data.data.status,
      input_label: data.data.input_label,
      input_component: data.data.input_component,
      input_config: data.data.input_config,
      output_label: data.data.output_label,
      output_component: data.data.output_component,
      output_config: data.data.output_config,
      processors:
        data.data.processors?.map((p: any) => ({
          label: p.label,
          component: p.component,
          config: p.config,
        })) || [],
      createdAt: new Date(data.data.created_at).toLocaleString(),
      is_http_server: data.data.input_component === "http_server",
      is_mcp_tool: data.data.input_component === "mcp_tool",
      is_ready: data.data.is_ready || false,
      builder_state: data.data.builder_state || undefined,
    };
  } catch (error) {
    console.error("Error updating flow status:", error);
    throw error;
  }
}

export async function fetchSecrets(): Promise<Secret[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/secrets`, {
        headers: getAuthHeaders(),
      }),
    );
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    if (data.data && Array.isArray(data.data)) {
      return data.data.map((secret: any, index: number) => {
        if (!secret || typeof secret.key === "undefined") {
          throw new Error(`Invalid secret data at index ${index}`);
        }

        return {
          key: secret.key,
          createdAt: secret.created_at
            ? new Date(secret.created_at).toLocaleString()
            : "Unknown",
        };
      });
    } else {
      return [];
    }
  } catch (error) {
    console.error("Error fetching secrets:", error);
    throw error;
  }
}

export async function createSecret(secretData: {
  key: string;
  value: string;
}): Promise<Secret> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/secrets`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          key: secretData.key,
          value: secretData.value,
        }),
      }),
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    return {
      key: secretData.key,
      createdAt: new Date().toLocaleString(),
    };
  } catch (error) {
    console.error("Error creating secret:", error);
    throw error;
  }
}

export async function deleteSecret(key: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/secrets/${key}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting secret:", error);
    throw error;
  }
}

export async function fetchConnections(): Promise<Connection[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/connections`, {
        headers: getAuthHeaders(),
      }),
    );
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    if (data.data && Array.isArray(data.data)) {
      return data.data.map((conn: any) => ({
        name: conn.name,
        provider: conn.provider,
        scopes: conn.scopes || [],
        clientId: conn.client_id || "",
        clientSecretHint: conn.client_secret_hint || "",
        createdAt: conn.created_at
          ? new Date(conn.created_at).toLocaleString()
          : "Unknown",
        updatedAt: conn.updated_at
          ? new Date(conn.updated_at).toLocaleString()
          : "Unknown",
      }));
    } else {
      return [];
    }
  } catch (error) {
    console.error("Error fetching connections:", error);
    throw error;
  }
}

export async function deleteConnection(name: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/connections/${name}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting connection:", error);
    throw error;
  }
}

export type ProviderScope = {
  scope: string;
  label: string;
  description: string;
};

export type Provider = {
  id: string;
  scopes: ProviderScope[];
};

export async function fetchProviders(): Promise<Provider[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/connections/providers`, {
        headers: getAuthHeaders(),
      }),
    );
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    return data.data || [];
  } catch (error) {
    console.error("Error fetching providers:", error);
    throw error;
  }
}

export async function fetchCaches(): Promise<Cache[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/caches`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return data?.data.map((cache: any) => ({
      id: cache.id,
      label: cache.label,
      component: cache.component,
      config: cache.config || "",
      createdAt: new Date(cache.created_at).toLocaleString(),
    }));
  } catch (error) {
    console.error("Error fetching caches:", error);
    throw error;
  }
}

export async function fetchCache(id: string): Promise<Cache> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/caches/${id}`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error fetching cache:", error);
    throw error;
  }
}

export async function createCache(cacheData: {
  label: string;
  component: string;
  config: any;
}): Promise<Cache> {
  try {
    const configYaml = yaml.dump(cacheData.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/caches`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          label: cacheData.label,
          component: cacheData.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error creating cache:", error);
    throw error;
  }
}

export async function updateCache(
  id: string,
  cacheData: {
    label: string;
    component: string;
    config: any;
  },
): Promise<Cache> {
  try {
    const configYaml = yaml.dump(cacheData.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/caches/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          id: parseInt(id),
          label: cacheData.label,
          component: cacheData.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error updating cache:", error);
    throw error;
  }
}

export async function deleteCache(id: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/caches/${id}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting cache:", error);
    throw error;
  }
}

export async function fetchRateLimits(): Promise<RateLimit[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/rate-limits`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return data?.data.map((rateLimit: any) => ({
      id: rateLimit.id,
      label: rateLimit.label,
      component: rateLimit.component,
      config: rateLimit.config || "",
      createdAt: new Date(rateLimit.created_at).toLocaleString(),
    }));
  } catch (error) {
    console.error("Error fetching rate limits:", error);
    throw error;
  }
}

export async function fetchRateLimit(id: string): Promise<RateLimit> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/rate-limits/${id}`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error fetching rate limit:", error);
    throw error;
  }
}

export async function createRateLimit(rateLimitData: {
  label: string;
  component: string;
  config: any;
}): Promise<RateLimit> {
  try {
    const configYaml = yaml.dump(rateLimitData.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/rate-limits`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          label: rateLimitData.label,
          component: rateLimitData.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error creating rate limit:", error);
    throw error;
  }
}

export async function updateRateLimit(
  id: string,
  rateLimitData: {
    label: string;
    component: string;
    config: any;
  },
): Promise<RateLimit> {
  try {
    const configYaml = yaml.dump(rateLimitData.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/rate-limits/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          id: parseInt(id),
          label: rateLimitData.label,
          component: rateLimitData.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config || "",
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error updating rate limit:", error);
    throw error;
  }
}

export async function deleteRateLimit(id: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/rate-limits/${id}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting rate limit:", error);
    throw error;
  }
}

export async function fetchBuffers(): Promise<any[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/buffers`, {
        headers: getAuthHeaders(),
      }),
    );

    const data = await response.json();

    return data.data.map((buffer: any) => ({
      id: buffer.id,
      label: buffer.label,
      component: buffer.component,
      config: buffer.config,
      createdAt: buffer.created_at,
    }));
  } catch (error) {
    console.error("Error fetching buffers:", error);
    throw error;
  }
}

export async function fetchBuffer(id: string): Promise<any> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/buffers/${id}`, {
        headers: getAuthHeaders(),
      }),
    );

    const data = await response.json();

    return {
      id: data.data.id,
      label: data.data.label,
      component: data.data.component,
      config: data.data.config,
      createdAt: data.data.created_at,
    };
  } catch (error) {
    console.error("Error fetching buffer:", error);
    throw error;
  }
}

export async function createBuffer(data: {
  label: string;
  component: string;
  config: any;
}): Promise<any> {
  try {
    const configYaml = yaml.dump(data.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/buffers`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          label: data.label,
          component: data.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.message || `HTTP error! status: ${response.status}`,
      );
    }

    const responseData = await response.json();

    return {
      id: responseData.data.id,
      label: responseData.data.label,
      component: responseData.data.component,
      config: responseData.data.config,
      createdAt: responseData.data.created_at,
    };
  } catch (error) {
    console.error("Error creating buffer:", error);
    throw error;
  }
}

export async function updateBuffer(
  id: string,
  data: {
    label: string;
    component: string;
    config: any;
  },
): Promise<any> {
  try {
    const configYaml = yaml.dump(data.config);

    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/buffers/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          id: Number(id),
          label: data.label,
          component: data.component,
          config: configYaml,
        }),
      }),
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.message || `HTTP error! status: ${response.status}`,
      );
    }

    const responseData = await response.json();

    return {
      id: responseData.data.id,
      label: responseData.data.label,
      component: responseData.data.component,
      config: responseData.data.config,
      createdAt: responseData.data.created_at,
    };
  } catch (error) {
    console.error("Error updating buffer:", error);
    throw error;
  }
}

export async function deleteBuffer(id: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/buffers/${id}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.message || `HTTP error! status: ${response.status}`,
      );
    }
  } catch (error) {
    console.error("Error deleting buffer:", error);
    throw error;
  }
}

export async function fetchFlowEvents(
  flowId: string,
  params: {
    limit: number;
    offset: number;
    startTime: string;
    endTime: string;
  },
): Promise<{ data: FlowEvent[]; total: number }> {
  try {
    const query = new URLSearchParams({
      limit: params.limit.toString(),
      offset: params.offset.toString(),
      start_time: params.startTime,
      end_time: params.endTime,
    });

    const response = await handleResponse(
      await fetch(
        `${API_BASE_URL}/flows/${flowId}/events?${query.toString()}`,
        {
          headers: getAuthHeaders(),
        },
      ),
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();

    return {
      data: (result.data || []).map((event: any) => ({
        id: event.id,
        worker_flow_id: event.workerFlowId,
        trace_id: event.traceId,
        section: event.section,
        component_label: event.componentLabel,
        type: event.type,
        content: event.content,
        meta: event.meta || {},
        created_at: event.createdAt,
      })),
      total: result.total || 0,
    };
  } catch (error) {
    console.error("Error fetching flow events:", error);
    throw error;
  }
}

// File methods

export async function fetchFiles(): Promise<FileEntry[]> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/files`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    return (data?.data || []).map((file: any) => ({
      id: file.id,
      parentID: file.parent_id,
      key: file.key,
      size: file.size || 0,
      createdAt: new Date(file.created_at).toLocaleString(),
      updatedAt: file.updated_at
        ? new Date(file.updated_at).toLocaleString()
        : undefined,
    }));
  } catch (error) {
    console.error("Error fetching files:", error);
    throw error;
  }
}

export async function fetchFile(id: string): Promise<FileEntry> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/files/${id}`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();

    let content = "";
    if (data.data.content) {
      content = atob(data.data.content);
    }

    return {
      id: data.data.id,
      parentID: data.data.parent_id,
      key: data.data.key,
      content,
      size: data.data.size || 0,
      createdAt: new Date(data.data.created_at).toLocaleString(),
      updatedAt: data.data.updated_at
        ? new Date(data.data.updated_at).toLocaleString()
        : undefined,
    };
  } catch (error) {
    console.error("Error fetching file:", error);
    throw error;
  }
}

export async function createFile(fileData: {
  key: string;
  content: string;
}): Promise<FileEntry> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/files`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          key: fileData.key,
          content: btoa(fileData.content),
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      key: data.data.key,
      size: data.data.size || 0,
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error creating file:", error);
    throw error;
  }
}

export async function updateFile(
  id: string,
  fileData: {
    key: string;
    content: string;
  },
): Promise<FileEntry> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/files/${id}`, {
        method: "PUT",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          id: parseInt(id),
          key: fileData.key,
          content: btoa(fileData.content),
        }),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    return {
      id: data.data.id,
      key: data.data.key,
      size: data.data.size || 0,
      createdAt: new Date(data.data.created_at).toLocaleString(),
    };
  } catch (error) {
    console.error("Error updating file:", error);
    throw error;
  }
}

export async function deleteFile(id: string): Promise<void> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/files/${id}`, {
        method: "DELETE",
        headers: getAuthHeaders(),
      }),
    );

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.message || `HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("Error deleting file:", error);
    throw error;
  }
}

export async function fetchSetupStatus(): Promise<{ first_run_complete: boolean }> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/setup/status`, {
      headers: getAuthHeaders(),
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function completeSetup(): Promise<void> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/setup/complete`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: "{}",
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
}

export async function testConnection(
  driver: string,
  connectionString: string,
): Promise<{ ok: boolean; error?: string }> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/setup/test-connection`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({ driver, connection_string: connectionString }),
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function testShopifyConnection(
  shopName: string,
  accessToken: string,
): Promise<{ ok: boolean; error?: string; shop_name?: string }> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/setup/test-shopify-connection`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({
        shop_name: shopName,
        api_access_token: accessToken,
      }),
    }),
  );
  return response.json();
}

export async function fetchAnalytics(): Promise<Analytics> {
  try {
    const response = await handleResponse(
      await fetch(`${API_BASE_URL}/analytics`, {
        headers: getAuthHeaders(),
      }),
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    const num = (v: unknown): number => Number(v) || 0;
    return {
      total_flows: num(data.total_flows),
      flows_by_status: (data.flows_by_status ?? []).map(
        (s: { status: string; count: unknown }) => ({
          status: s.status,
          count: num(s.count),
        }),
      ),
      total_input_events: num(data.total_input_events),
      total_output_events: num(data.total_output_events),
      total_processor_errors: num(data.total_processor_errors),
      active_workers: num(data.active_workers),
      total_events: num(data.total_events),
      error_events: num(data.error_events),
      events_over_time: (data.events_over_time ?? []).map(
        (pt: { timestamp: string; input_events: unknown; output_events: unknown; error_events: unknown }) => ({
          timestamp: pt.timestamp,
          input_events: num(pt.input_events),
          output_events: num(pt.output_events),
          error_events: num(pt.error_events),
        }),
      ),
      top_input_components: (data.top_input_components ?? []).map(
        (c: { component: string; count: unknown }) => ({
          component: c.component,
          count: num(c.count),
        }),
      ),
      top_output_components: (data.top_output_components ?? []).map(
        (c: { component: string; count: unknown }) => ({
          component: c.component,
          count: num(c.count),
        }),
      ),
    };
  } catch (error) {
    console.error("Error fetching analytics:", error);
    throw error;
  }
}

export async function fetchMCPSettings(): Promise<MCPSettings> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/settings/mcp`, {
      headers: getAuthHeaders(),
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function updateMCPProtected(
  protected_: boolean,
): Promise<{ protected: boolean }> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/settings/mcp`, {
      method: "PUT",
      headers: getAuthHeaders(),
      body: JSON.stringify({ protected: protected_ }),
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function createAPIToken(
  name: string,
  scopes: string[],
): Promise<APIToken> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/settings/mcp/tokens`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({ name, scopes }),
    }),
  );
  if (!response.ok) {
    if (response.status === 409) {
      throw new Error("A token with this name already exists");
    }
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  const result = await response.json();
  return result.data;
}

export async function deleteAPIToken(id: number): Promise<void> {
  const response = await handleResponse(
    await fetch(`${API_BASE_URL}/settings/mcp/tokens/${id}`, {
      method: "DELETE",
      headers: getAuthHeaders(),
    }),
  );
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
}
