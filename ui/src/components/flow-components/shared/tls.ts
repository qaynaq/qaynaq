import { z } from "zod";

export const tlsSchema = z.object({
  enabled: z.boolean(),
  skip_cert_verify: z.boolean(),
  enable_renegotiation: z.boolean(),
  root_cas: z.string(),
  root_cas_file: z.string(),
  client_certs: z.array(z.unknown()),
});

export type Tls = z.infer<typeof tlsSchema>;

export const defaultTls: Tls = {
  enabled: false,
  skip_cert_verify: false,
  enable_renegotiation: false,
  root_cas: "",
  root_cas_file: "",
  client_certs: [],
};
