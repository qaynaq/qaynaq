import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import {
  basicAuthSchema,
  defaultBasicAuth,
  oauthSchema,
  defaultOAuth,
  oauth2Schema,
  defaultOAuth2,
  jwtSchema,
  defaultJwt,
} from "../../shared/auth";
import { tlsSchema, defaultTls } from "../../shared/tls";

const metaSchema = z.object({
  include_prefixes: z.array(z.string()),
  include_patterns: z.array(z.string()),
});

const configSchema = z.object({
  url: z.string().min(1, "Required"),
  verb: z.enum(["GET", "POST", "PUT", "DELETE"]),
  headers: z.record(z.string(), z.string()),
  metadata: metaSchema,
  dump_request_log_level: z.string(),
  oauth: oauthSchema,
  oauth2: oauth2Schema,
  basic_auth: basicAuthSchema,
  jwt: jwtSchema,
  tls: tlsSchema,
  extract_headers: metaSchema,
  rate_limit: z.string(),
  timeout: z.string(),
  retry_period: z.string(),
  max_retry_backoff: z.string(),
  retries: z.number().int().min(0),
  backoff_on: z.array(z.number()),
  drop_on: z.array(z.number()),
  successful_on: z.array(z.number()),
  proxy_url: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  url: "",
  verb: "GET",
  headers: {},
  metadata: { include_prefixes: [], include_patterns: [] },
  dump_request_log_level: "",
  oauth: defaultOAuth,
  oauth2: defaultOAuth2,
  basic_auth: defaultBasicAuth,
  jwt: defaultJwt,
  tls: defaultTls,
  extract_headers: { include_prefixes: [], include_patterns: [] },
  rate_limit: "",
  timeout: "5s",
  retry_period: "1s",
  max_retry_backoff: "300s",
  retries: 3,
  backoff_on: [429],
  drop_on: [],
  successful_on: [],
  proxy_url: "",
};

const component: FlowComponent<Config> = {
  id: "http_client",
  name: "HTTP Client",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
