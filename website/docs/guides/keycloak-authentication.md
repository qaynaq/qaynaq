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
export AUTH_OAUTH2_AUTHORIZATION_URL=http://localhost:8090/realms/qaynaq/protocol/openid-connect/auth
export AUTH_OAUTH2_TOKEN_URL=http://keycloak:8080/realms/qaynaq/protocol/openid-connect/token
export AUTH_OAUTH2_REDIRECT_URL=http://localhost:8080/auth/callback
export AUTH_OAUTH2_SCOPES=openid,email,profile
export AUTH_OAUTH2_USER_INFO_URL=http://keycloak:8080/realms/qaynaq/protocol/openid-connect/userinfo
```

:::tip
The authorization URL uses `localhost:8090` because it's accessed from the user's browser. The token and user info URLs use `keycloak:8080` because they're accessed server-side within the Docker network.
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

See [Authentication](/docs/getting-started/authentication#access-restrictions) for more details.

## Production Considerations

:::warning
The pre-configured realm and client secret are for development only.
:::

For production deployments:

- Create your own Keycloak realm and client with a strong client secret.
- Enable HTTPS on both Keycloak and Qaynaq.
- Update the authorization, token, and user info URLs to use your production Keycloak domain.
- Set `AUTH_OAUTH2_REDIRECT_URL` to your production Qaynaq URL (e.g., `https://qaynaq.example.com/auth/callback`).
- Configure the Keycloak client's **Valid redirect URIs** to match your production redirect URL.
