---
sidebar_position: 4
---

# Google OAuth Setup

This guide walks you through setting up a Google OAuth connection in Qaynaq. Once configured, the connection works with all Google components (Calendar, Drive, Sheets, and more) and never expires.

## Why OAuth Instead of Service Accounts?

| | OAuth Connection | Service Account |
|---|---|---|
| **Best for** | Personal Google accounts | Google Workspace organizations |
| **Storage quota** | Uses your account's quota | Has its own limited 15GB quota |
| **Setup** | One-time browser authorization | JSON key file + sharing permissions |
| **File ownership** | Files owned by you | Files owned by the service account |
| **Access scope** | All your files and data | Only explicitly shared resources |

## Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown at the top and select **New Project**
3. Enter a name (e.g. "Qaynaq") and click **Create**
4. Make sure the new project is selected in the dropdown

## Step 2: Enable APIs and Register Scopes

For each Google service you want to use, you need to do **two things**: enable the API and register its scope. Both are required - if you miss either one, authorization will fail.

### 2a. Enable the APIs

Go to **APIs & Services** > **Library** and search for and enable each API you need:

| API Name in Library | Google Service | Scope to Register |
|---|---|---|
| Google Calendar API | Calendar | `https://www.googleapis.com/auth/calendar` |
| People API | Contacts | `https://www.googleapis.com/auth/contacts.readonly` |
| Google Docs API | Docs | `https://www.googleapis.com/auth/documents` |
| Google Drive API | Drive | `https://www.googleapis.com/auth/drive` |
| Gmail API | Gmail (read only) | `https://www.googleapis.com/auth/gmail.readonly` |
| Gmail API | Gmail (full access) | `https://www.googleapis.com/auth/gmail.modify` |
| Google Slides API | Slides | `https://www.googleapis.com/auth/presentations` |
| Google Sheets API | Sheets | `https://www.googleapis.com/auth/spreadsheets` |

You do not need to enable all APIs. Only enable the ones for services you plan to use. You can always come back and enable more later.

:::warning Important
For Gmail, both the read-only and full access scopes use the same **Gmail API**. You only need to enable it once, but you must register the specific scope you want in the next step.
:::

### 2b. Register Scopes in the OAuth Consent Screen

Go to **APIs & Services** > **OAuth consent screen**:

1. If you haven't set one up yet:
   - Select **External** user type and click **Create**
   - Fill in **App name** (e.g. "Qaynaq"), **User support email**, and **Developer contact email**
   - Click **Save and Continue**

2. On the **Scopes** step, click **Add or Remove Scopes**

3. For each API you enabled above, find and check the matching scope. You can search by pasting the scope URL:
   - `https://www.googleapis.com/auth/calendar`
   - `https://www.googleapis.com/auth/contacts.readonly`
   - `https://www.googleapis.com/auth/documents`
   - `https://www.googleapis.com/auth/drive`
   - `https://www.googleapis.com/auth/gmail.readonly` (or `gmail.modify` for full access)
   - `https://www.googleapis.com/auth/presentations`
   - `https://www.googleapis.com/auth/spreadsheets`

4. Click **Update** and then **Save and Continue**

:::danger Common Mistake
Every scope you select when creating a connection in Qaynaq **must also be registered here**. If you select a scope in Qaynaq but didn't register it in the consent screen, Google will either return a 500 error or silently skip the scope during authorization.

**Rule of thumb:** If you enable an API in Step 2a, register its scope here in Step 2b.
:::

### 2c. Add Test Users

1. On the **Test users** step, click **Add Users**
2. Enter the Google email address you will use to authorize
3. Click **Add** and then **Save and Continue**

:::info
While the app is in "Testing" mode, only test users you add here can authorize. If you see an "Access denied" error, check that your email is listed here. This is fine for personal use. For production deployments, you can publish the app through Google's verification process.
:::

## Step 3: Create OAuth Client Credentials

1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Select **Web application** as the application type
4. Give it a name (e.g. "Qaynaq")
5. Under **Authorized redirect URIs**, click **Add URI** and enter:
   - For local development: `http://localhost:8080/connections/oauth/callback`
   - For production: `https://your-domain/connections/oauth/callback`
   - If you use the Vite dev server: also add `http://localhost:5173/connections/oauth/callback`
6. Click **Create**
7. Copy the **Client ID** and **Client Secret**

:::warning
The redirect URI must exactly match your Qaynaq instance URL including the port number. If it doesn't match, Google will show a "Redirect URI mismatch" error.
:::

## Step 4: Create the Connection in Qaynaq

1. In Qaynaq, go to **Connections** in the sidebar
2. Click **New Connection**
3. Enter a **Connection Name** (e.g. `my_google`)
4. Select the **Provider** (e.g. Google)
5. Paste the **Client ID** and **Client Secret** from the previous step
6. **Select the scopes** you want - all scopes are selected by default. Deselect any you don't need. Only select scopes for APIs you enabled and registered in Step 2.
7. Click **Authorize**
8. A popup opens with Google's consent screen - sign in and grant access
9. The popup closes automatically and the connection appears in the list

## Using the Connection

In any Google component (Calendar, Sheets, Drive), select your connection from the **OAuth Connection** dropdown. The component will use your Google account for all API calls.

For MCP Tool Packs, select the connection in the shared configuration when deploying the pack.

## Re-authorizing

You may need to re-authorize when:
- You want to add or remove scopes (e.g. you enabled a new API)
- Your token becomes invalid or expired
- You changed your Google password

To re-authorize:

1. Go to the **Connections** page
2. Click the refresh icon on the connection
3. The **Client ID** is pre-filled and the **Client Secret** can be left empty (the existing secret is reused). Only enter a new secret if you regenerated it in GCP.
4. Adjust scopes if needed - your current scopes are pre-selected
5. Click **Re-authorize** and grant access in the popup

:::tip
When adding new scopes, remember to first enable the API and register the scope in GCP (Step 2), then re-authorize the connection in Qaynaq.
:::

## Troubleshooting

**Google returns a 500 error on the consent screen**
- This almost always means a scope is requested but its API is not enabled. Go to **APIs & Services** > **Library** and verify every scope you selected has its API enabled.

**A scope doesn't appear in Google's consent popup**
- The scope must be registered in the OAuth consent screen configuration (Step 2b). Go to **APIs & Services** > **OAuth consent screen** > **Edit App** > **Scopes** and verify it's listed there.

**"Access denied" or 403 error**
- Your email must be added as a test user in the OAuth consent screen configuration (Step 2c).

**"Redirect URI mismatch" error**
- The redirect URI in GCP must exactly match your Qaynaq URL. Check for: trailing slashes, `http` vs `https`, port numbers. If using Vite dev server (port 5173), add that as a separate redirect URI.

**Token stops working after 7 days**
- This can happen when the OAuth consent screen is in "Testing" mode with sensitive scopes. Re-authorize the connection to get fresh tokens.

**"Scope not granted" or missing permissions at runtime**
- You selected a scope in Qaynaq but didn't register it in GCP's consent screen. Register the scope (Step 2b) and re-authorize.
