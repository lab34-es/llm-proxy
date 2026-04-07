# LLM Proxy

An OpenAI-compatible LLM proxy that routes requests to any OpenAI-compatible backend. Manages upstream provider credentials, issues proxy API keys for clients, enforces per-key rate limits, and tracks token usage. All credentials are stored in a SQLite database.

## Features

- **OpenAI-compatible API** -- drop-in replacement for any OpenAI client library
- **Streaming (SSE)** -- full support for streamed chat completions
- **Provider management** -- register any number of OpenAI-compatible upstream backends
- **Client API keys** -- issue `llmp-` prefixed proxy keys so clients never see upstream credentials
- **Rate limiting** -- configurable per-key requests-per-minute limits (in-memory token bucket)
- **Usage tracking** -- records prompt, completion, and total tokens per request
- **Swagger UI** -- interactive API documentation at `/docs`
- **API-first** -- OpenAPI 3.1 spec is the source of truth, embedded in the binary

## Quick start

```bash
# Build
go build -o llm-proxy .

# Run (ADMIN_TOKEN is required)
ADMIN_TOKEN=my-secret-admin-token ./llm-proxy
```

The server starts on `:8080` by default. Open http://localhost:8080/docs for the Swagger UI.

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `ADDR` | `:8080` | Server listen address |
| `DSN` | `llm-proxy.db` | SQLite database file path |
| `ADMIN_TOKEN` | *(required)* | Bearer token for `/admin/*` endpoints |

## Usage

### 1. Register an upstream provider

```bash
curl -X POST http://localhost:8080/admin/providers \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "openai",
    "base_url": "https://api.openai.com",
    "api_key": "sk-..."
  }'
```

### 2. Create a proxy API key

```bash
curl -X POST http://localhost:8080/admin/keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "provider_id": "<provider-id-from-step-1>",
    "rate_limit_rpm": 60
  }'
```

The response includes a `key` field (e.g. `llmp-abc123...`). This is shown **only once**.

### 3. Use the proxy as an OpenAI endpoint

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer llmp-abc123..." \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

Or with any OpenAI-compatible client library by pointing the base URL to the proxy:

```python
from openai import OpenAI

client = OpenAI(
    api_key="llmp-abc123...",
    base_url="http://localhost:8080/v1",
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### 4. Query usage

```bash
curl "http://localhost:8080/admin/usage?limit=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## API endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/docs` | Public | Swagger UI |
| `GET` | `/openapi.yaml` | Public | Raw OpenAPI spec |
| `POST` | `/admin/providers` | Admin | Register upstream provider |
| `GET` | `/admin/providers` | Admin | List providers |
| `GET` | `/admin/providers/:id` | Admin | Get provider |
| `PUT` | `/admin/providers/:id` | Admin | Update provider |
| `DELETE` | `/admin/providers/:id` | Admin | Delete provider |
| `POST` | `/admin/keys` | Admin | Create proxy API key |
| `GET` | `/admin/keys` | Admin | List proxy API keys |
| `DELETE` | `/admin/keys/:id` | Admin | Revoke API key |
| `GET` | `/admin/usage` | Admin | Query usage records |
| `POST` | `/v1/chat/completions` | Proxy key | Chat completions (streaming + non-streaming) |
| `GET` | `/v1/models` | Proxy key | List upstream models |

## Project structure

```
llm-proxy/
├── openapi.yaml                    # OpenAPI 3.1 spec (embedded at compile time)
├── main.go                         # Entrypoint
└── internal/
    ├── config/config.go            # Environment-based configuration
    ├── db/
    │   ├── db.go                   # SQLite connection (WAL mode)
    │   └── migrations.go           # Auto-migrating schema
    ├── models/models.go            # Domain types
    ├── store/
    │   ├── provider.go             # Provider CRUD
    │   ├── apikey.go               # API key create/lookup/revoke
    │   └── usage.go                # Usage recording and queries
    ├── middleware/
    │   ├── auth.go                 # Admin + proxy key authentication
    │   └── ratelimit.go            # Per-key rate limiting
    ├── handler/
    │   ├── admin.go                # Admin API handlers
    │   ├── proxy.go                # Proxy API handlers
    │   └── docs.go                 # Swagger UI + spec serving
    └── proxy/forwarder.go          # Upstream request forwarding
```

## License

MIT -- see [LICENSE](LICENSE).
