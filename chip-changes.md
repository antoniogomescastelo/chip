# chip MCP Server — Change Summary

**Date:** 2026-05-05
**Author:** Antonio Gomes Castelo
**Branch:** main

---

## Overview

Two features were added to the chip MCP server: HTTPS support and OAuth 2.1 bearer token authentication via Scalekit.

---

## 1. HTTPS Support

### What changed

The HTTP server previously only supported plain HTTP and was hardcoded to bind on `localhost`. It now supports TLS termination and a configurable bind address.

### Files modified

- `cmd/chip/config.go` — new config fields, flags, env vars, validation
- `cmd/chip/main.go` — `runHttpServer` updated to call `ListenAndServeTLS` when cert and key are provided

### New configuration

| Field | CLI flag | Environment variable | Default |
|---|---|---|---|
| Bind address | `--host` | `COLLIBRA_MCP_HTTP_HOST` | `localhost` |
| TLS certificate path | `--tls-cert` | `COLLIBRA_MCP_HTTP_TLS_CERT` | *(disabled)* |
| TLS private key path | `--tls-key` | `COLLIBRA_MCP_HTTP_TLS_KEY` | *(disabled)* |

**Rules:**

- Both `--tls-cert` and `--tls-key` must be provided together; providing only one is a startup error.
- When both are set, the server uses HTTPS (`ListenAndServeTLS`). Otherwise it falls back to plain HTTP as before.
- Set `--host 0.0.0.0` to expose the server on all network interfaces (required for external access).

### Example configuration (`mcp.yaml`)

```yaml
mcp:
  mode: "http"
  http:
    host: "0.0.0.0"
    port: 8443
    tls-cert: "/path/to/cert.pem"
    tls-key:  "/path/to/key.pem"
```

### Testing

```bash
# Generate a self-signed certificate for local testing
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
  -days 30 -nodes -subj "/CN=localhost"

# Run server
./chip --mode http --tls-cert cert.pem --tls-key key.pem --port 8443

# Verify (the -k flag skips verification for self-signed certs)
curl -k https://localhost:8443/mcp
```

---

## 2. OAuth 2.1 Bearer Token Authentication

### What changed

Adds optional OAuth 2.1 authentication using [Scalekit](https://scalekit.com) as the authorization server. When enabled, all HTTP requests to the MCP server must carry a valid JWT bearer token. The server also exposes an OAuth discovery endpoint required by MCP clients.

Authentication is only supported with HTTP transport modes (`http`, `http-sse`, `http-streamable`). Enabling it with `stdio` is a startup error.

### Files modified / added

| File | Change |
|---|---|
| `cmd/chip/auth.go` | **New file.** Scalekit middleware and discovery endpoint handler. |
| `cmd/chip/config.go` | New `AuthConfig` struct, flags, env vars, and startup validation. |
| `cmd/chip/main.go` | Wires auth middleware and discovery route into the HTTP server. |
| `go.mod` / `go.sum` | Added `github.com/scalekit-inc/scalekit-sdk-go/v2 v2.6.0`. |

### How it works

```
MCP client (e.g. Snowflake)
  │
  │  1. POST /mcp  →  401 Unauthorized
  │     WWW-Authenticate: Bearer realm="OAuth",
  │       resource_metadata="https://<server>/.well-known/oauth-protected-resource"
  ▼
chip (resource server)
  │
  │  2. Client fetches /.well-known/oauth-protected-resource
  │     → returns authorization server URLs
  ▼
Scalekit (authorization server)
  │  3. Client completes OAuth 2.1 authorization code + PKCE flow
  │  4. Client receives access token
  ▼
chip (resource server)
  │  5. Client retries POST /mcp with Bearer <token>
  │  6. chip validates token via Scalekit SDK (audience = resource-url)
  │  7. Request proceeds to Collibra API
```

### New configuration

| Field | CLI flag | Environment variable |
|---|---|---|
| Enable auth | `--auth-enabled` | `COLLIBRA_MCP_AUTH_ENABLED` |
| Scalekit environment URL | `--auth-environment-url` | `COLLIBRA_MCP_AUTH_ENVIRONMENT_URL` |
| Scalekit client ID | `--auth-client-id` | `COLLIBRA_MCP_AUTH_CLIENT_ID` |
| Scalekit client secret | `--auth-client-secret` | `COLLIBRA_MCP_AUTH_CLIENT_SECRET` |
| This server's public URL | `--auth-resource-url` | `COLLIBRA_MCP_AUTH_RESOURCE_URL` |
| Authorization server URLs | `--auth-authorization-servers` | `COLLIBRA_MCP_AUTH_AUTHORIZATION_SERVERS` |

All auth fields except `--auth-authorization-servers` are required when `--auth-enabled` is set.

### Example configuration (`mcp.yaml`)

```yaml
mcp:
  mode: "http"
  auth:
    enabled: true
    environment-url: "https://your-env.scalekit.com"
    client-id: "your-client-id"
    client-secret: "your-client-secret"
    resource-url: "https://mcp.your-domain.com"
    authorization-servers:
      - "https://your-env.scalekit.com/resources/res_xxx"
```

### Scalekit dashboard setup

Two separate registrations are required:

1. **chip as a resource server** — creates the `client-id` and `client-secret` used in `mcp.yaml` above.
2. **Each OAuth client separately** (e.g. Snowflake) — creates a distinct `client_id` that the client uses in its own OAuth authorization request.

> **Important:** The OAuth client's `client_id` (e.g. Snowflake's) must not be confused with chip's resource server `client-id`. Using chip's server credential as the `client_id` in an `/oauth/authorize` request will result in an `invalid_client_metadata_url` error.

Required dashboard settings on the chip MCP server entry:

- **Dynamic Client Registration (DCR):** enabled
- **Client ID Metadata Document (CIMD):** enabled

### New endpoint

| Endpoint | Auth required | Description |
|---|---|---|
| `GET /.well-known/oauth-protected-resource` | No | OAuth discovery metadata for MCP clients |

### Testing

```bash
# 1. Verify discovery endpoint is public and returns correct JSON
curl -s https://<your-server>/.well-known/oauth-protected-resource | jq .

# 2. Verify unauthenticated requests are rejected with correct headers
curl -i -X POST https://<your-server>/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
# Expected: HTTP 401 + WWW-Authenticate header

# 3. Get a token via client credentials and test authenticated access
TOKEN=$(curl -s -X POST https://<scalekit-env>/oauth/token \
  -d "grant_type=client_credentials&client_id=<id>&client_secret=<secret>&audience=<resource-url>" \
  | jq -r .access_token)

curl -i -X POST https://<your-server>/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
# Expected: HTTP 200 + tools list
```

---

## New dependency

| Package | Version | Purpose |
|---|---|---|
| `github.com/scalekit-inc/scalekit-sdk-go/v2` | v2.6.0 | JWT token validation via Scalekit |
