import { z } from "zod";

export const basicAuthSchema = z.object({
  enabled: z.boolean(),
  username: z.string(),
  password: z.string(),
});
export type BasicAuth = z.infer<typeof basicAuthSchema>;
export const defaultBasicAuth: BasicAuth = {
  enabled: false,
  username: "",
  password: "",
};

export const oauthSchema = z.object({
  enabled: z.boolean(),
  consumer_key: z.string(),
  consumer_secret: z.string(),
  access_token: z.string(),
  access_token_secret: z.string(),
});
export type OAuth = z.infer<typeof oauthSchema>;
export const defaultOAuth: OAuth = {
  enabled: false,
  consumer_key: "",
  consumer_secret: "",
  access_token: "",
  access_token_secret: "",
};

export const oauth2Schema = z.object({
  enabled: z.boolean(),
  client_key: z.string(),
  client_secret: z.string(),
  token_url: z.string(),
  scopes: z.array(z.string()),
  endpoint_params: z.record(z.string(), z.string()),
});
export type OAuth2 = z.infer<typeof oauth2Schema>;
export const defaultOAuth2: OAuth2 = {
  enabled: false,
  client_key: "",
  client_secret: "",
  token_url: "",
  scopes: [],
  endpoint_params: {},
};

export const jwtSchema = z.object({
  enabled: z.boolean(),
  private_key_file: z.string(),
  signing_method: z.string(),
  claims: z.record(z.string(), z.string()),
  headers: z.record(z.string(), z.string()),
});
export type Jwt = z.infer<typeof jwtSchema>;
export const defaultJwt: Jwt = {
  enabled: false,
  private_key_file: "",
  signing_method: "",
  claims: {},
  headers: {},
};
