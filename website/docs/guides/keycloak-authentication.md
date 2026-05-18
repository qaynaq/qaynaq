---
sidebar_position: 4
---

# Keycloak Authentication

This guide walks through setting up Qaynaq with Keycloak as an OAuth2/OIDC identity provider.

## Prerequisites

- Docker installed
- Qaynaq running or ready to start

## Start Keycloak

Run Keycloak with Docker:

```bash
docker run -d --name keycloak \
  -p 8090:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  -e KC_DB=dev-file \
  -e KC_HOSTNAME=localhost \
  -e KC_HOSTNAME_PORT=8090 \
  -e KC_HOSTNAME_STRICT=false \
  -e KC_HTTP_ENABLED=true \
  quay.io/keycloak/keycloak:latest start-dev
```

Keycloak will be available at `http://localhost:8090` with admin credentials `admin` / `admin`.

Create an `qaynaq` realm and client in the Keycloak admin console before proceeding.

## Configure Qaynaq

Set the following environment variables to connect Qaynaq to Keycloak:

```bash
export AUTH_TYPE=oauth2
export AUTH_OAUTH2_CLIENT_ID=qaynaq
export AUTH_OAUTH2_CLIENT_SECRET=qaynaq-secret-change-in-production
export AUTH_OAUTH2_ISSUER_URL=http://localhost:8090/realms/qaynaq
```

Qaynaq fetches `http://localhost:8090/realms/qaynaq/.well-known/openid-configuration` at startup and discovers the authorization, token, and userinfo endpoints from there. The redirect URL is derived from the request host, so no extra config is needed for the standard `localhost:8080/auth/callback` flow.

:::tip
Use the same hostname for browser-side and server-side access to the issuer when possible. If Qaynaq and Keycloak are split across Docker network and the browser (so the issuer host differs between them), give Keycloak a `KC_HOSTNAME` that resolves the same way from both, or run them on the same host network.
:::

## Create Users

1. Open `http://localhost:8090` and log in with `admin` / `admin`.
2. Select the **qaynaq** realm.
3. Go to **Users** and click **Add user**.
4. Fill in a username and email address, then click **Create**.
5. Go to the **Credentials** tab and set a password.

You can now log in to Qaynaq using the credentials you created.

## Restrict Access

To limit which users can access Qaynaq, use email-based filtering:

```bash
# Allow only specific users
export AUTH_OAUTH2_ALLOWED_USERS=alice@company.com,bob@company.com

# Or allow entire domains
export AUTH_OAUTH2_ALLOWED_DOMAINS=company.com
```

See [Authentication](/docs/getting-started/authentication#access-restrictions) for more details. To split signed-in users into Admin and MCP-only roles (for example, mapped from Keycloak group claims), see [Access Control](/docs/guides/access-control).

## Production Considerations

:::warning
The pre-configured realm and client secret are for development only.
:::

For production deployments:

- Create your own Keycloak realm and client with a strong client secret.
- Enable HTTPS on both Keycloak and Qaynaq.
- Set `AUTH_OAUTH2_ISSUER_URL` to your production Keycloak realm URL (e.g. `https://keycloak.example.com/realms/qaynaq`).
- Configure the Keycloak client's **Valid redirect URIs** to match your production Qaynaq host (e.g. `https://qaynaq.example.com/auth/callback`).
