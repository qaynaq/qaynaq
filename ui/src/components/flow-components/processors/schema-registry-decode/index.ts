import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import {
  basicAuthSchema,
  defaultBasicAuth,
  oauthSchema,
  defaultOAuth,
  jwtSchema,
  defaultJwt,
} from "../../shared/auth";
import { tlsSchema, defaultTls } from "../../shared/tls";

const configSchema = z.object({
  url: z.string().min(1, "Required"),
  avro_raw_json: z.boolean(),
  avro_nested_schemas: z.boolean(),
  oauth: oauthSchema,
  basic_auth: basicAuthSchema,
  jwt: jwtSchema,
  tls: tlsSchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  url: "",
  avro_raw_json: false,
  avro_nested_schemas: false,
  oauth: defaultOAuth,
  basic_auth: defaultBasicAuth,
  jwt: defaultJwt,
  tls: defaultTls,
};

const component: FlowComponent<Config> = {
  id: "schema_registry_decode",
  name: "Schema Registry Decode",
  category: "processor",
  description:
    "Automatically decodes and validates messages with schemas from a Confluent Schema Registry service.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
