FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Build the binary. The openapi.yaml is embedded via //go:embed.
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /usr/local/bin/llm-proxy .

# ── Runtime stage ──────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /usr/local/bin/llm-proxy /usr/local/bin/llm-proxy

# Default data directory for the SQLite database.
RUN mkdir -p /data
ENV DSN=/data/llm-proxy.db

EXPOSE 8080

ENTRYPOINT ["llm-proxy"]
