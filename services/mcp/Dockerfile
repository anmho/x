# Build from repo root:
#   docker build -f services/mcp/Dockerfile -t x-mcp .

FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/mcp/go.mod services/mcp/go.sum ./services/mcp/
RUN cd services/mcp && go mod download
COPY services/mcp/ ./services/mcp/
COPY platform.controlplane.json ./platform.controlplane.json
RUN cd services/mcp && CGO_ENABLED=0 go build -o /bin/mcp-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
 && addgroup -S app \
 && adduser -S -G app -h /home/app app \
 && mkdir -p /app /var/lib/x-mcp \
 && chown -R app:app /app /var/lib/x-mcp /home/app
COPY --from=builder /bin/mcp-server /usr/local/bin/mcp-server
COPY --from=builder /src/platform.controlplane.json /app/platform.controlplane.json
WORKDIR /app
VOLUME ["/var/lib/x-mcp"]
ENV HOME=/home/app
ENV MCP_KEYS_FILE=/var/lib/x-mcp/keys.json
ENV MCP_PORT=8765
ENV MCP_ROOT=/app
USER app
EXPOSE 8765
CMD ["mcp-server"]
